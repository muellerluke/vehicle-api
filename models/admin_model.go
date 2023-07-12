package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Admin struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	FirstName string             `json:"first_name,omitempty" validate:"required"`
	LastName  string             `json:"last_name,omitempty" validate:"required"`
	Email     string             `json:"email,omitempty" validate:"required"`
	Password  string             `json:"password,omitempty" validate:"required"`
}
