package main

import (
	"reviser/controllers"
	"reviser/internal/inits"
	"reviser/middlewares"

	"github.com/gin-gonic/gin"
)

func init() {

	inits.LoadEnv()
	inits.DBInit()
}

// main function is the entry point of the application
func main() {
	r := gin.Default()

	// Middleware to handle CORS
	r.Use(middlewares.CORSMiddleware())

	// Middleware to log requests
	r.Use(middlewares.Logger())

	// Health Check route
	r.GET("/health", controllers.HealthCheck)

	// Authentication routes
	{
		authGroups := r.Group("/auth")
		authGroups.POST("/validate", middlewares.RequireAuth, controllers.Validate)
		authGroups.POST("/signup", controllers.Signup)
		authGroups.POST("/login", controllers.Login)
		authGroups.POST("/logout", middlewares.RequireAuth, controllers.Logout)

	}
	// Content routes
	{
		contentRoutes := r.Group("/api/content")
		contentRoutes.Use(middlewares.RequireAuth)
		contentRoutes.GET("/questions/count", controllers.FetchQuestionsCount)
		contentRoutes.GET("/questions/all", controllers.FetchAllQuestions)
		contentRoutes.GET("/submissions/:slug", controllers.FetchSubmissionsBySlug)
		contentRoutes.GET("/submissions", controllers.FetchSubmissionsForDay)
		contentRoutes.GET("/pages", controllers.FetchSubmissionsRange)
		contentRoutes.GET("/tags", controllers.FetchTagsBySlug)
		contentRoutes.POST("/tags/editor/upsert", controllers.UpsertTags)
		contentRoutes.DELETE("/tags/editor", controllers.DeleteTags)
	}

	// cron job routes
	{
		cronJobRoutes := r.Group("/api/cron")
		cronJobRoutes.Use(middlewares.RequireAuth)
		cronJobRoutes.POST("/questions/insert", controllers.InsertQuestions)
		cronJobRoutes.POST("/submissions/insert", controllers.InsertSubmissions)
	}

	r.Run()
}
