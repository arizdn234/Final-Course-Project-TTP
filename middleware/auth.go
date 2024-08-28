package middleware

import (
	"a21hc3NpZ25tZW50/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func Auth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		cookie, err := c.Cookie("session_token")
		if err != nil {
			if c.GetHeader("Content-Type") == "application/json" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			} else {
				c.Redirect(http.StatusSeeOther, "/client/login")
			}

			c.Abort()
			return
		}

		tokenClaims := &model.Claims{}
		token, err := jwt.ParseWithClaims(cookie, tokenClaims, func(t *jwt.Token) (interface{}, error) {
			return model.JwtKey, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Set("email", tokenClaims.Email)
		c.Next()
	})
}

func CreateToken(email string) (string, error) {
	claims := model.Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	tokenString, err := token.SignedString(model.JwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
