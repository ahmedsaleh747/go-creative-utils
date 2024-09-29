package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/ahmedsaleh747/go-creative-utils/shared"

	"github.com/dgrijalva/jwt-go"
)

const expirationHours = 48

var jwtKey = []byte{}

// Should be called from the child modules to configure the jwtKey“
func ConfigureJWT(appJwtKey []byte) {
	jwtKey = appJwtKey
}

func GenerateToken(claims *shared.UserMeta) (string, error) {
	expirationTime := time.Now().Add(expirationHours * time.Hour)
	claims.StandardClaims = jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func VerifyToken(tokenString string) (*shared.UserMeta, error) {
	claims := &shared.UserMeta{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, fmt.Errorf("invalid signature")
		}
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
