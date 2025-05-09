package middlewares

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"reviser/internal/inits"
	"reviser/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func RequireAuth(ctx *gin.Context) {
	tokenString, err := ctx.Cookie("Authorization")

	if err != nil {
		ctx.JSON(401, gin.H{"error": "unauthorized", "message": "No token provided"})
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		ctx.JSON(401, gin.H{"error": "Invalid token"})
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			ctx.JSON(401, gin.H{"error": "unauthorized", "message": "Token expired"})
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Validate the user using the username from the claims
		username, ok := claims["username"].(string)
		if !ok {
			ctx.JSON(401, gin.H{"error": "unauthorized", "message": "Invalid token payload"})
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var user models.User
		row := inits.DB.QueryRow("SELECT Name, Username FROM users WHERE username = $1 LIMIT 1", username)
		err := row.Scan(&user.Name, &user.Username)
		if err != nil {
			if err == sql.ErrNoRows {
				ctx.JSON(401, gin.H{"error": "unauthorized", "message": "User not found"})
				ctx.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			ctx.JSON(500, gin.H{"error": "Failed to query user", "details": err.Error()})
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.Set("user", user)
	} else {
		ctx.JSON(401, gin.H{"error": "Invalid token claims"})
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.Next()
}
