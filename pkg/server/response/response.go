package response

import (
	"github.com/richard-on/task-service/internal/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ListResponse struct {
	Tasks []model.Task
}

type Info struct {
	Message string `json:"message,omitempty"`
}

type AddResponse struct {
	ID           primitive.ObjectID `json:"id,omitempty"`
	Initiator    string             `json:"initiator"`
	Name         string             `json:"name"`
	Description  string             `json:"description,omitempty"`
	Coordinators []string           `json:"coordinators"`
	Status       model.Status       `json:"status"`
}

type Error struct {
	Error string `json:"error"`
}
