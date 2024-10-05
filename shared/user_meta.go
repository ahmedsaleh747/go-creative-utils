package shared

import (
	"github.com/ahmedsaleh747/go-creative-utils/storage"
	"github.com/dgrijalva/jwt-go"
)

type IdentityClaims interface {
	jwt.Claims
	GetUserId() uint
	GetUsername() string
	GetRole() string
	SetStandardClaims(jwt.StandardClaims)
	SetClaims(storage.Identity)
}

type UserMeta struct {
	jwt.StandardClaims
	IdentityClaims
	UserId   uint   `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Name     string `json:"name"`
}

func (claims *UserMeta) GetUserId() uint {
	return claims.UserId
}

func (claims *UserMeta) GetUsername() string {
	return claims.Username
}

func (claims *UserMeta) GetRole() string {
	return claims.Role
}

func (claims *UserMeta) SetStandardClaims(standardClaims jwt.StandardClaims) {
	claims.StandardClaims = standardClaims
}

func (claims *UserMeta) SetClaims(user storage.Identity) {
	claims.UserId = user.GetId()
	claims.Username = user.GetName()
	claims.Role = user.GetRole()
	//No extras, should be overridden by the child structs
}
