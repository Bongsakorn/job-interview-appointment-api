package comment

import "github.com/gin-gonic/gin"

func Router(router *gin.RouterGroup) {
	router.GET("/list", list)
	router.POST("/create", create)
	router.GET("/:comment_id", get)
	router.PUT("/:comment_id", update)
	router.DELETE("/:comment_id", delete)
}
