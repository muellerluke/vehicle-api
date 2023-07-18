package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Make struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty"`
	Name         string               `json:"name,omitempty" validate:"required"`
	FindableName string               `json:"findable_name,omitempty" validate:"required"`
	Year         string               `bson:"year"`
	Models       []primitive.ObjectID `bson:"models"`
}
