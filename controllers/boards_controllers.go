package controllers

import (
	"context"
	"fmt"
	"net/http"

	"time"

	"tt-board-api/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetBoards(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	matchStage := bson.D{
		{"$match", bson.D{}},
	}

	sortStage := bson.D{
		{"$sort", bson.D{
			{"order", 1},
		}},
	}
	pipeline := mongo.Pipeline{
		matchStage,
		sortStage,
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

func GetFirstBoard(ctx context.Context) (*models.TaskBoard, error) {

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	matchStage := bson.D{
		{"$match", bson.D{}},
	}

	pipeline := mongo.Pipeline{
		matchStage,
	}

	cursor, err := models.TaskBoardsCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch boards: %w", err)
	}
	defer cursor.Close(ctx)

	var taskBoards []models.TaskBoard
	if err := cursor.All(ctx, &taskBoards); err != nil {
		return nil, fmt.Errorf("failed to parse boards: %w", err)
	}

	if len(taskBoards) == 0 {
		return nil, fmt.Errorf("no boards found")
	}

	return &taskBoards[0], nil
}

func GetBoardTasks(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	matchStage := bson.D{
		{"$match", bson.D{}},
	}

	lookupTaskStage := bson.D{
		{"$lookup", bson.D{
			{"from", "tasks"},
			{"let", bson.D{{"board_id", "$_id"}}},
			{"pipeline", bson.A{
				bson.D{{"$match", bson.D{
					{"$expr", bson.D{{"$eq", bson.A{"$board_id", "$$board_id"}}}},
				}}},
				bson.D{{"$lookup", bson.D{
					{"from", "task_types"},
					{"localField", "type_id"},
					{"foreignField", "_id"},
					{"as", "type"},
				}}},
				bson.D{{"$unwind", bson.D{
					{"path", "$type"},
					{"preserveNullAndEmptyArrays", true},
				}}},
			}},
			{"as", "cards"},
		}},
	}

	addCountStage := bson.D{
		{"$addFields", bson.D{
			{"count", bson.D{
				{"$size", "$cards"},
			}},
		}},
	}

	sortStage := bson.D{
		{"$sort", bson.D{
			{"order", 1},
		}},
	}
	pipeline := mongo.Pipeline{
		matchStage,
		lookupTaskStage,
		addCountStage,
		sortStage,
	}

	cursor, err := models.TaskBoardsCollection().Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch radio list"})
		return
	}

	var boardTasks []models.BoardTasks

	if err := cursor.All(ctx, &boardTasks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse radio list"})
		return
	}

	defer cursor.Close(ctx)

	c.JSON(http.StatusOK, gin.H{"data": boardTasks})
}
