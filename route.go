package main

import (
	"job-interview-appointment-api/activitylog"
	"job-interview-appointment-api/comment"
	"job-interview-appointment-api/middleware"
	"job-interview-appointment-api/post"
	"job-interview-appointment-api/user"

	"github.com/gin-gonic/gin"
)

func (resource Resource) loadRoute(g *gin.Engine) {
	g.GET("/healthcheck", healthcheck)
	g.POST("/login", login)

	authorized := g.Group("/")
	authorized.Use(middleware.Authentication())
	{
		v1 := authorized.Group("/v1")
		{
			user.Router(v1.Group("/user"))
			post.Router(v1.Group("/post"))
			comment.Router(v1.Group("/comment"))
			activitylog.Router(v1.Group("/activity_log"))
		}
	}
}
