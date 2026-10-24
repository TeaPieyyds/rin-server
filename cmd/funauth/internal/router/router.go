package router

import (
	"io"
	"log"
	"os"

	"github.com/Yeah114/FunAuth/cmd/funauth/internal/handlers"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	// 确保 gin 的默认日志写到 stdout
	gin.DefaultWriter = io.MultiWriter(os.Stdout)
	gin.DefaultErrorWriter = io.MultiWriter(os.Stdout)
	log.SetOutput(os.Stdout)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	_ = r.SetTrustedProxies([]string{"127.0.0.1"})

	// API：账号密码 -> Cookie
	api := r.Group("/api")
	handlers.RegisterGetCookieRoute(api)

	return r
}
