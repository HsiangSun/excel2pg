package main

import (
	"excel_upload_project/db"
	"excel_upload_project/handlers"
	"excel_upload_project/middlewares"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	handlers.InitLogger()
	db.InitDB()

	logrus.SetReportCaller(true)

	// 设置日志格式，包含行号和文件名
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logrus.Infof("===========START=============")

	router := gin.Default()

	router.MaxMultipartMemory = 500 << 20 //500MB

	// 设置 session 存储引擎，这里使用 cookie 存储
	store := cookie.NewStore([]byte("secret"))

	store.Options(sessions.Options{
		MaxAge:   int(30 * time.Minute / time.Second), // 设置过期时间为30分钟
		Path:     "/",
		HttpOnly: true, // 确保 cookie 只能通过 HTTP 传输，增加安全性
	})

	router.Use(sessions.Sessions("my_session", store))

	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/login")
	})

	// 登录路由，无需鉴权
	router.GET("/login", func(c *gin.Context) {
		c.File("./static/login.html")
	})
	router.GET("/tailwind.css", func(c *gin.Context) {
		c.File("./static/tailwind.css")
	})
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./static/favicon.ico")
	})

	router.POST("/login", handlers.Login)
	router.POST("/logout", handlers.Logout)

	// 需要鉴权的路由组
	authorized := router.Group("/")
	authorized.Use(middlewares.AuthRequired)
	{
		// table_config
		authorized.GET("/api/table_config", handlers.GetTableConfigs)
		authorized.GET("/api/table_config/:table_name", handlers.GetTableConfigsByTableName)
		authorized.POST("/api/table_config", handlers.AddTableConfig)
		authorized.DELETE("/api/table_config/:id", handlers.DeleteTableConfig)

		// task
		authorized.GET("/api/task/recent", handlers.ListRecentTask)
		authorized.GET("/api/task/:task_id/status", handlers.GetTaskStatus)

		// 文件上传
		authorized.POST("/upload", handlers.UploadFile)

		//user
		authorized.GET("/api/userinfo", handlers.Userinfo)

		// 受保护的静态页面
		authorized.GET("/home", func(c *gin.Context) {
			c.File("./static/home.html")
		})
	}

	router.Run(":7777")
}
