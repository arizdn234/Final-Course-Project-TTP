package model

import "github.com/golang-jwt/jwt"

var JwtKey = []byte("secret-key")

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}
