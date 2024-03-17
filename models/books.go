package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Books struct {
	ID          primitive.ObjectID `bson:"_id"`
	Name        string             `json:"name" validate:"required"`
	Author_name string             `json:"author_name" validate:"required" default:"anonymous"`
	Price       int                `json:"price" validate:"required"`
	Description string             `json:"description"`
	Author_info string             `json:"author_info"`
	Publication string             `json:"publication" validate:"required"`
	Genre       string             `json:"genre"`
	Category    string             `json:"category" default:"NA"`
	Created_at  time.Time          `json:"created_at"`
	Updated_at  time.Time          `json:"updated_at"`
}
