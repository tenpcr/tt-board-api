package models

import (
	"os"
	"tt-board-api/config"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TaskBoard struct {
	ObjectID string `json:"_id" bson:"_id"`
	Name     string `json:"name" bson:"name"`
	Color    string `json:"color" bson:"color"`
}

type BoardTasks struct {
	ObjectID primitive.ObjectID `json:"_id" bson:"_id"`
	Name     string             `json:"name" bson:"name"`
	Color    string             `json:"color" bson:"color"`
	Cards    []Task             `json:"cards" bson:"cards"`
	Count    int64              `json:"count" bson:"count"`
}

func TaskBoardsCollection() *mongo.Collection {
	client := config.MongoClient
	return client.Database(os.Getenv("MONGODB_NAME")).Collection("task_boards")
}
