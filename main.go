// Package srs_cloud
// @title
// @description
// @author njy
// @since 2023/5/29 14:34
package main

import (
	"log"
	"srs-cloud/router"
)

func main() {
	// 启动gin
	r := router.StartRouter()
	err := r.Run(":7070")
	if err != nil {
		log.Fatal("server start failed,err:", err)
	}
}
