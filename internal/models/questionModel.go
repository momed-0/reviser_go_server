package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type StringArray []string

// Implement the driver.Valuer interface for StringArray
func (sa StringArray) Value() (driver.Value, error) {
	return json.Marshal(sa)
}

// Implement the sql.Scanner interface for StringArray
func (sa *StringArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert database value to []byte")
	}
	return json.Unmarshal(bytes, sa)
}

type Question struct {
	gorm.Model
	Slug        string `gorm:"unique"`
	Title       string
	Description string
}

type Question_Tags struct {
	gorm.Model
	Slug     string      `gorm:"unique" json:"slug"`
	Tags     StringArray `gorm:"type:json" json:"tags"`           // Store as JSON in the database
	Question Question    `gorm:"foreignKey:Slug;references:Slug"` // Foreign key relationship
}

type Leetcode_submissions struct {
	gorm.Model
	Submission_ID uint   `gorm:"unique"`
	Slug          string `gorm:"not null"`
	Code          string
	Submitted_At  time.Time
	Question      Question `gorm:"foreignKey:Slug;references:Slug"` // Foreign key relationship
}
