// handlers/config.go
package handlers

import (
	"excel_upload_project/db"
	"excel_upload_project/models"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"net/http"
)

func GetTableConfigsByTableName(c *gin.Context) {
	table_name := c.Param("table_name")
	if table_name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Table name is required", "success": false})
		return
	}

	rows, err := db.DB.Query("SELECT id,table_name,column_name,column_order FROM table_config WHERE table_name=$1 ORDER BY column_order", table_name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve task recent"})
		return
	}

	defer rows.Close()

	var results []models.Config

	for rows.Next() {
		var id, column_order int
		var table_name, column_name string
		if err := rows.Scan(&id, &table_name, &column_name, &column_order); err != nil {
			logrus.Error(err)
			return
		}

		config := models.Config{
			Id:          id,
			TableName:   table_name,
			ColumnName:  column_name,
			ColumnOrder: column_order,
		}

		results = append(results, config)
	}

	c.JSON(http.StatusOK, results)

}

func GetTableConfigs(c *gin.Context) {
	var configs []struct {
		ID          int    `db:"id" json:"id"`
		TableName   string `db:"table_name" json:"table_name"`
		ColumnName  string `db:"column_name" json:"column_name"`
		ColumnOrder int    `db:"column_order" json:"column_order"`
	}

	err := db.DB.Select(&configs, "SELECT id, table_name, column_name, column_order FROM table_config ORDER BY table_name, column_order")
	if err != nil {
		logrus.Errorf("get table config error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve table configurations"})
		return
	}

	c.JSON(http.StatusOK, configs)
}

func AddTableConfig(c *gin.Context) {
	tableName := c.PostForm("table_name")
	columnName := c.PostForm("column_name")
	columnOrder := c.PostForm("column_order")

	if tableName == "" || columnName == "" || columnOrder == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "All fields are required"})
		return
	}

	_, err := db.DB.Exec("INSERT INTO table_config (table_name, column_name, column_order) VALUES ($1, $2, $3)", tableName, columnName, columnOrder)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // Unique violation error code in PostgreSQL
				c.JSON(http.StatusConflict, gin.H{"message": "Duplicate entry: table_name, column_name, or column_order already exists"})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to add table configuration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Table configuration added successfully"})
}

func DeleteTableConfig(c *gin.Context) {
	id := c.Param("id")

	_, err := db.DB.Exec("DELETE FROM table_config WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete table configuration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Table configuration deleted successfully"})
}
