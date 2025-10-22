package models

import (
	"os"
	"tt-board-api/config"

	"go.mongodb.org/mongo-driver/mongo"
)

type TaskType struct {
	ObjectID string `json:"_id" bson:"_id"`
	Name     string `json:"name" bson:"name"`
	Color    string `json:"color" bson:"color"`
}

func TaskTypesCollection() *mongo.Collection {
	client := config.MongoClient
	return client.Database(os.Getenv("MONGODB_NAME")).Collection("task_types")
}
