package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// User is the model for the user 
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	FirstName string `bson:"first_name,required"`
	LastName  string `bson:"last_name,omitempty"`
	Email    string `bson:"email,required"`
	Mobile   string `bson:"mobile,omitempty"`
	PasswordHash string `bson:"password_hash"`
	Tokens []string `bson:"tokens"`
	CreatedAt int64 `bson:"created_at"`
	UpdatedAt int64 `bson:"updated_at"`
}