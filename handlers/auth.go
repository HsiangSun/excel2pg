package handlers

import (
	"database/sql"
	"excel_upload_project/db"
	"excel_upload_project/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

//func Login(c *gin.Context) {
//	username := c.PostForm("username")
//	password := c.PostForm("password")
//
//	// Perform authentication logic (simplified for example)
//	if username == "admin" && password == "admin" {
//		c.SetCookie("session_token", "some_session_token", 3600, "/", "localhost", false, true)
//		log.Infof("User %s logged in successfully", username)
//		c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
//	} else {
//		log.Warnf("Failed login attempt for user %s", username)
//		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
//	}
//}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear() // 清除所有会话数据
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to clear session"})
		return
	}
	// 可以选择返回 JSON 或者重定向到登录页面
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful", "redirect": "/login"})
}

func Userinfo(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not logged in"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"username": user})
}

//func Login(c *gin.Context) {
//	username := c.PostForm("username")
//	password := c.PostForm("password")
//
//	// 假设我们有一个验证用户的函数
//	if authenticateUser(username, password) {
//		session := sessions.Default(c)
//		session.Set("user", username)
//		err := session.Save()
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save session"})
//			return
//		}
//		// 登录成功后重定向到受保护的首页
//		c.JSON(http.StatusOK, gin.H{"message": "Login successful", "redirect": "/home"})
//	} else {
//		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
//	}
//}

func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// 从数据库中获取用户信息
	var user models.User
	err := db.DB.QueryRow("SELECT id, username, password_hash FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid username or password"})
		} else {
			log.Println("Database error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		}
		return
	}

	// 验证密码
	//err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if user.PasswordHash != password {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid username or password"})
		return
	}

	// 登录成功，保存会话
	session := sessions.Default(c)
	session.Set("user", user.Username)
	err = session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "redirect": "/home"})
}

// 模拟的用户验证函数
func authenticateUser(username, password string) bool {
	// 这里应该是你验证用户名和密码的逻辑，比如查询数据库
	return username == "admin" && password == "password" // 仅供示例
}
