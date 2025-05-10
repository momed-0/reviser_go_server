package controllers

import (
	"database/sql"
	"net/http"
	"os"
	"reviser/internal/inits"
	"reviser/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var domain = os.Getenv("DOMAIN")
var secure = domain != "localhost"

func Signup(ctx *gin.Context) {
	var body struct {
		Name     string
		Username string
		Password string
	}

	if ctx.BindJSON(&body) != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request", "message": "Error trying to parse body!"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		ctx.JSON(500, gin.H{"error": err, "message": "Error Trying to generate hash!"})
		return
	}

	user := models.User{Name: body.Name, Username: body.Username, Password: string(hash)}
	_, err = inits.DB.Exec("INSERT INTO users (Name, Username, Password) VALUES ($1, $2, $3)",
		user.Name, user.Username, user.Password)

	if err != nil {
		ctx.JSON(500, gin.H{"error": err, "message": "Error Trying to insert credentials to db!"})
		return
	}

	ctx.JSON(200, gin.H{"data": user.Name})
}

func Login(ctx *gin.Context) {
	var body struct {
		Username string
		Password string
	}

	if ctx.BindJSON(&body) != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request", "message": "Error trying to parse body!"})
		return
	}

	var user models.User

	row := inits.DB.QueryRow("SELECT Name,Username, Password FROM users WHERE Username = $1", body.Username)

	if err := row.Scan(&user.Name, &user.Username, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(500, gin.H{"error": "User not found"})
			return
		}
		ctx.JSON(400, gin.H{"error": "Invalid request", "message": "Error Trying to check credentials from db!"})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))

	if err != nil {
		ctx.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	// generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      jwt.TimeFunc().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		ctx.JSON(500, gin.H{"error": "error signing token"})
		return
	}
	if secure {
		ctx.SetSameSite(http.SameSiteNoneMode)
	} else {
		ctx.SetSameSite(http.SameSiteLaxMode)
	}
	ctx.SetCookie("Authorization", tokenString, 3600*24*30, "/", domain, secure, true)
	ctx.JSON(200, gin.H{"data": "Successfully logged in!", "user": body.Username})
}

func Validate(ctx *gin.Context) {
	// Retrieve the user from the context (set by AuthMiddleware)
	_, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(500, gin.H{"error": "User not found in context"})
		return
	}
	// Return a success response with the user details
	ctx.JSON(200, gin.H{"data": "You are logged in!"})
}

func Logout(ctx *gin.Context) {
	if secure {
		ctx.SetSameSite(http.SameSiteNoneMode)
	} else {
		ctx.SetSameSite(http.SameSiteLaxMode)
	}
	ctx.SetCookie("Authorization", "", -1, "/", domain, secure, true)
	ctx.JSON(200, gin.H{"data": "You are logged out!"})
}
