package comment

import (
	"context"
	"encoding/json"
	"fmt"
	"job-interview-appointment-api/common"
	"job-interview-appointment-api/database"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func list(c *gin.Context) {
	var page, limit int

	pageInput := c.Query("page")
	if pageInput == "" {
		page = 1
	} else {
		page, _ = strconv.Atoi(pageInput)
	}

	limitInput := c.Query("limit")
	if limitInput == "" {
		limit = 10
	} else {
		limit, _ = strconv.Atoi(limitInput)
	}

	resp := common.ResponseData{}
	errorList := make(map[string][]string)

	if c.Query("post_id") == "" {
		errorList["post_id"] = []string{"this field is required"}
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("validate error")
		resp.Errors = errorList
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := context.Background()
	postID, _ := primitive.ObjectIDFromHex(c.Query("post_id"))
	commentsCollection := database.OpenCollection(database.Client, "comments")
	initMatchStage := bson.D{{Key: "$match", Value: bson.M{"post_id": postID}}}
	lookupStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "users"},
		{Key: "localField", Value: "created_by"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "user_info"},
	}}}
	unwind := bson.D{{Key: "$unwind", Value: "$user_info"}}
	projectStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "message", Value: 1},
		{Key: "post_id", Value: 1},
		{Key: "created_at", Value: 1},
		{Key: "updated_at", Value: 1},
		{Key: "user_info.name", Value: 1},
	}}}
	sortedStage := bson.D{{Key: "$sort", Value: bson.M{"created_at": -1}}}
	facetStage := bson.D{{Key: "$facet", Value: bson.M{"metadata": bson.A{bson.M{"$count": "total"}, bson.M{"$addFields": bson.M{"page": page}}}, "comments": bson.A{bson.M{"$skip": (page - 1) * limit}, bson.M{"$limit": limit}}}}}
	opts := options.Aggregate().SetAllowDiskUse(true)

	showLoadedCursor, err := commentsCollection.Aggregate(ctx, mongo.Pipeline{initMatchStage, lookupStage, unwind, projectStage, sortedStage, facetStage}, opts)
	if err != nil {
		resp.Status.Code = "FAILED"
		resp.Status.Message = "something went wrong when aggregate the data"
		resp.Errors = err
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	var results []bson.M
	if err = showLoadedCursor.All(ctx, &results); err != nil {
		resp.Status.Code = "FAILED"
		resp.Status.Message = "something went wrong when show the data"
		resp.Errors = err
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp.Status.Code = "SUCCESS"
	resp.Status.Message = fmt.Sprintf("list comments successfully")
	resp.Data = results[0]
	c.JSON(http.StatusOK, resp)
}

func create(c *gin.Context) {
	reqInp, _ := c.GetRawData()
	resp := common.ResponseData{}

	// RequestInput struct
	type RequestInput struct {
		PostID  string `json:"post_id" validate:"required"`
		Message string `json:"message" validate:"required"`
	}

	commentInput := RequestInput{}
	errParseBody := json.Unmarshal(reqInp, &commentInput)
	if errParseBody != nil {
		fmt.Printf("cannot parse body request: %s\n", errParseBody.Error())
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("cannot parse body request: %s", errParseBody.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// validate input data
	if ok, errors := common.ValidateInputs(commentInput); !ok {
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("validate error")
		resp.Errors = errors
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// check post exists before insert comment
	postID, _ := primitive.ObjectIDFromHex(commentInput.PostID)

	ctx := context.Background()
	postsCollection := database.OpenCollection(database.Client, "posts")
	count, err := postsCollection.CountDocuments(ctx, bson.M{"_id": postID})
	if err != nil {
		// fmt.Printf("cannot get data because %s\n", err.Error())
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("cannot get data because %s", err.Error())
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	if count == 0 {
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("cannot create comment because post does not exists")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	commentsCollection := database.OpenCollection(database.Client, "comments")
	comment := Comment{}
	comment.ID = primitive.NewObjectID()
	comment.PostID = postID
	comment.Message = commentInput.Message
	currentUserID, _ := primitive.ObjectIDFromHex(c.MustGet("currUserID").(string))
	comment.CreatedBy = currentUserID
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	result, err := commentsCollection.InsertOne(context.TODO(), comment)
	if err != nil {
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("cannot create comment")
		resp.Errors = err
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp.Status.Code = "SUCCESS"
	resp.Status.Message = fmt.Sprintf("the comment has been created successfully")
	resp.Data = map[string]interface{}{
		"comment_id": result.InsertedID,
	}
	c.JSON(http.StatusOK, resp)
}

func get(c *gin.Context) {

}

func update(c *gin.Context) {
	reqInp, _ := c.GetRawData()
	resp := common.ResponseData{}
	commentID := c.MustGet("comment_id").(primitive.ObjectID)
	type RequestInput struct {
		Message string `json:"message" validate:"required"`
	}

	commentUpdateValue := RequestInput{}
	errParseBody := json.Unmarshal(reqInp, &commentUpdateValue)
	if errParseBody != nil {
		fmt.Printf("cannot parse body request: %s\n", errParseBody.Error())
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("cannot parse body request: %s", errParseBody.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// validate input data
	if ok, errors := common.ValidateInputs(commentUpdateValue); !ok {
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("validate error")
		resp.Errors = errors
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	commentsCollection := database.OpenCollection(database.Client, "comments")

	updateValue := map[string]interface{}{}
	updateValue["message"] = commentUpdateValue.Message
	updateValue["updated_at"] = time.Now()

	if _, err := commentsCollection.UpdateOne(context.Background(), bson.M{"_id": commentID}, bson.M{"$set": updateValue}); err != nil {
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("something went wrong when update data")
		resp.Errors = err
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp.Status.Code = "SUCCESS"
	resp.Status.Message = fmt.Sprintf("the comment has been updated successfully")
	c.JSON(http.StatusOK, resp)
}

func delete(c *gin.Context) {
	resp := common.ResponseData{}
	commentsCollection := database.OpenCollection(database.Client, "comments")
	commentID := c.MustGet("comment_id").(primitive.ObjectID)

	if _, err := commentsCollection.DeleteOne(context.Background(), bson.M{"_id": commentID}); err != nil {
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("something went wrong when delete data")
		resp.Errors = err
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp.Status.Code = "SUCCESS"
	resp.Status.Message = fmt.Sprintf("the comment has been deleted successfully")
	c.JSON(http.StatusOK, resp)
}
