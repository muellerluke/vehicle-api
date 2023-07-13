package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Call struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	User           primitive.ObjectID `bson:"user,omitempty" validate:"required"`
	Key            primitive.ObjectID `bson:"key,omitempty" validate:"required"`
	RequestURL     string             `json:"request_url,omitempty" validate:"required"`
	ResponseStatus int                `json:"response_status,omitempty"`
	ResponseTime   int                `json:"response_time,omitempty"`
	CreatedAt      int64              `json:"created_at,omitempty" validate:"required"`
}
