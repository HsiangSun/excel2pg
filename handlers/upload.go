package handlers

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"excel_upload_project/db"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func UploadFile(c *gin.Context) {
	table := c.PostForm("table")
	if table == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Table name is required", "success": false})
		return
	}

	file, _ := c.FormFile("file")
	fileContent, _ := file.Open()
	defer fileContent.Close()

	// Calculate MD5 hash
	hasher := md5.New()
	_, err := io.Copy(hasher, fileContent)
	if err != nil {
		logrus.Errorf("Failed to read file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upload file", "success": false})
		return
	}
	md5Str := hex.EncodeToString(hasher.Sum(nil))

	// Check if the file already exists by comparing MD5
	var status string
	err = db.DB.Get(&status, "SELECT status FROM uploaded_files WHERE md5 = $1", md5Str)
	if err != nil {
		if !strings.Contains(err.Error(), "no rows in result set") {
			logrus.Errorf("get uploaded_files status error, err: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "db error", "success": false})
			return
		}
	}

	if status == "Success" {
		c.JSON(http.StatusConflict, gin.H{"message": "Duplicate file upload", "success": false})
		return
	} else if status == "In Progress" {
		c.JSON(http.StatusConflict, gin.H{"message": "File upload is already in progress", "success": false})
		return
	} //if failed pass

	// Save file to disk
	savePath := fmt.Sprintf("./uploads/%s", file.Filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		logrus.Errorf("Failed to save file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upload file", "success": false})
		return
	}

	// Create a new task in the database
	var taskID int
	err = db.DB.QueryRow("INSERT INTO tasks (file_name, table_name, status) VALUES ($1, $2, $3) RETURNING id", file.Filename, table, "Pending").Scan(&taskID)
	if err != nil {
		logrus.Errorf("Failed to create task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create processing task"})
		return
	}

	// Asynchronous processing of file data
	go processFileData(savePath, md5Str, table, taskID, status)

	logrus.Infof("File %s uploaded successfully by user, target table: %s", file.Filename, table)
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "task_id": taskID, "success": true})
}

func getColumnTypes(table string) (map[string]string, error) {
	columnTypes := make(map[string]string)
	rows, err := db.DB.Query("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = $1", table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName, dataType string
		if err := rows.Scan(&columnName, &dataType); err != nil {
			return nil, err
		}
		columnTypes[columnName] = dataType
		fmt.Printf("%s=>%s \n", columnName, dataType)
	}

	return columnTypes, nil
}

//func processFileData(filePath, md5Str, table string, taskID int, status string) {
//	var upload_id int
//	if status == "" {
//		err := db.DB.QueryRow("INSERT INTO uploaded_files (filename, md5, status) VALUES ($1, $2, $3) RETURNING id", filePath, md5Str, "In Progress").Scan(&upload_id)
//		if err != nil {
//			logrus.Errorf("Failed to create uploaded_files: %v", err)
//			return
//		}
//	}
//
//	columnTypes, err := getColumnTypes(table)
//	if err != nil {
//		updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to retrieve column types for table %s: %v", table, err))
//		return
//	}
//
//	rowSql := fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE table_name = '%s' ORDER BY ordinal_position", table)
//	var columns []string
//	sqlRows, err := db.DB.Query(rowSql)
//	if err != nil {
//		updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to retrieve column configuration for table %s: %v", table, err))
//		return
//	}
//	defer sqlRows.Close()
//
//	for sqlRows.Next() {
//		var columnName string
//		if err := sqlRows.Scan(&columnName); err != nil {
//			updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to scan column name for table %s: %v", table, err))
//			return
//		}
//		columns = append(columns, columnName)
//	}
//
//	if len(columns) == 0 {
//		updateTaskStatus(taskID, "Failed", fmt.Sprintf("No columns configured for table %s", table))
//		return
//	}
//
//	_, err = db.DB.Exec("UPDATE tasks SET status = $1 WHERE id = $2", "Processing", taskID)
//	if err != nil {
//		logrus.Errorf("Failed to update task status to Processing: %v", err)
//		return
//	}
//
//	f, err := excelize.OpenFile(filePath)
//	if err != nil {
//		updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to open file: %v", err))
//		return
//	}
//
//	sheetMap := f.GetSheetMap()
//	logrus.Infof("Processing sheets: %+v", sheetMap)
//
//	for _, sheetName := range sheetMap {
//		logrus.Infof("Processing sheet: %s", sheetName)
//
//		rows, err := f.Rows(sheetName)
//		if err != nil {
//			logrus.Errorf("Failed to get rows from sheet %s: %v", sheetName, err)
//			return
//		}
//
//		var rowIndex int
//		for rows.Next() {
//			column := rows.Columns()
//
//			rowIndex++
//			if rowIndex == 1 {
//				continue // 跳过表头行
//			}
//
//			values := make([]interface{}, len(columns))
//			for j := range columns {
//				if j < len(column) {
//					val := column[j]
//					dataType := columnTypes[columns[j]]
//
//					if dataType == "integer" || dataType == "numeric" || dataType == "bigint" || dataType == "double precision" {
//						if val != "" {
//							values[j] = val
//						} else {
//							values[j] = nil // 数值类型的空值用 SQL 的 NULL 表示
//						}
//					} else {
//						values[j] = val
//					}
//				} else {
//					values[j] = sql.NullString{} // 填充多余的列为 NULL
//				}
//			}
//
//			placeholders := make([]string, len(columns))
//			for k := range placeholders {
//				placeholders[k] = fmt.Sprintf("$%d", k+1)
//			}
//			placeholderString := strings.Join(placeholders, ", ")
//
//			query := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`, table, strings.Join(columns, ", "), placeholderString)
//
//			_, err = db.DB.Exec(query, values...)
//			if err != nil {
//				logrus.Error(err)
//				_, _ = db.DB.Exec("UPDATE uploaded_files SET status = 'Failed' WHERE id = $1", upload_id) // 更新状态为 Failed
//				return
//			}
//		}
//
//		logrus.Infof("Successfully processed sheet: %s", sheetName)
//
//		// 显式清理
//		rows = nil
//		runtime.GC() // 触发 GC 回收内存
//	}
//
//	err = os.Remove(filePath)
//	if err != nil {
//		logrus.Printf("Failed to remove file %s: %v", filePath, err)
//	}
//
//	logrus.Infof("All sheets processed successfully for file: %s", filePath)
//
//	_, _ = db.DB.Exec("UPDATE uploaded_files SET status = 'Success' WHERE id = $1", upload_id)
//
//	updateTaskStatus(taskID, "Success", "")
//	runtime.GC() // 最后再手动触发一次 GC
//}

// Convert Excel date (serial number) to time.Time
func excelDateToTime(excelDate float64) (time.Time, error) {
	// Excel has a bug: 1900 is treated as a leap year, so we need to subtract one day from dates before March 1, 1900.
	const excelEpoch = "1899-12-30"
	const secondsInDay = 24 * 60 * 60

	epoch, err := time.Parse("2006-01-02", excelEpoch)
	if err != nil {
		return time.Time{}, err
	}

	duration := time.Duration(excelDate * secondsInDay * 1e9) // convert days to nanoseconds
	return epoch.Add(duration), nil
}

// 判断行是否为空
func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false // 行中有内容，不为空行
		}
	}
	return true // 行中所有单元格都为空
}

func processFileData(filePath, md5Str, table string, taskID int, status string) {
	var upload_id int
	if status == "" {
		err := db.DB.QueryRow("INSERT INTO uploaded_files (filename, md5, status) VALUES ($1, $2, $3) RETURNING id", filePath, md5Str, "In Progress").Scan(&upload_id)
		if err != nil {
			logrus.Errorf("Failed to create uploaded_files: %v", err)
			return
		}
	}

	columnTypes, err := getColumnTypes(table)
	if err != nil {
		updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to retrieve column types for table %s: %v", table, err))
		return
	}

	//rowSql := fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE table_name = '%s' ORDER BY ordinal_position", table)
	rowSql := fmt.Sprintf("SELECT column_name FROM table_config WHERE table_name = '%s' ORDER BY column_order", table)
	var columns []string
	sqlRows, err := db.DB.Query(rowSql)
	if err != nil {
		updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to retrieve column configuration for table %s: %v", table, err))
		return
	}
	defer sqlRows.Close()

	for sqlRows.Next() {
		var columnName string
		if err := sqlRows.Scan(&columnName); err != nil {
			updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to scan column name for table %s: %v", table, err))
			return
		}

		//only insert configured filed
		lowerColumnName := strings.ToLower(columnName)
		if _, ok := columnTypes[lowerColumnName]; ok {
			columns = append(columns, fmt.Sprintf(`%s`, lowerColumnName))
		}

	}

	if len(columns) == 0 {
		updateTaskStatus(taskID, "Failed", fmt.Sprintf("No columns configured for table %s", table))
		return
	}

	_, err = db.DB.Exec("UPDATE tasks SET status = $1 WHERE id = $2", "Processing", taskID)
	if err != nil {
		logrus.Errorf("Failed to update task status to Processing: %v", err)
		return
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to open file: %v", err))
		return
	}

	sheetMap := f.GetSheetMap()
	logrus.Infof("Processing sheets: %+v", sheetMap)

	for _, sheetName := range sheetMap {
		logrus.Infof("Processing sheet: %s", sheetName)

		rows, err := f.Rows(sheetName)
		if err != nil {
			logrus.Errorf("Failed to get rows from sheet %s: %v", sheetName, err)
			return
		}

		tx, err := db.DB.Beginx()
		if err != nil {
			logrus.Errorf("db begin transaction error:%s", err.Error())
			return
		}

		var rowIndex int
		for rows.Next() {
			column := rows.Columns()

			if isEmptyRow(column) {
				logrus.Infof("Skipping empty row: %d", rowIndex)
				continue
			}

			rowIndex++
			if rowIndex == 1 {
				continue // 跳过表头行
			}

			values := make([]interface{}, len(columns))
			for j := range columns {
				if j < len(column) {
					val := column[j]
					dataType := columnTypes[columns[j]]

					if dataType == "integer" || dataType == "numeric" || dataType == "bigint" || dataType == "double precision" {
						if val != "" {
							values[j] = val
						} else {
							values[j] = nil // 数值类型的空值用 SQL 的 NULL 表示
						}
					} else if dataType == "timestamp without time zone" {
						floatValue, err := strconv.ParseFloat(val, 64)
						if err == nil {
							timeValue, err := excelDateToTime(floatValue)
							if err != nil {
								log.Fatalf("Failed to convert Excel date to time: %v", err)
							}
							timeVal := timeValue.Format("2006-01-02 15:04:05")
							values[j] = timeVal
						} else {
							//logrus.Errorf("Cell value is not a date sequence:%s", val)
							//excelDateToTime is text
							values[j] = val
						}
					} else {
						values[j] = val
					}
				} else {
					values[j] = sql.NullString{} // 填充多余的列为 NULL
				}
			}

			placeholders := make([]string, len(columns))
			for k := range placeholders {
				placeholders[k] = fmt.Sprintf("$%d", k+1)
			}
			placeholderString := strings.Join(placeholders, ", ")

			columnsWithQuo := make([]string, len(columns))
			for i, col := range columns {
				columnsWithQuo[i] = fmt.Sprintf(`"%s"`, col)
			}

			sqlValues := strings.Join(columnsWithQuo, ", ")

			query := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`, table, sqlValues, placeholderString)

			_, err = tx.Exec(query, values...)
			if err != nil {
				errMsg := err.Error()

				if err, ok := err.(*pq.Error); ok {
					where := err.Where
					errMsg += ">>>" + where
				}

				logrus.Error("%+v", err)
				logrus.Errorf("query:%s \nvalues:%+v", query, values)
				tx.Rollback()                                                                           // 发生错误时回滚事务
				_, _ = db.DB.Exec("UPDATE uploaded_files SET status = 'Failed' WHERE md5 = $1", md5Str) // 更新状态为 Failed
				updateTaskStatus(taskID, "Failed", errMsg)
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback() // 提交失败时回滚事务
			logrus.Errorf("transaction commit error:%s", err.Error())
			_, _ = db.DB.Exec("UPDATE uploaded_files SET status = 'Failed' WHERE md5 = $1", md5Str) // 更新状态为 Failed
			return
		}

		logrus.Infof("Successfully processed sheet: %s", sheetName)

		// 显式清理
		rows = nil
		runtime.GC() // 触发 GC 回收内存
	}

	err = os.Remove(filePath)
	if err != nil {
		logrus.Printf("Failed to remove file %s: %v", filePath, err)
	}

	logrus.Infof("All sheets processed successfully for file: %s", filePath)

	_, _ = db.DB.Exec("UPDATE uploaded_files SET status = 'Success' WHERE md5 = $1", md5Str)

	updateTaskStatus(taskID, "Success", "")
	runtime.GC() // 最后再手动触发一次 GC
}

//func processFileData(filePath, md5Str, table string, taskID int) {
//	// 记录 MD5 哈希以防止重新上传相同的文件
//	_, err := db.DB.Exec("INSERT INTO uploaded_files (filename, md5) VALUES ($1, $2)", filePath, md5Str)
//	if err != nil {
//		updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to insert file record into DB: %v", err))
//		return
//	}
//
//	// 获取列的数据类型
//	columnTypes, err := getColumnTypes(table)
//	if err != nil {
//		updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to retrieve column types for table %s: %v", table, err))
//		return
//	}
//
//	// 从数据库中获取表格配置的列
//	var columns []string
//	rows, err := db.DB.Query("SELECT column_name FROM information_schema.columns WHERE table_name = $1 ORDER BY ordinal_position", table)
//	if err != nil {
//		updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to retrieve column configuration for table %s: %v", table, err))
//		return
//	}
//	defer rows.Close()
//
//	for rows.Next() {
//		var columnName string
//		if err := rows.Scan(&columnName); err != nil {
//			updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to scan column name for table %s: %v", table, err))
//			return
//		}
//		columns = append(columns, columnName)
//	}
//
//	if len(columns) == 0 {
//		updateTaskStatus(taskID, "Failed", fmt.Sprintf("No columns configured for table %s", table))
//		return
//	}
//
//	// 更新任务状态为 "Processing"
//	_, err = db.DB.Exec("UPDATE tasks SET status = $1 WHERE id = $2", "Processing", taskID)
//	if err != nil {
//		logrus.Errorf("Failed to update task status to Processing: %v", err)
//		return
//	}
//
//	f, err := excelize.OpenFile(filePath)
//	if err != nil {
//		updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to open file: %v", err))
//		return
//	}
//
//	start := time.Now()
//	excelRows := f.GetRows(f.GetSheetName(1))
//	end := time.Now()
//
//	cost := end.Sub(start).Seconds()
//	logrus.Infof("read excel file cost %f s", cost)
//
//	// 开启事务
//	tx, err := db.DB.Beginx()
//	if err != nil {
//		logrus.Errorf("db begin transaction error:%s", err.Error())
//		return
//	}
//
//	// 遍历行，跳过表头行
//	for i, row := range excelRows {
//		if i == 0 {
//			continue
//		}
//
//		// 填充不足列为 NULL
//		values := make([]interface{}, len(columns))
//		for j := range columns {
//			if j < len(row) {
//				val := row[j]
//				dataType := columnTypes[columns[j]]
//
//				if dataType == "integer" || dataType == "numeric" || dataType == "bigint" || dataType == "double precision" {
//					if val != "" {
//						values[j] = val
//					} else {
//						values[j] = nil // 数值类型的空值用 SQL 的 NULL 表示
//					}
//				} else {
//					values[j] = val
//				}
//			} else {
//				values[j] = sql.NullString{} // 填充多余的列为 NULL
//			}
//		}
//
//		placeholders := make([]string, len(columns))
//		for k := range placeholders {
//			placeholders[k] = fmt.Sprintf("$%d", k+1)
//		}
//		placeholderString := strings.Join(placeholders, ", ")
//
//		// 构建插入语句
//		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ", "), placeholderString)
//
//		// 打印可执行的 SQL 语句
//		//logrus.Infof("Executing SQL: %s with values: %v", query, values)
//
//		// 执行插入查询
//		_, err = tx.Exec(query, values...)
//		if err != nil {
//			logrus.Error(err)
//			updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to insert data into table %s: %v", table, err))
//			tx.Rollback() // 发生错误时回滚事务
//			return
//		}
//	}
//
//	// 处理完文件后删除
//	err = os.Remove(filePath)
//	if err != nil {
//		logrus.Printf("Failed to remove file %s: %v", filePath, err)
//	}
//
//	// 提交事务
//	err = tx.Commit()
//	if err != nil {
//		tx.Rollback() // 提交失败时回滚事务
//		logrus.Errorf("transaction commit error:%s", err.Error())
//		updateTaskStatus(taskID, "Failed", fmt.Sprintf("Failed to commit transaction: %v", err))
//		return
//	}
//
//	logrus.Infof("Insert data successfully with file:%s", filePath)
//
//	// 更新任务状态为 "Success"
//	updateTaskStatus(taskID, "Success", "")
//}

func updateTaskStatus(taskID int, status, errorMessage string) {
	_, err := db.DB.Exec("UPDATE tasks SET status = $1, error_message = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3", status, errorMessage, taskID)
	if err != nil {
		logrus.Errorf("Failed to update task status to %s: %v", status, err)
	}
}
