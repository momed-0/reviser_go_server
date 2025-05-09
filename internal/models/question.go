package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
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

type Leetcode_Questions struct {
	Slug        string
	Title       string
	Description string
}

type Question_Tags struct {
	Slug string
	Tags StringArray
}

type Leetcode_submissions struct {
	Submission_ID uint
	Question_Slug string
	Code          string
	Submitted_At  time.Time
}
