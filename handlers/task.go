package handlers

import (
	"excel_upload_project/db"
	"excel_upload_project/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GetTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")

	var is_id = true

	_, err := strconv.Atoi(taskID)
	if err != nil {
		is_id = false
	}

	var id int
	var created_at, updated_at time.Time
	var file_name, table_name, status, error_message string

	if is_id {
		err = db.DB.QueryRow("SELECT id,COALESCE(file_name, '') as file_name,COALESCE(table_name, '') as table_name,COALESCE(status, '') AS status,COALESCE(error_message, '') AS error_message,created_at,updated_at FROM tasks WHERE id = $1", taskID).Scan(&id, &file_name, &table_name, &status, &error_message, &created_at, &updated_at)
	} else {
		err = db.DB.QueryRow("SELECT id,COALESCE(file_name, '') as file_name,COALESCE(table_name, '') as table_name,COALESCE(status, '') AS status,COALESCE(error_message, '') AS error_message,created_at,updated_at FROM tasks WHERE file_name = $1", taskID).Scan(&id, &file_name, &table_name, &status, &error_message, &created_at, &updated_at)
	}

	if err != nil {

		if strings.Contains(err.Error(), "no rows in result set") {
			c.JSON(http.StatusNoContent, gin.H{})
		}
		return
	}

	// 计算时间差
	duration := updated_at.Sub(created_at)

	// 将时间差转换为秒数
	seconds := duration.Seconds()

	task := models.Task{
		ID:           id,
		FileName:     file_name,
		TableName:    table_name,
		Status:       status,
		ErrorMessage: error_message,
		Elapsed:      seconds,
		CreatedAt:    created_at,
		UpdatedAt:    updated_at,
	}

	c.JSON(http.StatusOK, task)
}

func ListRecentTask(c *gin.Context) {

	rows, err := db.DB.Query("SELECT id,COALESCE(file_name, '') as file_name,COALESCE(table_name, '') as table_name,COALESCE(status, '') AS status,COALESCE(error_message, '') AS error_message,created_at,updated_at FROM tasks ORDER BY id DESC LIMIT 10")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve task recent"})
		return
	}

	defer rows.Close()

	var results []models.Task

	for rows.Next() {
		var id int
		var created_at, updated_at time.Time
		var file_name, table_name, status, error_message string
		if err := rows.Scan(&id, &file_name, &table_name, &status, &error_message, &created_at, &updated_at); err != nil {
			logrus.Error(err)
			return
		}

		// 计算时间差
		duration := updated_at.Sub(created_at)

		// 将时间差转换为秒数
		seconds := duration.Seconds()

		task := models.Task{
			ID:           id,
			FileName:     file_name,
			TableName:    table_name,
			Status:       status,
			ErrorMessage: error_message,
			Elapsed:      seconds,
			CreatedAt:    created_at,
			UpdatedAt:    updated_at,
		}

		results = append(results, task)
	}

	c.JSON(http.StatusOK, results)
}
