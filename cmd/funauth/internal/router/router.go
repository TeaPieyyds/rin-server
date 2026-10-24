package router

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Yeah114/FunAuth/cmd/funauth/internal/handlers"
	"github.com/gin-gonic/gin"
)

// corsMiddleware 允许 GitHub Pages 等静态站点跨域调用 API
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := false

		if origin == "" {
			// 非浏览器请求（curl 等）
			allowed = true
		} else if strings.HasSuffix(origin, ".github.io") {
			allowed = true
		} else if extra := strings.TrimSpace(os.Getenv("FUNAUTH_CORS_ORIGINS")); extra != "" {
			for _, o := range strings.Split(extra, ",") {
				if strings.TrimSpace(o) == origin {
					allowed = true
					break
				}
			}
		}

		if allowed && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func NewRouter() *gin.Engine {
	// 确保 gin 的默认日志写到 stdout
	gin.DefaultWriter = io.MultiWriter(os.Stdout)
	gin.DefaultErrorWriter = io.MultiWriter(os.Stdout)
	log.SetOutput(os.Stdout)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), corsMiddleware())
	_ = r.SetTrustedProxies([]string{"127.0.0.1"})

	// API：账号密码 -> Cookie
	api := r.Group("/api")
	handlers.RegisterGetCookieRoute(api)

	return r
}
