package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID                     primitive.ObjectID   `bson:"_id,omitempty"`
	FirstName              string               `json:"first_name,omitempty" validate:"required"`
	LastName               string               `json:"last_name,omitempty" validate:"required"`
	Email                  string               `json:"email,omitempty" validate:"required"`
	Password               string               `json:"password,omitempty" validate:"required"`
	Keys                   []primitive.ObjectID `bson:"keys"`
	CreatedAt              int64                `json:"created_at,omitempty"`
	IsVerified             bool                 `json:"is_verified"`
	EmailToken             string               `json:"email_token,omitempty"`
	EmailTokenExpiry       int64                `json:"email_token_expiry,omitempty"`
	ResetToken             string               `json:"reset_token,omitempty"`
	ResetTokenExpiry       int64                `json:"reset_token_expiry,omitempty"`
	IsActive               bool                 `json:"is_active,omitempty"`
	StripeCustomerID       string               `json:"stripe_customer_id,omitempty"`
	PaymentMethodID        string               `json:"payment_method_id,omitempty"`
	SetupIntentID          string               `json:"setup_intent_id,omitempty"`
	AutofillSubscriptionID string               `json:"autofill_subscription_id,omitempty"`
}
