// Package router
// @title
// @description
// @author njy
// @since 2023/5/29 15:08
package router

import "github.com/gin-gonic/gin"

func StartRouter() *gin.Engine {
	router := gin.Default()
	// srs回调相关接口
	srs := router.Group("srs")
	{
		srs.POST("/hls")
		srs.POST("/unPublish")
	}
	// 文件上传相关接口
	file := router.Group("ffmpeg")
	{
		file.POST("")
	}
	return router
}
