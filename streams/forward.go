// Package streams
// @title
// @description
// @author njy
// @since 2023/5/29 15:17
package streams

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"os/exec"
	"srs-cloud/errors"
	"srs-cloud/logger"
	"srs-cloud/srs"
	"strings"
	"sync"
	"syscall"
	"time"
)

var forwardWorker *ForwardWorker

type ForwardWorker struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup
	// The tasks we have started to forward streams, key is platform in string, value is *ForwardTask.
	tasks sync.Map
}

// ForwardConfigure is the configuration for forwarding.
type ForwardConfigure struct {
	// The platform name, for example, wx
	Platform string `json:"platform"`
	// The RTMP server url, for example, rtmp://localhost/live
	Server string `json:"server"`
	// The RTMP stream and secret, for example, livestream
	Secret string `json:"secret"`
	// Whether enabled.
	Enabled bool `json:"enabled"`
	// Whether custom platform.
	Customed bool `json:"custom"`
	// The label for this configure.
	Label string `json:"label"`
}

func (v *ForwardConfigure) String() string {
	return fmt.Sprintf("platform=%v, server=%v, secret=%v, enabled=%v, customed=%v, label=%v",
		v.Platform, v.Server, v.Secret, v.Enabled, v.Customed, v.Label,
	)
}

// ForwardTask is a task for FFmpeg to forward stream, with a configuration.
type ForwardTask struct {
	// The ID for task.
	UUID string `json:"uuid"`
	// The platform for task.
	Platform string `json:"platform"`
	// The input url.
	Input string `json:"input"`
	// The input stream URL.
	inputStreamURL string
	// The output url
	Output string `json:"output"`
	// FFmpeg pid.
	PID int32 `json:"pid"`
	// FFmpeg last frame.
	frame string
	// The last update time.
	update string
	// The context for current task.
	cancel context.CancelFunc
	// The configuration for forwarding task.
	config *ForwardConfigure
	// The forward worker.
	forwardWorker *ForwardWorker
	// To protect the fields.
	lock sync.Mutex
}

func (v *ForwardTask) String() string {
	return fmt.Sprintf("uuid=%v, platform=%v, input=%v, output=%v, pid=%v, frame=%vB, config is %v",
		v.UUID, v.Platform, v.Input, v.Output, v.PID, len(v.frame), v.config.String(),
	)
}

func (v *ForwardTask) saveTask(ctx context.Context) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	if b, err := json.Marshal(v); err != nil {
		return errors.Wrapf(err, "marshal %v", v.String())
	} else if err = rdb.HSet(ctx, redis.SRS_FORWARD_TASK, v.UUID, string(b)).Err(); err != nil && err != redis.Nil {
		return errors.Wrapf(err, "hset %v %v %v", SRS_FORWARD_TASK, v.UUID, string(b))
	}

	return nil
}

func (v *ForwardTask) cleanup(ctx context.Context) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.PID <= 0 {
		return nil
	}

	//logger.Wf(ctx, "kill task pid=%v", v.PID)
	err := syscall.Kill(int(v.PID), syscall.SIGKILL)
	if err != nil {
		return err
	}

	v.PID = 0
	v.cancel = nil

	return nil
}

func (v *ForwardTask) updateFrame(frame string) {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.frame = frame
	v.update = time.Now().Format(time.RFC3339)
}

func (v *ForwardTask) queryFrame() (int32, string, string, string) {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.PID, v.inputStreamURL, v.frame, v.update
}

func (v *ForwardTask) Initialize(ctx context.Context, w *ForwardWorker) error {
	v.forwardWorker = w
	logger.Tf(ctx, "forward initialize uuid=%v, platform=%v", v.UUID, v.Platform)

	if err := v.saveTask(ctx); err != nil {
		return errors.Wrapf(err, "save task")
	}

	return nil
}

func (v *ForwardTask) doForward(ctx context.Context, input *srs.Stream) error {
	// Create context for current task.
	parentCtx := ctx
	ctx, v.cancel = context.WithCancel(ctx)

	// Build input URL.
	host := "localhost"
	inputURL := fmt.Sprintf("rtmp://%v/%v/%v", host, input.App, input.Stream)

	// Build output URL.
	outputServer := strings.ReplaceAll(v.config.Server, "localhost", host)
	if !strings.HasSuffix(outputServer, "/") && !strings.HasPrefix(v.config.Secret, "/") {
		outputServer += "/"
	}
	outputURL := fmt.Sprintf("%v%v", outputServer, v.config.Secret)

	// Start FFmpeg process.
	args := []string{
		"-stream_loop", "-1", "-i", inputURL, "-c", "copy", "-f", "flv", outputURL,
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.Wrapf(err, "pipe process")
	}

	if err := cmd.Start(); err != nil {
		return errors.Wrapf(err, "execute ffmpeg %v", strings.Join(args, " "))
	}

	v.PID = int32(cmd.Process.Pid)
	v.Input, v.inputStreamURL, v.Output = inputURL, input.StreamURL(), outputURL
	defer func() {
		// If we got a PID, sleep for a while, to avoid too fast restart.
		if v.PID > 0 {
			select {
			case <-ctx.Done():
			case <-time.After(1 * time.Second):
			}
		}

		// When canceled, we should still write to redis, so we must not use ctx(which is cancelled).
		v.cleanup(parentCtx)
		v.saveTask(parentCtx)
	}()
	log.Println("forward start, platform=", v.Platform, "stream=", input.StreamURL(), "pid=", v.PID)

	if err := v.saveTask(ctx); err != nil {
		return errors.Wrapf(err, "save task %v", v.String())
	}

	buf := make([]byte, 4096)
	for {
		nn, err := stderr.Read(buf)
		if err != nil || nn == 0 {
			break
		}

		line := string(buf[:nn])
		for strings.Contains(line, "= ") {
			line = strings.ReplaceAll(line, "= ", "=")
		}
		v.updateFrame(line)
	}

	err = cmd.Wait()
	logger.Tf(ctx, "forward done, platform=%v, stream=%v, pid=%v, err=%v",
		v.Platform, input.StreamURL(), v.PID, err,
	)

	return nil
}
