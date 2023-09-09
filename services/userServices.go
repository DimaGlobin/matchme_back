package services

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/DimaGlobin/matchme/initializers"
	"github.com/DimaGlobin/matchme/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

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
			return nil, nil
		}

		//Find the user with token sub

		var user models.User

		initializers.DB.First(&user, claims["sub"])

		if user.ID == 0 {
			return nil, nil
		}

		//Attach to request

		c.Set("user", user)

		fmt.Println(claims["foo"], claims["nbf"])

		retClaim := claims["sub"]

		return &user, retClaim
	} else {
		return nil, nil
	}
}
