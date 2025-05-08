package inits

import (
	"log"
	"reviser/internal/models"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func DBInit() {
	// Use an in-memory SQLite database for testing and development
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the in-memory database: %v", err)
	}

	// Run migrations to create the User and Question tables
	err = db.AutoMigrate(&models.User{}, &models.Question{}, &models.Question_Tags{}, &models.Leetcode_submissions{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	// Seed the database with initial data
	seedDatabase(db)

	DB = db
	log.Println("In-memory database initialized successfully with User and Question tables")
}

func seedDatabase(db *gorm.DB) {

	// Seed Questions
	questions := []models.Question{
		{Slug: "3sum", Title: "3Sum", Description: "Explain the basics of Golang."},
		{Slug: "trapping-rain-water", Title: "Trapping Rain Water?", Description: "rain rain go away"},
	}
	for _, question := range questions {
		db.Create(&question)
	}

	// Seed Question_Tags
	questionTags := []models.Question_Tags{
		{Slug: "3sum", Tags: []string{"two pointers"}},
		// {Slug: "trapping-rain-water", Tags: []string{"two pointers"}},
	}
	for _, qt := range questionTags {
		db.Create(&qt)
	}
	leetcodeSubmissions := []models.Leetcode_submissions{
		{Submission_ID: 1596890781, Slug: "3sum", Code: "package main\n\nfunc main() {}", Submitted_At: time.Now()},
		{Submission_ID: 1597307455, Slug: "3sum", Code: "package main\n\nfunc main() {}", Submitted_At: time.Now()},
		{Submission_ID: 1599797703, Slug: "trapping-rain-water", Code: "package main\n\nfunc main() {}", Submitted_At: time.Now()},
		{Submission_ID: 1599784766, Slug: "trapping-rain-water", Code: "package main\n\nfunc main() {}", Submitted_At: time.Now()},
	}
	for _, submission := range leetcodeSubmissions {
		db.Create(&submission)
	}
}
