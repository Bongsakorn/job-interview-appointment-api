package middleware

import (
	"context"
	"errors"
	"fmt"
	"job-interview-appointment-api/common"
	"job-interview-appointment-api/database"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Authentication
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		claims, err := ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Add the username to the context so the next handlers can use it
		c.Set("currUserID", claims.UID)
		c.Next()
	}
}

// GenerateJWT creates a signed JSON Web Token using a Google API Service Account.
func GenerateJWT(saKeyfile string, userID string, expiryLength int) (string, error) {
	// Extract the RSA private key from the service account keyfile.
	sa, err := os.ReadFile(saKeyfile)
	if err != nil {
		return "", fmt.Errorf("Could not read service account file: %v", err)
	}

	// Create the Claims
	claims := MyCustomClaims{
		userID,
		jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiryLength) * time.Second)),
			Issuer:    "PongsKorner",
			Subject:   "For access this service",
			Audience:  []string{"AnyoneWhoMayKnow"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(sa)
}

// ValidateToken function
func ValidateToken(tokenString string) (*MyCustomClaims, error) {
	claims := &MyCustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		sa, err := os.ReadFile("./private.key")
		if err != nil {
			return "", fmt.Errorf("Could not read service account file: %v", err)
		}
		return sa, nil
	})

	switch {
	case token.Valid:
		if claims, ok := token.Claims.(*MyCustomClaims); ok {
			return claims, nil
		}
	case errors.Is(err, jwt.ErrTokenMalformed):
		return claims, fmt.Errorf("That's not even a token")
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		// Invalid signature
		return claims, fmt.Errorf("Invalid Signature")
	case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
		// Token is either expired or not active yet
		return claims, fmt.Errorf("Timing is everything")
	default:
		return claims, fmt.Errorf("Cloudn't handle this token: %s", err.Error())
	}

	return claims, fmt.Errorf("something went wrong when validate token")
}

// OnlyOwner function using for validate only owner can interact with the comment
func OnlyOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		resp := common.ResponseData{}
		errorList := make(map[string][]string)

		if c.Param("comment_id") == "" || c.Param("comment_id") == ":comment_id" {
			errorList["comment_id"] = []string{"this field is required"}
			resp.Status.Code = "FAILED"
			resp.Status.Message = fmt.Sprintf("validate error")
			resp.Errors = errorList
			c.AbortWithStatusJSON(http.StatusBadRequest, resp)
			return
		}

		commentID, _ := primitive.ObjectIDFromHex(c.Param("comment_id"))

		commentsCollection := database.OpenCollection(database.Client, "comments")
		var currentComment bson.M
		if err := commentsCollection.FindOne(context.Background(), bson.M{"_id": commentID}).Decode(&currentComment); err != nil {
			resp.Status.Code = "FAILED"
			resp.Status.Message = fmt.Sprintf("something went wrong when find data")
			resp.Errors = err
			c.AbortWithStatusJSON(http.StatusInternalServerError, resp)
			return
		}

		// check owner
		currentUserID, _ := primitive.ObjectIDFromHex(c.MustGet("currUserID").(string))
		if currentComment["created_by"].(primitive.ObjectID) != currentUserID {
			resp.Status.Code = "FAILED"
			resp.Status.Message = fmt.Sprintf("you are not allowed to do any action to this comment")
			c.AbortWithStatusJSON(http.StatusUnauthorized, resp)
			return
		}

		c.Set("comment_id", commentID)
		c.Next()
	}
}
