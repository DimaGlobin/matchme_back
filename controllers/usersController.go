package controllers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/DimaGlobin/matchme/initializers"
	"github.com/DimaGlobin/matchme/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SignUp(c *gin.Context) {
	// Get the email/pas off req body
	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Sex      string `json:"sex" binding:"required"`
		Location string `json:"location" binding:"required"`
	}

	if c.ShouldBindJSON(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})

		return
	}
	//Hash password

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
		return
	}
	//Create the user
	user := models.User{Email: body.Email, Password: string(hash), Sex: body.Sex}
	result := initializers.DB.Create(&user)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user",
		})
		return
	}
	//Respond
	c.JSON(http.StatusOK, gin.H{})
}

func Login(c *gin.Context) {
	//Get the email and pas off req body

	var body struct {
		Email    string
		Password string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})

		return
	}

	//Look up requested user

	var user models.User
	initializers.DB.First(&user, "email = ?", body.Email)

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email or password",
		})

		return
	}

	//Compare sent in pass with saved user pas hash

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email or password",
		})

		return
	}

	//Generate a jwt-token

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create token",
		})
	}

	//Send it back
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{})
}

func Validate(c *gin.Context) {

	user, _ := c.Get("user")

	c.JSON(http.StatusOK, gin.H{
		"mesage": user,
	})
}

func getUserFromReq(c *gin.Context) *models.User {
	//Get the cookie of request

	tokenString, err := c.Cookie("Authorization")

	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	//Decode/validate it

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("SECRET")), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		//Check the exp

		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		//Find the user with token sub

		var user models.User

		initializers.DB.First(&user, claims["sub"])

		if user.ID == 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		//Attach to request

		c.Set("user", user)

		fmt.Println(claims["foo"], claims["nbf"])

		return &user
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
		return nil
	}
}

func ShowRandomUser(c *gin.Context) {

	user := *(getUserFromReq(c))
	var RateUser models.User

	var result *gorm.DB

	if user.Sex == "male" {
		result = initializers.DB.Where("sex = ?", "female").Order("RANDOM()").First(&RateUser)
	} else if user.Sex == "female" {
		result = initializers.DB.Where("sex = ?", "male").Order("RANDOM()").First(&RateUser)
	}

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to find random user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": RateUser,
	})
}

func checkAnswer(ans string) bool {
	if ans == "like" || ans == "dislike" {
		return true
	} else {
		return false
	}
}

func HandleReaction(c *gin.Context) {
	userFromReq := getUserFromReq(c)
	var user models.User

	var body struct {
		ProfileID int    `json:"profile_id" binding:"required"`
		Answer    string `json:"answer" binding:"required"`
	}

	if !checkAnswer(body.Answer) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unavailable answer",
		})
	}

	if err := initializers.DB.First(&user, userFromReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Can't find user in Database",
		})
	}

	user.liked = append(user.liked, body.ProfileID)

	if err := initializers.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Can't save changes",
		})
	}
}
