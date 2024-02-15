package middleware

import "github.com/golang-jwt/jwt/v5"

// MyCustomClaims strucxt
type MyCustomClaims struct {
	UID string `json:"uid"`
	jwt.RegisteredClaims
}
