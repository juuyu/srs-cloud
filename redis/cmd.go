// Package redis
// @title
// @description
// @author njy
// @since 2023/5/29 15:54
package redis

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"log"
	"srs-cloud/conf"
)

var Rdb *redis.Client

func Exists(ctx context.Context, key string) bool {
	return Rdb.Exists(ctx, key).Val() == 1 // 1存在，0不存在
}
func Get(ctx context.Context, key string) string {
	return Rdb.Get(ctx, key).Val()
}
func LPush(ctx context.Context, key string, value interface{}) {
	Rdb.LPush(ctx, key, marshal(value))
}
func RPush(ctx context.Context, key string, value interface{}) {
	Rdb.RPush(ctx, key, marshal(value))
}
func LRange(ctx context.Context, key string) []string {
	return Rdb.LRange(ctx, key, 0, -1).Val()
}
func Del(ctx context.Context, key string) {
	Rdb.Del(ctx, key)
}

func marshal(value interface{}) string {
	data, err := json.Marshal(value)
	if err != nil {
		log.Fatal("序列化错误 err=", err)
	}
	return string(data)
}
func unMarshal(value string) {
}

func InitRedis() {
	var options redis.Options
	if conf.Cfg.Redis.Secure {
		options = redis.Options{
			Addr:     conf.Cfg.Redis.Addr,
			Password: conf.Cfg.Redis.Pwd,
			DB:       conf.Cfg.Redis.DB,
		}
	} else {
		options = redis.Options{
			Addr: conf.Cfg.Redis.Addr,
			DB:   conf.Cfg.Redis.DB,
		}
	}
	Rdb = redis.NewClient(&options)
	result, err := Rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("redis 启动失败")
	}
	log.Println(result)
}
