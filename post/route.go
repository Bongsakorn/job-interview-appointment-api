package post

import "github.com/gin-gonic/gin"

func Router(router *gin.RouterGroup) {
	router.GET("/list", list)
	router.POST("/create", create)
	router.GET("/:post_id", get)
	router.PUT("/:post_id", update)
	router.PATCH("/archive/:post_id", archive)
}
