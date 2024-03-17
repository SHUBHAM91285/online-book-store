package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/SHUBHAM91285/online_book_store/database"
	"github.com/SHUBHAM91285/online_book_store/tokens"

	"github.com/SHUBHAM91285/online_book_store/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var booksCollection *mongo.Collection = database.OpenCollection(database.Client, "books")
var validate = validator.New()

func GetBooks() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		result, err := booksCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing books"})
			return
		}
		var booksNames []string
		for result.Next(ctx) {
			var book models.Books
			if err := result.Decode(&book); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode book"})
				return
			}
			booksNames = append(booksNames, book.Name)
		}
		c.JSON(http.StatusOK, booksNames)
	}
}
func GetBookByParameter() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		Parameter := c.Param("parameter")
		var books []models.Books
		filter := bson.M{
			"$or": []bson.M{
				{"name": Parameter},
				{"author_name": Parameter},
				{"genre": Parameter},
				{"description": Parameter},
				{"publication": Parameter},
				{"category": Parameter},
			},
		}
		cursor, err := booksCollection.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding books"})
			return
		}
		defer cursor.Close(ctx)
		if err := cursor.All(ctx, &books); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding books"})
			return
		}
		c.JSON(http.StatusOK, books)
	}
}

func AddBook() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var foundUser models.User
		var book models.Books

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

		if foundUser.Role != "admin" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "only admin is allowed to add books"})
			return
		}
		if err := c.BindJSON(&book); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		book.ID = primitive.NewObjectID()
		book.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		book.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		_, insertErr := booksCollection.InsertOne(ctx, book)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Book is not created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, gin.H{"message": "book inserted properly"})
	}
}
func UpdateBookInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var foundUser models.User
		var book models.Books

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
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}
		if foundUser.Role != "admin" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "only admin is allowed to add books"})
			return
		}

		bookID := c.Param("book_id")
		objID, err := primitive.ObjectIDFromHex(bookID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
			return
		}

		if err := c.BindJSON(&book); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updateObj := bson.D{}

		if book.Author_info != "" {
			updateObj = append(updateObj, bson.E{"author_info", book.Author_info})
		}
		if book.Author_name != "" {
			updateObj = append(updateObj, bson.E{"author_name", book.Author_name})
		}
		if book.Category != "" {
			updateObj = append(updateObj, bson.E{"category", book.Category})
		}
		if book.Description != "" {
			updateObj = append(updateObj, bson.E{"description", book.Description})
		}
		if book.Genre != "" {
			updateObj = append(updateObj, bson.E{"genre", book.Genre})
		}
		if book.Name != "" {
			updateObj = append(updateObj, bson.E{"name", book.Name})
		}
		if book.Publication != "" {
			updateObj = append(updateObj, bson.E{"publication", book.Publication})
		}
		book.Updated_at = time.Now()

		updateObj = append(updateObj, bson.E{"updated_at", book.Updated_at})

		filter := bson.M{"_id": objID}

		_, err = booksCollection.UpdateOne(
			ctx,
			filter,
			bson.D{{"$set", updateObj}},
		)

		if err != nil {
			fmt.Println(err)
			msg := fmt.Sprintf("book update failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusOK, "data updated successfully")

	}
}

func DeleteBook() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

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
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}
		if foundUser.Role != "admin" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "only admin is allowed to add books"})
			return
		}
		bookID := c.Param("book_id")
		objID, err := primitive.ObjectIDFromHex(bookID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
			return
		}
		result, err := booksCollection.DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete book"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "book deleted successfully"})
	}
}
