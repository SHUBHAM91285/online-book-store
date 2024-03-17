package main

import (
	"os"

	routes "github.com/SHUBHAM91285/online_book_store/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.BooksRoutes(router)
	routes.UserRoutes(router)
	router.Run(":" + port)
}
