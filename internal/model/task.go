package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	FatalError = 0

	NotStarted = iota
	InProgress
	Approved
	Declined
)

type Status uint8

// Task represents a coordination service task.
type Task struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name         string             `json:"name" bson:"name"`
	Description  string             `json:"description" bson:"description"`
	Initiator    string             `json:"initiator" bson:"initiator"`
	Coordinators []string           `json:"coordinators" bson:"coordinators"`
	Next         int                `json:"next" bson:"next"`
	Status       Status             `json:"status" bson:"status"`
}
