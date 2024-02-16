package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"

	"job-interview-appointment-api/configuration"
	"job-interview-appointment-api/database"
	"job-interview-appointment-api/middleware"
)

func main() {
	configuration.Load()
	database.Connect()

	resource := Resource{}

	gin.SetMode(os.Getenv("GIN_MODE"))
	r := gin.New()

	limit, _ := strconv.Atoi(os.Getenv("REQUEST_RATE"))
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: uint(limit),
	})
	mw := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: middleware.ErrorHandler,
		KeyFunc:      middleware.KeyFunc,
	})

	r.Use(mw)
	r.Use(CORSMiddleware())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	resource.loadRoute(r)
	r.Run(fmt.Sprintf(":%s", os.Getenv("PORT")))
}

// CORSMiddleware todo
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
