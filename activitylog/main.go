package activitylog

import (
	"context"
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
	activityLogCollection := database.OpenCollection(database.Client, "activity_logs")
	initMatchStage := bson.D{{Key: "$match", Value: bson.M{"post_id": postID}}}
	lookupStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "users"},
		{Key: "localField", Value: "actor"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "user_info"},
	}}}
	unwind := bson.D{{Key: "$unwind", Value: "$user_info"}}
	projectStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "action", Value: 1},
		{Key: "post_id", Value: 1},
		{Key: "created_at", Value: 1},
		{Key: "user_info.name", Value: 1},
	}}}
	sortedStage := bson.D{{Key: "$sort", Value: bson.M{"created_at": -1}}}
	facetStage := bson.D{{Key: "$facet", Value: bson.M{"metadata": bson.A{bson.M{"$count": "total"}, bson.M{"$addFields": bson.M{"page": page}}}, "logs": bson.A{bson.M{"$skip": (page - 1) * limit}, bson.M{"$limit": limit}}}}}
	opts := options.Aggregate().SetAllowDiskUse(true)

	showLoadedCursor, err := activityLogCollection.Aggregate(ctx, mongo.Pipeline{initMatchStage, lookupStage, unwind, projectStage, sortedStage, facetStage}, opts)
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
	resp.Status.Message = fmt.Sprintf("list activity logs successfully")
	resp.Data = results[0]
	c.JSON(http.StatusOK, resp)
}

// Create function
func Create(postID, userID primitive.ObjectID, message string) error {
	activityLogCollection := database.OpenCollection(database.Client, "activity_logs")

	activityLog := ActivityLog{}
	activityLog.ID = primitive.NewObjectID()
	activityLog.PostID = postID
	activityLog.Actor = userID
	activityLog.Action = message
	activityLog.CreatedAt = time.Now()

	_, err := activityLogCollection.InsertOne(context.TODO(), activityLog)
	if err != nil {
		return err
	}

	return nil
}
