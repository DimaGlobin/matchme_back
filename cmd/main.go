package main

import (
	// "fmt"

	"net/http"

	"github.com/DimaGlobin/matchme/controllers"
	"github.com/DimaGlobin/matchme/initializers"
	"github.com/DimaGlobin/matchme/middleware"
	"github.com/DimaGlobin/matchme/models"
	"github.com/DimaGlobin/matchme/services"
	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
	initializers.SyncDatabase()
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("../templates/*")

	r.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup_form.html", gin.H{})
	})
	r.GET("/login", func(c *gin.Context) {
		var user *models.User
		user, _ = services.GetUserFromReq(c)
		if user.ID != 0 {
			c.Redirect(http.StatusMovedPermanently, "/rate_users")
		}
		c.HTML(http.StatusOK, "login_form.html", gin.H{})
	})

	r.POST("/signup", controllers.SignUp)
	r.POST("/login", controllers.Login)
	r.GET("/validate", middleware.RequireAuth, controllers.Validate)
	r.GET("/rate_users", middleware.RequireAuth, controllers.ShowRandomUser)
	r.POST("/react", middleware.RequireAuth, controllers.HandleReaction)
	r.POST("/upload_photo", middleware.RequireAuth, controllers.UploadPhoto)
	r.GET("/get_photo/:hash", middleware.RequireAuth, controllers.GetPhoto)
	r.DELETE("/delete_photo/:hash", middleware.RequireAuth, controllers.DeletePhoto)
	r.DELETE("/delete_user", middleware.RequireAuth, controllers.DeleteUser)
	r.POST("/logout", middleware.RequireAuth, controllers.Logout, func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/login")
	})

	r.Run()
}
