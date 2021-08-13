package jwtService

import (
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	_ "github.com/joho/godotenv/autoload"
)

// this struct is the claim part of the jwt
type claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var secret = []byte(os.Getenv("JWT_SECRET"))

// GenerateJWT function is called when user sign in or sign up an account, the claim of JWT takes only the username as
// an argument, and return the jwt to the caller
func GenerateJWT(username string) (string, error) {
	expireTime := time.Now().Add(300 * time.Second).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime,
		},
		Username: username,
	})

	JWTString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return JWTString, nil
}

func ParseJWT(tokenString string) (string, error) {
	c := claims{}

	token, err := jwt.ParseWithClaims(tokenString, &c, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return "", err
	}

	if token == nil || !token.Valid {
		return "", err
	}

	// if there are no errors above, then the claims from client is trust-worthy.
	return c.Username, nil
}
