package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type CreateTask struct {
	Title	string	`json:"title" validate:"required"`
	Completed bool `json:"completed"`
	Metadata map[string]string `json:"metadata"`
}

// Task is the model for the task
type Task struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserId    primitive.ObjectID `bson:"user_id,omitempty"`
	Title     string             `bson:"title,required"`
	Completed bool               `bson:"completed,default:false"`
	Metadata  map[string]string  `bson:"metadata,omitempty"`
	CreatedAt int64              `bson:"created_at"`
	UpdatedAt int64              `bson:"updated_at"`
}