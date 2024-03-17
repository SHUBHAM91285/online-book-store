package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id"`
	Name     string             `json:"name" validate:"required"`
	Email    string             `json:"email" validate:"required"`
	Password string             `json:"password" validate:"required"`
	Role     string             `json:"role" enum:"user,admin" default:"user" validate:"required"`
	Cart     []Cart             `json:"cart"`
}

type Cart struct {
	ID       primitive.ObjectID `json:"id"`
	Name     string             `json:"name" validate:"required"`
	Price    int                `json:"price" validate:"required"`
	Quantity int                `json:"quantity" default:"1"`
	Author   string             `json:"author" validate:"required"`
	Amount   int                `json:"amount"`
}
