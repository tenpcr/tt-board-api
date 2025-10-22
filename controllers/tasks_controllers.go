package controllers

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"tt-board-api/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetTasks(c *gin.Context) {
	queryLimitStr := c.Query("limit")
	queryPageStr := c.Query("page")

	page := 0
	limit := 1

	if queryLimitStr != "" {
		parsedLimit, err := strconv.Atoi(queryLimitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if queryPageStr != "" {
		parsedPage, err := strconv.Atoi(queryPageStr)
		if err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	skip := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	matchStage := bson.D{
		{"$match", bson.D{}},
	}

	lookupTypeStage := bson.D{
		{"$lookup", bson.D{
			{"from", "task_types"},
			{"let", bson.D{{"type_id", "$type_id"}}},
			{"pipeline", bson.A{
				bson.D{{"$match", bson.D{
					{"$expr", bson.D{{"$eq", bson.A{"$_id", "$$type_id"}}}},
				}}},
			}},
			{"as", "type"},
		}},
	}

	unwindTypeStage := bson.D{{
		"$unwind", bson.D{
			{"path", "$type"},
			{"preserveNullAndEmptyArrays", true},
		},
	}}

	lookupStatusStage := bson.D{
		{"$lookup", bson.D{
			{"from", "task_boards"},
			{"let", bson.D{{"board_id", "$board_id"}}},
			{"pipeline", bson.A{
				bson.D{{"$match", bson.D{
					{"$expr", bson.D{{"$eq", bson.A{"$_id", "$$board_id"}}}},
				}}},
			}},
			{"as", "status"},
		}},
	}

	unwindStatusStage := bson.D{{
		"$unwind", bson.D{
			{"path", "$status"},
			{"preserveNullAndEmptyArrays", true},
		},
	}}

	pipeline := mongo.Pipeline{
		matchStage,
		lookupTypeStage,
		unwindTypeStage,
		lookupStatusStage,
		unwindStatusStage,
		bson.D{{"$sort", bson.M{"_id": -1}}},
		bson.D{{"$skip", skip}},
		bson.D{{"$limit", limit}},
	}

	cursor, err := models.TasksCollection().Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch radio list"})
		return
	}

	var tasks []models.Task

	if err := cursor.All(ctx, &tasks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse radio list"})
		return
	}

	defer cursor.Close(ctx)

	count := CountTasks(c)

	c.JSON(http.StatusOK, gin.H{"data": tasks, "task_count": count, "page": page, "page_count": int(math.Ceil(float64(count) / float64(limit)))})
}

func GetTaskById(c *gin.Context) {
	limit := 1

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID is required"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	matchStage := bson.D{
		{"$match", bson.M{"_id": objectID}},
	}

	lookupTypeStage := bson.D{
		{"$lookup", bson.D{
			{"from", "task_types"},
			{"let", bson.D{{"type_id", "$type_id"}}},
			{"pipeline", bson.A{
				bson.D{{"$match", bson.D{
					{"$expr", bson.D{{"$eq", bson.A{"$_id", "$$type_id"}}}},
				}}},
			}},
			{"as", "type"},
		}},
	}

	unwindTypeStage := bson.D{{
		"$unwind", bson.D{
			{"path", "$type"},
			{"preserveNullAndEmptyArrays", true},
		},
	}}

	lookupStatusStage := bson.D{
		{"$lookup", bson.D{
			{"from", "task_boards"},
			{"let", bson.D{{"board_id", "$board_id"}}},
			{"pipeline", bson.A{
				bson.D{{"$match", bson.D{
					{"$expr", bson.D{{"$eq", bson.A{"$_id", "$$board_id"}}}},
				}}},
			}},
			{"as", "status"},
		}},
	}

	unwindStatusStage := bson.D{{
		"$unwind", bson.D{
			{"path", "$status"},
			{"preserveNullAndEmptyArrays", true},
		},
	}}

	pipeline := mongo.Pipeline{
		matchStage,
		lookupTypeStage,
		unwindTypeStage,
		lookupStatusStage,
		unwindStatusStage,
		bson.D{{"$limit", limit}},
	}

	cursor, err := models.TasksCollection().Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch radio list"})
		return
	}

	var tasks []models.Task

	if err := cursor.All(ctx, &tasks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse radio list"})
		return
	}

	defer cursor.Close(ctx)

	if len(tasks) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tasks[0]})
}

func CountTasks(c *gin.Context) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := models.TasksCollection().CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to count tasks: %v", err)
	}

	return count
}

func PostAddTask(c *gin.Context) {
	var input models.TaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Println("Bind error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "detail": err.Error()})
		return
	}

	var TypeId primitive.ObjectID
	if input.TypeId != "" {
		var err error
		TypeId, err = primitive.ObjectIDFromHex(input.TypeId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type ID"})
			return
		}
	}

	var dueDate *time.Time
	if input.DueDate != "" {
		parsed, err := time.Parse(time.RFC3339, input.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due_date format"})
			return
		}
		dueDate = &parsed
	}

	board, _ := GetFirstBoard(c)

	task := models.Task{
		Title:       input.Title,
		TypeId:      TypeId,
		DueDate:     dueDate,
		Description: input.Description,
		BoardId: func() primitive.ObjectID {
			objID, err := primitive.ObjectIDFromHex(board.ObjectID)
			if err != nil {
				return primitive.NilObjectID
			}
			return objID
		}(),
		CreatedAt: func() *time.Time { t := time.Now(); return &t }(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := models.TasksCollection().InsertOne(ctx, task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert task", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Task created successfully", "data": task})
}

func GetTaskBoards(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	matchStage := bson.D{
		{"$match", bson.D{}},
	}

	pipeline := mongo.Pipeline{
		matchStage,
	}

	cursor, err := models.TaskBoardsCollection().Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch radio list"})
		return
	}

	var taskBoards []models.TaskBoard

	if err := cursor.All(ctx, &taskBoards); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse radio list"})
		return
	}

	defer cursor.Close(ctx)

	c.JSON(http.StatusOK, gin.H{"data": taskBoards})
}

func GetTaskTypes(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	matchStage := bson.D{
		{"$match", bson.D{}},
	}

	pipeline := mongo.Pipeline{
		matchStage,
	}

	cursor, err := models.TaskTypesCollection().Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch radio list"})
		return
	}

	var taskTypes []models.TaskType

	if err := cursor.All(ctx, &taskTypes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse radio list"})
		return
	}

	defer cursor.Close(ctx)

	c.JSON(http.StatusOK, gin.H{"data": taskTypes})
}

func DeleteTaskById(c *gin.Context) {
	idParam := c.Param("id")

	taskID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": taskID}

	result, err := models.TasksCollection().DeleteOne(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task", "detail": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func UpdateTaskById(c *gin.Context) {
	idParam := c.Param("id")

	taskID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var reqBody struct {
		Title       *string `json:"title,omitempty"`
		TypeId      string  `json:"type_id,omitempty"`
		Description string  `json:"description,omitempty"`
		DueDate     string  `json:"due_date,omitempty"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updateFields := bson.M{}

	if reqBody.Title != nil {
		updateFields["title"] = *reqBody.Title
	}
	if reqBody.Description != "" {
		updateFields["description"] = reqBody.Description
	}

	var TypeId primitive.ObjectID

	if reqBody.TypeId != "" {
		var err error
		TypeId, err = primitive.ObjectIDFromHex(reqBody.TypeId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid progress ID"})
			return
		}
	}

	if reqBody.TypeId != "" {
		updateFields["type_id"] = TypeId
	}

	if reqBody.DueDate != "" {
		updateFields["due_date"] = reqBody.DueDate
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": taskID}
	update := bson.M{"$set": updateFields}

	result, err := models.TasksCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task", "detail": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task updated successfully",
		"updated": updateFields,
	})
}

func UpdateTaskProgressById(c *gin.Context) {
	idParam := c.Param("id")

	taskID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var reqBody struct {
		ProgressId string `json:"progress_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var ProgressId primitive.ObjectID
	if reqBody.ProgressId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid progress ID"})
		return
	} else {
		var err error
		ProgressId, err = primitive.ObjectIDFromHex(reqBody.ProgressId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid progress ID"})
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{"board_id": ProgressId}}
	filter := bson.M{"_id": taskID}

	result, err := models.TasksCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update progress", "detail": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task progress updated", "progress": reqBody.ProgressId})
}
