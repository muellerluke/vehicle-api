package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Organization struct {
	ID                     primitive.ObjectID   `bson:"_id,omitempty"`
	Name                   string               `json:"name,omitempty" validate:"required"`
	Owner                  primitive.ObjectID   `json:"owner,omitempty" validate:"required"`
	IsActive               bool                 `json:"is_active,omitempty"`
	IsVerified             bool                 `json:"is_verified,omitempty"`
	StripeCustomerID       string               `json:"stripe_customer_id,omitempty"`
	Keys                   []primitive.ObjectID `json:"keys,omitempty"`
	AutofillSubscriptionID string               `json:"autofill_subscription_id,omitempty"`
}
