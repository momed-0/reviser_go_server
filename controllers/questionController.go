package controllers

import (
	"database/sql"
	"net/http"
	"reviser/internal/inits"
	"reviser/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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
	query := "SELECT tags FROM question_tags WHERE slug = $1"
	err := inits.DB.QueryRowContext(ctx, query, slug).Scan(&tags.Tags)
	if err == sql.ErrNoRows {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Tags not found"})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	ctx.JSON(200, gin.H{"tags": tags.Tags})
}

// FetchQuestionsCount retrieves the count of questions from the database
func FetchQuestionsCount(ctx *gin.Context) {
	var count int64
	query := "SELECT COUNT(*) FROM question_tags"
	err := inits.DB.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Database error"})
		return
	}
	ctx.JSON(200, gin.H{"count": count})
}

// FetchAllQuestions retrieves all questions from the database
func FetchAllQuestions(ctx *gin.Context) {
	query := "SELECT slug, title, description FROM leetcode_questions"
	rows, err := inits.DB.QueryContext(ctx, query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var questions []models.Leetcode_Questions
	for rows.Next() {
		var q models.Leetcode_Questions
		if err := rows.Scan(&q.Slug, &q.Title, &q.Description); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
			return
		}
		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Row iteration error"})
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

	query := `
		SELECT 
			s.submission_id, s.question_slug, s.code, s.submitted_at,
			q.title, q.description
		FROM leetcode_submissions s
		JOIN leetcode_questions q ON s.question_slug = q.slug
		WHERE s.question_slug = $1
	`
	rows, err := inits.DB.QueryContext(ctx, query, slug)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var results []struct {
		Submission models.Leetcode_submissions
		Question   models.Leetcode_Questions
	}
	for rows.Next() {
		var s struct {
			Submission models.Leetcode_submissions
			Question   models.Leetcode_Questions
		}
		if err := rows.Scan(
			&s.Submission.Submission_ID,
			&s.Submission.Question_Slug,
			&s.Submission.Code,
			&s.Submission.Submitted_At,
			&s.Question.Title,
			&s.Question.Description,
		); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
			return
		}
		results = append(results, s)
	}

	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Row iteration error"})
		return
	}
	ctx.JSON(200, gin.H{"submissions": results})
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
	query := `
		SELECT
		s.submission_id, s.question_slug, s.code, s.submitted_at,
			q.title, q.description
		FROM leetcode_submissions s
		JOIN leetcode_questions q ON s.question_slug = q.slug
		WHERE s.submitted_at BETWEEN $1 AND $2`
	rows, err := inits.DB.QueryContext(ctx, query, startOfDay, endOfDay)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var results []struct {
		Submission models.Leetcode_submissions
		Question   models.Leetcode_Questions
	}
	for rows.Next() {
		var s struct {
			Submission models.Leetcode_submissions
			Question   models.Leetcode_Questions
		}
		if err := rows.Scan(
			&s.Submission.Submission_ID,
			&s.Submission.Question_Slug,
			&s.Submission.Code,
			&s.Submission.Submitted_At,
			&s.Question.Title,
			&s.Question.Description,
		); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
			return
		}
		results = append(results, s)
	}

	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Row iteration error"})
		return
	}
	ctx.JSON(200, gin.H{"submissions": results})
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

	query := `
		SELECT
		s.submission_id, s.question_slug, s.code, s.submitted_at,
			q.title, q.description
		FROM leetcode_submissions s
		JOIN leetcode_questions q ON s.question_slug = q.slug
		LIMIT $1 OFFSET $2`
	rows, err := inits.DB.QueryContext(ctx, query, to, from)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var results []struct {
		Submission models.Leetcode_submissions
		Question   models.Leetcode_Questions
	}
	for rows.Next() {
		var s struct {
			Submission models.Leetcode_submissions
			Question   models.Leetcode_Questions
		}
		if err := rows.Scan(
			&s.Submission.Submission_ID,
			&s.Submission.Question_Slug,
			&s.Submission.Code,
			&s.Submission.Submitted_At,
			&s.Question.Title,
			&s.Question.Description,
		); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
			return
		}
		results = append(results, s)
	}

	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Row iteration error"})
		return
	}
	ctx.JSON(200, gin.H{"submissions": results})
}

// UpsertTags will insert the tags if it doesn;t exists,
// if exists it will update the tags
func UpsertTags(ctx *gin.Context) {
	var quesTag models.Question_Tags
	if err := ctx.ShouldBindJSON(&quesTag); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON payload", "details": err.Error()})
		return
	}

	_, err := inits.DB.Exec(
		`INSERT INTO question_tags (slug, tags) 
			VALUES ($1, $2)
			ON CONFLICT (slug)
			DO UPDATE SET tags = $2`,
		quesTag.Slug, quesTag.Tags)

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

	_, err := inits.DB.Exec(
		`DELETE FROM question_tags
		WHERE slug = $1`, slug)

	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to delete tags", "details": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"status": "Tags deleted successfully"})
}

// InsertQuestions will upsert the questions into db
func InsertQuestions(ctx *gin.Context) {
	var question models.Leetcode_Questions
	if err := ctx.ShouldBindJSON(&question); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON payload", "details": err.Error()})
		return
	}

	_, err := inits.DB.Exec(
		`INSERT INTO LEETCODE_QUESTIONS (slug, title, description) 
			VALUES ($1, $2, $3)
			ON CONFLICT (slug)
			DO UPDATE SET title = $2, description = $3`,
		question.Slug, question.Title, question.Description)

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
	var question models.Leetcode_Questions
	row := inits.DB.QueryRow("SELECT slug FROM leetcode_questions WHERE slug = $1 LIMIT 1", submission.Question_Slug)
	err := row.Scan(&question.Slug)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(400, gin.H{"error": "Referenced question not found", "slug": submission.Question_Slug})
			return
		}
		ctx.JSON(500, gin.H{"error": "Failed to query question", "details": err.Error()})
		return
	}

	// Create the submission
	_, err = inits.DB.Exec(
		`INSERT INTO LEETCODE_SUBMISSIONS (Submission_ID, Question_Slug, Code,Submitted_At) 
			VALUES ($1, $2, $3, $4)`,
		submission.Submission_ID, submission.Question_Slug, submission.Code, submission.Submitted_At)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to insert submission", "details": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"status": "Submission inserted successfully"})
}
