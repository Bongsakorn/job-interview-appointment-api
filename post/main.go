package post

import (
	"context"
	"encoding/json"
	"fmt"
	"job-interview-appointment-api/activitylog"
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
	resp := common.ResponseData{}

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

	postsCollection := database.OpenCollection(database.Client, "posts")

	initMatchStage := bson.D{{Key: "$match", Value: bson.M{"archived": false}}}
	lookupStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "users"},
		{Key: "localField", Value: "created_by"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "user_info"},
	}}}
	unwind := bson.D{{Key: "$unwind", Value: "$user_info"}}
	projectStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "title", Value: 1},
		{Key: "description", Value: 1},
		{Key: "status", Value: 1},
		{Key: "created_at", Value: 1},
		{Key: "user_info.name", Value: 1},
	}}}
	sortedStage := bson.D{{Key: "$sort", Value: bson.M{"created_at": -1}}}
	facetStage := bson.D{{Key: "$facet", Value: bson.M{"metadata": bson.A{bson.M{"$count": "total"}, bson.M{"$addFields": bson.M{"page": page}}}, "posts": bson.A{bson.M{"$skip": (page - 1) * limit}, bson.M{"$limit": limit}}}}}

	opts := options.Aggregate().SetAllowDiskUse(true)
	showLoadedCursor, err := postsCollection.Aggregate(context.TODO(), mongo.Pipeline{initMatchStage, lookupStage, unwind, projectStage, sortedStage, facetStage}, opts)
	if err != nil {
		resp.Status.Code = "FAILED"
		resp.Status.Message = "something went wrong when aggregate the data"
		resp.Errors = err
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	var results []bson.M
	if err = showLoadedCursor.All(context.TODO(), &results); err != nil {
		resp.Status.Code = "FAILED"
		resp.Status.Message = "something went wrong when show the data"
		resp.Errors = err
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp.Status.Code = "SUCCESS"
	resp.Status.Message = fmt.Sprintf("list posts successfully")
	resp.Data = results[0]
	c.JSON(http.StatusOK, resp)
}

func create(c *gin.Context) {
	reqInp, _ := c.GetRawData()
	resp := common.ResponseData{}
	postInput := RequestInput{}
	errParseBody := json.Unmarshal(reqInp, &postInput)
	if errParseBody != nil {
		fmt.Printf("cannot parse body request: %s\n", errParseBody.Error())
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("cannot parse body request: %s", errParseBody.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// validate input data
	if ok, errors := common.ValidateInputs(postInput); !ok {
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("validate error")
		resp.Errors = errors
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	postsCollection := database.OpenCollection(database.Client, "posts")
	post := Post{}
	post.ID = primitive.NewObjectID()
	post.Title = postInput.Title
	post.Description = postInput.Description
	post.Status = "to_do"
	currentUserID, _ := primitive.ObjectIDFromHex(c.MustGet("currUserID").(string))
	post.CreatedBy = currentUserID
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()

	result, err := postsCollection.InsertOne(context.TODO(), post)
	if err != nil {
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("cannot create post")
		resp.Errors = err
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	errActivityLog := activitylog.Create(result.InsertedID.(primitive.ObjectID), currentUserID, "Post Created")
	if errActivityLog != nil {
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("cannot create activity log")
		resp.Errors = errActivityLog
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp.Status.Code = "SUCCESS"
	resp.Status.Message = fmt.Sprintf("the post has been created successfully")
	resp.Data = map[string]interface{}{
		"post_id": result.InsertedID,
	}
	c.JSON(http.StatusOK, resp)
}

func get(c *gin.Context) {
	resp := common.ResponseData{}
	errorList := make(map[string][]string)

	if c.Param("post_id") == "" || c.Param("post_id") == ":post_id" {
		errorList["post_id"] = []string{"this field is required"}
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("validate error")
		resp.Errors = errorList
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	postID, _ := primitive.ObjectIDFromHex(c.Param("post_id"))

	postsCollection := database.OpenCollection(database.Client, "posts")

	initMatchStage := bson.D{{Key: "$match", Value: bson.M{"_id": postID}}}
	lookupStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "users"},
		{Key: "localField", Value: "created_by"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "user_info"},
	}}}
	unwind := bson.D{{Key: "$unwind", Value: "$user_info"}}
	projectStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "title", Value: 1},
		{Key: "description", Value: 1},
		{Key: "status", Value: 1},
		{Key: "created_at", Value: 1},
		{Key: "user_info.name", Value: 1},
	}}}

	opts := options.Aggregate().SetAllowDiskUse(true)
	showLoadedCursor, err := postsCollection.Aggregate(context.TODO(), mongo.Pipeline{initMatchStage, lookupStage, unwind, projectStage}, opts)
	if err != nil {
		resp.Status.Code = "FAILED"
		resp.Status.Message = "something went wrong when aggregate the data"
		resp.Errors = err
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	var results []bson.M
	if err = showLoadedCursor.All(context.TODO(), &results); err != nil {
		resp.Status.Code = "FAILED"
		resp.Status.Message = "something went wrong when show the data"
		resp.Errors = err
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp.Status.Code = "SUCCESS"
	resp.Status.Message = fmt.Sprintf("get post successfully")
	resp.Data = results
	c.JSON(http.StatusOK, resp)
}

func update(c *gin.Context) {
	reqInp, _ := c.GetRawData()
	resp := common.ResponseData{}
	errorList := make(map[string][]string)

	if c.Param("post_id") == "" || c.Param("post_id") == ":post_id" {
		errorList["post_id"] = []string{"this field is required"}
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("validate error")
		resp.Errors = errorList
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	postID, _ := primitive.ObjectIDFromHex(c.Param("post_id"))

	type RequestInput struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status" validate:"oneof=to_do in_progress done"`
	}

	postUpdateValue := RequestInput{}
	errParseBody := json.Unmarshal(reqInp, &postUpdateValue)
	if errParseBody != nil {
		fmt.Printf("cannot parse body request: %s\n", errParseBody.Error())
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("cannot parse body request: %s", errParseBody.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// validate input data
	if ok, errors := common.ValidateInputs(postUpdateValue); !ok {
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("validate error")
		resp.Errors = errors
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	postsCollection := database.OpenCollection(database.Client, "posts")
	var currentPost bson.M
	if err := postsCollection.FindOne(context.Background(), bson.M{"_id": postID}).Decode(&currentPost); err != nil {
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("something went wrong when find data")
		resp.Errors = err
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	changes := make(bson.M)
	newValues := map[string]interface{}{}
	jsonEnc, _ := json.Marshal(postUpdateValue)
	json.Unmarshal(jsonEnc, &newValues)
	for key, newValue := range newValues {
		if currentPost[key] != newValue {
			changes[key] = bson.M{"old": currentPost[key], "new": newValue}
		}
	}
	newValues["updated_at"] = time.Now()

	if _, err := postsCollection.UpdateOne(context.Background(), bson.M{"_id": postID}, bson.M{"$set": newValues}); err != nil {
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("something went wrong when update data")
		resp.Errors = err
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	if len(changes) > 0 {
		changeMessage, _ := json.Marshal(changes)
		currentUserID, _ := primitive.ObjectIDFromHex(c.MustGet("currUserID").(string))
		errActivityLog := activitylog.Create(postID, currentUserID, string(changeMessage))
		if errActivityLog != nil {
			resp.Status.Code = "FAILED"
			resp.Status.Message = fmt.Sprintf("cannot create activity log")
			resp.Errors = errActivityLog
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
	}

	resp.Status.Code = "SUCCESS"
	resp.Status.Message = fmt.Sprintf("the post has been updated")
	c.JSON(http.StatusOK, resp)
}

func archive(c *gin.Context) {
	resp := common.ResponseData{}
	errorList := make(map[string][]string)

	if c.Param("post_id") == "" || c.Param("post_id") == ":post_id" {
		errorList["post_id"] = []string{"this field is required"}
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("validate error")
		resp.Errors = errorList
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	postID, _ := primitive.ObjectIDFromHex(c.Param("post_id"))

	postsCollection := database.OpenCollection(database.Client, "posts")

	updateValue := map[string]interface{}{}
	updateValue["archived"] = true
	updateValue["updated_at"] = time.Now()

	if _, err := postsCollection.UpdateOne(context.Background(), bson.M{"_id": postID}, bson.M{"$set": updateValue}); err != nil {
		resp.Status.Code = "FAILED"
		resp.Status.Message = fmt.Sprintf("something went wrong when update data")
		resp.Errors = err
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp.Status.Code = "SUCCESS"
	resp.Status.Message = fmt.Sprintf("the post has been archived successfully")
	c.JSON(http.StatusOK, resp)
}
