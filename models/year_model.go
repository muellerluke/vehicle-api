package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Year struct {
	ID    primitive.ObjectID   `bson:"_id,omitempty"`
	Name  string               `json:"name,omitempty" validate:"required"`
	Makes []primitive.ObjectID `bson:"makes"`
}
