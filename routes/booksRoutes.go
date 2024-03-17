package routes

import (
	controller "github.com/SHUBHAM91285/online_book_store/controllers"

	"github.com/gin-gonic/gin"
)

func BooksRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/books", controller.GetBooks())
	incomingRoutes.GET("/books/:parameter", controller.GetBookByParameter())
	incomingRoutes.POST("/admin/book", controller.AddBook())
	incomingRoutes.PATCH("/admin/book/:book_id", controller.UpdateBookInfo())
	incomingRoutes.DELETE("/admin/book/:book_id", controller.DeleteBook())
}
