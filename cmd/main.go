package main

import (
	// "fmt"

	"github.com/DimaGlobin/matchme/controllers"
	"github.com/DimaGlobin/matchme/initializers"
	"github.com/DimaGlobin/matchme/middleware"
	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
	initializers.SyncDatabase()
}

func main() {
	r := gin.Default()

	r.POST("/signup", controllers.SignUp)
	r.POST("/login", controllers.Login)
	r.GET("/validate", middleware.RequireAuth, controllers.Validate)
	r.GET("/rate_users", middleware.RequireAuth, controllers.ShowRandomUser)
	r.POST("/react", middleware.RequireAuth, controllers.HandleReaction)
	r.POST("/upload_photo", middleware.RequireAuth, controllers.UploadPhoto)
	r.GET("/get_photo/:hash", middleware.RequireAuth, controllers.GetPhoto)
	r.DELETE("/delete_photo/:hash", middleware.RequireAuth, controllers.DeletePhoto)
	r.DELETE("/delete_user", middleware.RequireAuth, controllers.DeleteUser)
	r.POST("/logout", middleware.RequireAuth, controllers.Logout)

	r.Run()
}
