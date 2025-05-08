package controllers

import (
	"reviser/internal/inits"
	"reviser/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FetchTagsBySlug retrieves tags by slug from the database
func FetchTagsBySlug(ctx *gin.Context) {
	slug := ctx.Query("slug")
	if slug == "" {
		ctx.JSON(400, gin.H{"error": "Slug is required"})
		return
	}
	// Fetch the tags from the database using the slug
	var tags models.Question_Tags
	results := inits.DB.First(&tags, "slug = ?", slug)
	if results.Error != nil {
		ctx.JSON(404, gin.H{"error": "Tags not found"})
		return
	}
	ctx.JSON(200, gin.H{"tags": tags.Tags})
}

// FetchQuestionsCount retrieves the count of questions from the database
func FetchQuestionsCount(ctx *gin.Context) {
	var count int64
	results := inits.DB.Table("question_tags").Count(&count)
	if results.Error != nil {
		ctx.JSON(500, gin.H{"error": results.Error})
		return
	}
	ctx.JSON(200, gin.H{"count": count})
}

// FetchAllQuestions retrieves all questions from the database
func FetchAllQuestions(ctx *gin.Context) {
	var questions []models.Question
	results := inits.DB.Find(&questions)
	if results.Error != nil {
		ctx.JSON(500, gin.H{"error": results.Error})
		return
	}
	ctx.JSON(200, gin.H{"questions": questions})
}

// FetchSubmissionsBySlug retrieves submissions by slug from the database
// with joined question data
func FetchSubmissionsBySlug(ctx *gin.Context) {
	slug := ctx.Param("slug")
	if slug == "" {
		ctx.JSON(400, gin.H{"error": "Slug is required"})
		return
	}
	var submissions []models.Leetcode_submissions
	results := inits.DB.Preload("Question").Where("slug = ?", slug).Find(&submissions)
	if results.Error != nil {
		ctx.JSON(404, gin.H{"error": "Question not found"})
		return
	}
	ctx.JSON(200, gin.H{"submissions": submissions})
}

// FetchSubmissionsForDay retrieves submissions for a specific day
// with joined question data
func FetchSubmissionsForDay(ctx *gin.Context) {
	date := ctx.Query("date")
	if date == "" {
		ctx.JSON(400, gin.H{"error": "Date is required"})
		return
	}
	// Parse the date string into a time.Time object (only date part)
	const layout = "2006-01-02"
	startOfDay, error := time.Parse(layout, date)
	if error != nil {
		ctx.JSON(400, gin.H{"error": "Invalid date format"})
		return
	}
	endOfDay := startOfDay.Add(time.Hour*23 + time.Minute*59 + time.Second*59)

	var submissions []models.Leetcode_submissions
	err := inits.DB.Preload("Question").
		Where("submitted_at >= ? AND submitted_at <= ?", startOfDay, endOfDay).
		Order("submitted_at DESC").
		Find(&submissions).Error

	if err != nil {
		ctx.JSON(500, gin.H{"error": "Internal Error"})
		return
	}
	ctx.JSON(200, gin.H{"submissions": submissions})
}

// FetchSubmissionsRange retrieves paginated response
// within from and two in query string parameters
func FetchSubmissionsRange(ctx *gin.Context) {
	from, err := strconv.Atoi(ctx.Query("from"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "invalid 'from' query parameter"})
		return
	}
	var to int
	to, err = strconv.Atoi(ctx.Query("to"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "invalid 'to' query parameter"})
		return
	}

	var submissions []models.Leetcode_submissions
	err = inits.DB.Preload("Question").
		Limit(to - from + 1).
		Offset(from).
		Find(&submissions).Error
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Internal Error"})
		return
	}
	ctx.JSON(200, gin.H{"submissions": submissions})
}

// UpsertTags will insert the tags if it doesn;t exists,
// if exists it will update the tags
func UpsertTags(ctx *gin.Context) {
	var quesTag models.Question_Tags
	if err := ctx.ShouldBindJSON(&quesTag); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON payload", "details": err.Error()})
		return
	}

	err := inits.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{"tags"}),
	}).Create(&quesTag).Error

	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to upsert tags", "details": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"status": "Tags updated successfully"})
}

// DeleteTags will remove the tags entry for particular slug
func DeleteTags(ctx *gin.Context) {
	slug := ctx.Query("slug")
	if slug == "" {
		ctx.JSON(400, gin.H{"error": "Slug is required"})
		return
	}

	result := inits.DB.Where("slug = ?", slug).Delete(&models.Question_Tags{})
	if result.Error != nil {
		ctx.JSON(500, gin.H{"error": "Failed to delete tags", "details": result.Error.Error()})
		return
	}

	ctx.JSON(200, gin.H{"status": "Tags deleted successfully"})
}

// InsertQuestions will upsert the questions into db
func InsertQuestions(ctx *gin.Context) {
	var question models.Question
	if err := ctx.ShouldBindJSON(&question); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON payload", "details": err.Error()})
		return
	}

	err := inits.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&question).Error

	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to upsert Question", "details": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"status": "Question upserted succesfully"})
}

func InsertSubmissions(ctx *gin.Context) {
	var submission models.Leetcode_submissions
	if err := ctx.ShouldBindJSON(&submission); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON payload", "details": err.Error()})
		return
	}

	// Check if the referenced Question exists
	var question models.Question
	if err := inits.DB.Where("slug = ?", submission.Slug).First(&question).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(400, gin.H{"error": "Referenced question not found", "slug": submission.Slug})
			return
		}
		ctx.JSON(500, gin.H{"error": "Failed to query question", "details": err.Error()})
		return
	}

	// Create the submission
	if err := inits.DB.Create(&submission).Error; err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to insert submission", "details": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"status": "Submission inserted successfully"})
}
