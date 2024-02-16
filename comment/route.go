package comment

import (
	"job-interview-appointment-api/middleware"

	"github.com/gin-gonic/gin"
)

func Router(router *gin.RouterGroup) {
	router.GET("/list", list)
	router.POST("/create", create)
	router.GET("/:comment_id", get)
	router.PUT("/:comment_id", middleware.OnlyOwner(), update)
	router.DELETE("/:comment_id", middleware.OnlyOwner(), delete)
}
