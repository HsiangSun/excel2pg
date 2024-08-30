package db

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var DB *sqlx.DB

func InitDB() {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// 从环境变量中读取数据库连接信息
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")
	host := os.Getenv("DB_HOST")

	// 构建连接字符串
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s port=%s sslmode=%s host=%s",
		user, password, dbname, port, sslmode, host)

	// 连接到数据库
	DB, err = sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
}
