package controllers

import (
	"net/http"
	"os"
	"reviser/internal/inits"
	"reviser/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func Signup(ctx *gin.Context) {
	var body struct {
		Name     string
		Username string
		Password string
	}

	if ctx.BindJSON(&body) != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		ctx.JSON(500, gin.H{"error": err})
		return
	}

	user := models.User{Name: body.Name, Username: body.Username, Password: string(hash)}
	result := inits.DB.Create(&user)

	if result.Error != nil {
		ctx.JSON(500, gin.H{"error": result.Error})
		return
	}

	ctx.JSON(200, gin.H{"data": user})
}

func Login(ctx *gin.Context) {
	var body struct {
		Username string
		Password string
	}

	if ctx.BindJSON(&body) != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	result := inits.DB.Where("username = ?", body.Username).First(&user)

	if result.Error != nil {
		ctx.JSON(500, gin.H{"error": "User not found"})
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

	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("Authorization", tokenString, 3600*24*30, "", "localhost", false, true)
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
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("Authorization", "", -1, "", "localhost", false, true)
	ctx.JSON(200, gin.H{"data": "You are logged out!"})
}
