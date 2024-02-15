package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

// func ValidateToken() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		fmt.Println("@@@@@@@@@ Authorize @@@@@@@@@@@@@@@")
// 		fmt.Printf("%v\n", c.Request.Header)
// 		var val string
// 		if val = c.Request.Header.Get("Payload"); val == "" {
// 			c.AbortWithStatusJSON(401, map[string]string{"Message": "user hasn't logged in yet"})
// 			return
// 		}

// 		var payload map[string]interface{}
// 		err := json.Unmarshal([]byte(val), &payload)
// 		if err != nil {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]string{"Message": fmt.Sprintf("cannot unmarshal payload because %s", err.Error())})
// 			return
// 		}
// 		fmt.Println(payload)

// 		// collection := database.OpenCollection(database.Client, "users")
// 		// userDB := User{}
// 		// ctx := context.Background()

// 		// if err := collection.FindOne(ctx, bson.M{"uid": payload["user_id"]}).Decode(&userDB); err != nil {
// 		// 	fmt.Println(err.Error())
// 		// 	c.AbortWithStatusJSON(403, map[string]string{"Message": "not found user in system"})
// 		// 	return
// 		// }

// 		// fmt.Println(userDB)
// 		c.Set("uid", payload["user_id"])
// 		// c.Set("uid", userDB.UID)
// 		c.Next()
// 	}
// }
