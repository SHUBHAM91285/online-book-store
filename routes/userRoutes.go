package routes

import (
	controller "github.com/SHUBHAM91285/online_book_store/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/user/signup", controller.SignUp())
	incomingRoutes.POST("/user/login", controller.Login())
	incomingRoutes.GET("/user/profile", controller.UserProfile())
	incomingRoutes.PATCH("/user/profile/password", controller.UpdatePassword())
	incomingRoutes.PATCH("/cart/add", controller.AddBookToCart())
	incomingRoutes.PATCH("/cart/update/:id", controller.UpdateBookQuantity())
	incomingRoutes.PATCH("/cart/remove/:id", controller.RemoveBookFromCart())
}
