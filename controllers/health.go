package controllers

import (
	"github.com/gin-gonic/gin"
)

func HealthCheck(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status":  "ok",
		"message": "Server is running",
	})
}
