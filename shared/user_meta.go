package shared

import "github.com/dgrijalva/jwt-go"

type UserMeta struct {
	jwt.StandardClaims
	UserId   uint   `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Name     string `json:"name"`
}
