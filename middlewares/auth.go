package middlewares

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthRequired(c *gin.Context) {
	// 这里假设使用 session 来管理用户登录状态
	session := sessions.Default(c)
	user := session.Get("user")

	// 如果用户未登录，重定向到登录页面
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}

	// 用户已登录，继续处理请求
	c.Next()
}
