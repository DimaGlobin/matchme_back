package controllers

/*
TODO:
2) Добавить ограничение лайков (50 в сутки для пользователя без подписки)
3) Сделать таблицу лайков для дальнеёшего её использования в рекомендательной системе
*/

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/DimaGlobin/matchme/initializers"
	"github.com/DimaGlobin/matchme/models"
	"github.com/DimaGlobin/matchme/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SignUp(c *gin.Context) {
	// Get the email/pas off req body
	var body struct {
		Email     string `json:"email" binding:"required"`
		Password  string `json:"password" binding:"required"`
		Sex       string `json:"sex" binding:"required"`
		Location  string `json:"location" binding:"required"`
		BirthDate string `json:"birthdate" binding:"required"`
		MaxAge    int    `json:"maxage" binding:"required"`
		Radius    int    `json:"radius" binding:"required"`
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

	bDate, err := utils.ParseDate(body.BirthDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse date",
		})
		return
	}

	age := utils.CalculateAge(bDate)
	if age < 18 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Age less than 18",
		})
		return
	}

	user := models.User{
		Email:    body.Email,
		Password: string(hash),
		Sex:      body.Sex, Location: body.Location,
		BirthDate: bDate,
		Age:       age,
		MaxAge:    body.MaxAge,
		Radius:    body.Radius,
		Rights:    "DEFAULT",
	}

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

func GetLoaction(c *gin.Context) *models.LocationInfo {
	// ip := c.ClientIP()
	ip := "93.115.28.181"

	apiURL := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	response, err := http.Get(apiURL)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	defer response.Body.Close()

	var location models.LocationInfo
	if err := json.NewDecoder(response.Body).Decode(&location); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return nil
	}

	location.LastIP = ip
	fmt.Println("Location info: ", location)

	return &location
}

func Login(c *gin.Context) {
	//Get the email and pas off req body

	location := GetLoaction(c)

	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if c.ShouldBind(&body) != nil {
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

	user.LastIP = location.LastIP
	user.Longitude = location.Lon
	user.Latitude = location.Lat

	age := utils.CalculateAge(user.BirthDate)
	if user.Age < age {
		user.Age = age
	}

	result := initializers.DB.Save(user)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to save info",
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

func GetUserFromReq(c *gin.Context) (*models.User, interface{}) {
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

		retClaim := claims["sub"]

		return &user, retClaim
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
		return nil, nil
	}
}

func ShowRandomUser(c *gin.Context) {

	var user *models.User
	user, _ = GetUserFromReq(c)
	var RateUser models.User

	subQuery := fmt.Sprintf(`
        SELECT unnest(array_cat(liked, disliked)) FROM users WHERE id = %d
    `, user.ID)

	rows, err := initializers.DB.Raw(
		"SELECT * FROM users WHERE sex != ? AND id != ? AND id NOT IN ("+subQuery+") AND location = ? AND deleted_at IS NULL AND max_age <= ? ORDER BY RANDOM() LIMIT 1",
		user.Sex, user.ID, user.Location, user.MaxAge,
	).Rows()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to find random user",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		initializers.DB.ScanRows(rows, &RateUser)
		break // Выход из цикла после первой строки
	}

	Distance := utils.VincentyDistance(user.Latitude, user.Longitude, RateUser.Latitude, RateUser.Longitude) / 1000
	if math.IsNaN(Distance) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to calculate distance",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user":     RateUser,
		"distance": int(Distance),
	})
}

func HandleReaction(c *gin.Context) {
	_, claim := GetUserFromReq(c)

	var user models.User
	var RateUser models.User

	var body struct {
		ProfileID int    `json:"profile_id" binding:"required"`
		Answer    string `json:"answer" binding:"required"`
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})

		return
	}

	initializers.DB.First(&user, claim)
	initializers.DB.First(&RateUser, body.ProfileID)

	if user.ID == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	if body.Answer == "like" {
		err := initializers.DB.Exec("UPDATE users SET liked = array_append(liked, ?) WHERE id = ?", body.ProfileID, user.ID).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update liked array",
			})
			return
		}

		fmt.Println("RateUser liked array: ", RateUser.Liked, "user.ID: ", user.ID)

		if utils.Contains(RateUser.Liked, int64(user.ID)) {
			// Обновляем оба массива сразу
			err = initializers.DB.Exec("UPDATE users SET matches = array_append(matches, ?) WHERE id = ?", user.ID, RateUser.ID).Error
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to update matches",
				})
				return
			}
			err = initializers.DB.Exec("UPDATE users SET matches = array_append(matches, ?) WHERE id = ?", RateUser.ID, user.ID).Error
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to update matches array",
				})
				return
			}
		}
	} else if body.Answer == "dislike" {
		err := initializers.DB.Exec("UPDATE users SET disliked = array_append(disliked, ?) WHERE id = ?", body.ProfileID, user.ID).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update disliked array",
			})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to update rates",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rates updated successfully",
	})
}

func DeleteUser(c *gin.Context) {
	user, _ := GetUserFromReq(c)
	var result *gorm.DB
	var photo models.Photo

	for _, value := range user.PhotoHashes {
		result = initializers.DB.First(&photo, "where hash = ?", value)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to find photo in database",
			})
			return
		}

		filePath := "../temp-files/" + photo.ImageName

		if _, err := os.Stat(filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to find photo in storage",
			})
			return
		}

		err := os.Remove(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete file from object storage",
			})
			return
		}

		result = initializers.DB.Unscoped().Delete(&photo)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete file from database",
			})
			return
		}
	}

	result = initializers.DB.Unscoped().Delete(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed delete user from database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

func Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", "", -1, "", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
