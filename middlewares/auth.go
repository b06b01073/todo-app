package middlewares

import (
	"net/http"
	"todo-app/jwtService"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		BEARER_SCHEMA := "BEARER "
		prefixLen := len(BEARER_SCHEMA)
		requestAuth := c.GetHeader("Authorization")
		if len(requestAuth) <= prefixLen {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid token",
			})
			return
		}

		tokenString := c.GetHeader("Authorization")[prefixLen:]
		username, err := jwtService.ParseJWT(tokenString)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}
		c.Set("username", username)
		c.Next()
	}
}
