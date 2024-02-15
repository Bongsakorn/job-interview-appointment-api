package user

import (
	"context"
	"fmt"
	"job-interview-appointment-api/common"
	"job-interview-appointment-api/database"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// CheckUserCredentials function
func CheckUserCredentials(username, password string) (bool, string) {
	collection := database.OpenCollection(database.Client, "users")
	userDB := DataWithPassword{}
	ctx := context.Background()

	if err := collection.FindOne(ctx, bson.M{"email": username}).Decode(&userDB); err != nil {
		fmt.Println(err.Error())
		// c.AbortWithStatusJSON(403, map[string]string{"Message": "not found user in system"})
		return false, ""
	}

	err := bcrypt.CompareHashAndPassword([]byte(userDB.Password), []byte(password))
	return err == nil, userDB.ID.Hex()
}

func getMe(c *gin.Context) {
	userID, _ := primitive.ObjectIDFromHex(c.MustGet("currUserID").(string))
	collection := database.OpenCollection(database.Client, "users")
	userDB := User{}
	ctx := context.Background()
	resp := common.ResponseData{}

	if err := collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&userDB); err != nil {
		// fmt.Println(err.Error())
		resp.Status.Code = "FAILED"
		resp.Status.Message = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp.Status.Code = "SUCCESS"
	resp.Status.Message = "get user successfully"
	resp.Data = userDB
	c.JSON(http.StatusOK, resp)
}
