package models

import (
	"os"
	"tt-board-api/config"

	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TaskInput struct {
	Title       string `json:"title",omitempty`
	TypeId      string `json:"type_id",omitempty`
	Description string `json:"description,omitempty"`
	DueDate     string `json:"due_date"`
}

type Task struct {
	ObjectID    primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title"`
	Type        TaskType           `json:"type" bson:"type"`
	TypeId      primitive.ObjectID `json:"type_id" bson:"type_id"`
	Description string             `json:"description" bson:"description"`
	DueDate     *time.Time         `json:"due_date,omitempty" bson:"due_date,omitempty"`
	BoardId     primitive.ObjectID `json:"board_id,omitempty" bson:"board_id,omitempty"`
	Status      TaskBoard          `json:"status" bson:"status"`
	CreatedAt   *time.Time         `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt   *time.Time         `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

func TasksCollection() *mongo.Collection {
	client := config.MongoClient
	return client.Database(os.Getenv("MONGODB_NAME")).Collection("tasks")
}
