package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"job-interview-appointment-api/middleware"
	"job-interview-appointment-api/user"

	"github.com/gin-gonic/gin"
)

// Resource TODO
type Resource struct {
}

func healthcheck(c *gin.Context) {
	if ok, _ := strconv.ParseBool(os.Getenv("DEBUG")); ok {
		fmt.Println("debugging")
	}
	c.JSON(200, "The service is running fine :)")
}

func login(c *gin.Context) {
	var loginParams struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var ok bool
	var userID string

	if ok, userID = user.CheckUserCredentials(loginParams.Username, loginParams.Password); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "The credentials are invalid"})
		return
	}

	tokenExpired, _ := strconv.Atoi(os.Getenv("TOKEN_EXPIRATION_IN_SECONDS"))
	tokenString, err := middleware.GenerateJWT("./private.key", userID, tokenExpired)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
