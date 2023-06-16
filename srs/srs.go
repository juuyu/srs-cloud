// Package srs
// @title
// @description
// @author njy
// @since 2023/5/29 15:25
package srs

import "fmt"

// Stream is a stream in SRS.
type Stream struct {
	Vhost  string `json:"vhost"`
	App    string `json:"app"`
	Stream string `json:"stream"`
	Param  string `json:"param"`
	Server string `json:"server_id"`
	Client string `json:"client_id"`
	Update string `json:"update"`
}

func (v *Stream) String() string {
	return fmt.Sprintf("vhost=%v, app=%v, stream=%v, param=%v, server=%v, client=%v, update=%v",
		v.Vhost, v.App, v.Stream, v.Param, v.Server, v.Client, v.Update,
	)
}

func (v *Stream) StreamURL() string {
	streamURL := fmt.Sprintf("%v/%v/%v", v.Vhost, v.App, v.Stream)
	if v.Vhost == "__defaultVhost__" {
		streamURL = fmt.Sprintf("%v/%v", v.App, v.Stream)
	}
	return streamURL
}
