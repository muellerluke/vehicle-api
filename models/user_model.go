package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID               primitive.ObjectID   `bson:"_id,omitempty"`
	FirstName        string               `json:"first_name,omitempty" validate:"required"`
	LastName         string               `json:"last_name,omitempty" validate:"required"`
	Email            string               `json:"email,omitempty" validate:"required"`
	Password         string               `json:"password,omitempty" validate:"required"`
	IsVerified       bool                 `json:"is_verified,omitempty"`
	EmailToken       string               `json:"email_token,omitempty"`
	EmailTokenExpiry int64                `json:"email_token_expiry,omitempty"`
	Organizations    []primitive.ObjectID `json:"organizations,omitempty"`
	ResetToken       string               `json:"reset_token,omitempty"`
	ResetTokenExpiry int64                `json:"reset_token_expiry,omitempty"`
}
