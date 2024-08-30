package models

import (
	"time"
)

type Task struct {
	ID           int       `json:"id"`
	ErrorMessage string    `json:"error_message"`
	FileName     string    `json:"file_name"`
	Status       string    `json:"status"`
	TableName    string    `json:"table_name"`
	Elapsed      float64   `json:"elapsed_time"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
