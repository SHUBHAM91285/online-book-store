package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/SHUBHAM91285/online_book_store/database"
	"github.com/SHUBHAM91285/online_book_store/models"
	"github.com/SHUBHAM91285/online_book_store/tokens"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		adminCount, err := userCollection.CountDocuments(ctx, bson.M{"role": "admin"})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking admin count"})
			return
		}

		if adminCount > 0 && user.Role == "admin" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Only one admin user is allowed"})
			return
		}
		err = userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&user)
		if err == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user already exist"})
			return
		}
		defer cancel()
		user.ID = primitive.NewObjectID()
		password := HashPassword(user.Password)
		user.Password = password
		tokenString, err := tokens.TokenGenerator(user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		userInfo, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User is not created"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"userInfo": userInfo,
			"token":    tokenString,
		})
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found,login seems to be incorrect"})
			return
		}
		passwordIsValid := VerifyPassword(user.Password, foundUser.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "password is invalid"})
			return
		}
		tokenString, err := tokens.TokenGenerator(foundUser.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "you are welcomed",
			"token":   tokenString,
		})
	}
}

func UserProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var foundUser models.User

		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		tokenString = tokenString[len("Bearer "):]

		claims, msg := tokens.VerifyToken(tokenString)
		if msg != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		err := userCollection.FindOne(ctx, bson.M{"email": claims.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"name":  foundUser.Name,
			"email": foundUser.Email,
			"role":  foundUser.Role,
			"cart":  foundUser.Cart,
		})
	}
}

func UpdatePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var foundUser models.User
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		tokenString = tokenString[len("Bearer "):]

		claims, msg := tokens.VerifyToken(tokenString)
		if msg != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		err := userCollection.FindOne(ctx, bson.M{"email": claims.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}

		hashedPassword := HashPassword(user.Password)
		update := bson.M{"$set": bson.M{"password": hashedPassword}}
		_, err = userCollection.UpdateOne(ctx, bson.M{"email": foundUser.Email}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
	}
}

func AddBookToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var foundUser models.User
		var book models.Books
		var foundBook models.Books
		var cart models.Cart

		if err := c.BindJSON(&book); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		tokenString = tokenString[len("Bearer "):]

		claims, msg := tokens.VerifyToken(tokenString)
		if msg != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		err := userCollection.FindOne(ctx, bson.M{"email": claims.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}
		err = booksCollection.FindOne(ctx, bson.M{"name": book.Name}).Decode(&foundBook)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "book not found"})
			return
		}
		cart.ID = primitive.NewObjectID()
		cart.Name = foundBook.Name
		cart.Price = foundBook.Price
		cart.Author = foundBook.Author_name
		cart.Quantity = 1
		cart.Amount = cart.Quantity * cart.Price
		foundUser.Cart = append(foundUser.Cart, cart)
		_, err = userCollection.UpdateOne(
			ctx,
			bson.M{"_id": foundUser.ID},
			bson.M{"$set": bson.M{"cart": foundUser.Cart}},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add item in the cart"})
			return
		}
		c.JSON(http.StatusOK, "item successfully added to the cart")
	}
}

func UpdateBookQuantity() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		tokenString = tokenString[len("Bearer "):]

		claims, msg := tokens.VerifyToken(tokenString)
		if msg != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		var foundUser models.User
		err := userCollection.FindOne(ctx, bson.M{"email": claims.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}

		cartID := c.Param("id")
		if cartID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cart_id is required"})
			return
		}
		var updatedCart []models.Cart
		var updatedItem models.Cart
		var itemFound bool
		for _, cartItem := range foundUser.Cart {
			if cartItem.ID.Hex() == cartID {
				cartItem.Quantity++
				cartItem.Amount = cartItem.Price * cartItem.Quantity
				updatedItem = cartItem
				itemFound = true
			}
			updatedCart = append(updatedCart, cartItem)
		}

		if !itemFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "cart item not found"})
			return
		}

		_, err = userCollection.UpdateOne(ctx, bson.M{"_id": foundUser.ID}, bson.M{"$set": bson.M{"cart": updatedCart}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update cart"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "cart item quantity increased successfully", "updated_item": updatedItem})

	}
}

func RemoveBookFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		tokenString = tokenString[len("Bearer "):]

		claims, msg := tokens.VerifyToken(tokenString)
		if msg != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		var foundUser models.User
		err := userCollection.FindOne(ctx, bson.M{"email": claims.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}

		cartIDToRemove := c.Param("id")
		if cartIDToRemove == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cart_id is required"})
			return
		}

		var indexToRemove = -1
		for i, cartItem := range foundUser.Cart {
			if cartItem.ID.Hex() == cartIDToRemove {
				indexToRemove = i
				break
			}
		}

		if indexToRemove == -1 {
			c.JSON(http.StatusNotFound, gin.H{"error": "cart item not found"})
			return
		}

		foundUser.Cart = append(foundUser.Cart[:indexToRemove], foundUser.Cart[indexToRemove+1:]...)

		_, err = userCollection.UpdateOne(ctx, bson.M{"_id": foundUser.ID}, bson.M{"$set": bson.M{"cart": foundUser.Cart}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete item from cart"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "item deleted successfully from cart"})

	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword, providedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	if err != nil {
		check = false
	}
	return check
}
