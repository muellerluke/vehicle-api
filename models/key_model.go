package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Key struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	User              primitive.ObjectID `bson:"user,omitempty" validate:"required"`
	Key               string             `json:"key,omitempty" validate:"required"`
	Routes            []string           `json:"routes,omitempty" validate:"required"`
	AuthorizedDomains []string           `json:"authorized_domains"`
	IsActive          bool               `json:"is_active,omitempty"`
}

//authorizedDomains are only used for autofill routes. If it is left emty, then all domains are valid
