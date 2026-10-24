package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Yeah114/g79client/account/mpay"
	"github.com/gin-gonic/gin"
)

// GetCookieRequest 换 Cookie 请求
type GetCookieRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	APIKey   string `json:"api_key,omitempty"`
}

// GetCookieResponse 换 Cookie 响应
type GetCookieResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Cookie  string `json:"cookie,omitempty"`
}

// defaultCookieKey 固定的接口密钥，不可通过环境变量覆盖
const defaultCookieKey = "rin_k0BtF3xU2kk742RW5gEZhDkFFvRBAI7r"

func getCookieAuthKey() string {
	return defaultCookieKey
}

var (
	getCookieRateMu     sync.Mutex
	getCookieRateBuckets = map[string][]time.Time{}
	getCookieRateLimit   = 12
	getCookieRateWindow  = time.Minute
)

func clientKeyForRate(c *gin.Context) string {
	if ip := c.ClientIP(); ip != "" {
		return ip
	}
	if c.Request != nil && c.Request.RemoteAddr != "" {
		return c.Request.RemoteAddr
	}
	return "unknown"
}

func allowGetCookieRate(c *gin.Context) bool {
	key := clientKeyForRate(c)
	now := time.Now()

	getCookieRateMu.Lock()
	defer getCookieRateMu.Unlock()

	slice := getCookieRateBuckets[key]
	kept := make([]time.Time, 0, len(slice)+1)
	for _, ts := range slice {
		if now.Sub(ts) < getCookieRateWindow {
			kept = append(kept, ts)
		}
	}

	if len(kept) >= getCookieRateLimit {
		getCookieRateBuckets[key] = kept
		return false
	}

	kept = append(kept, now)
	getCookieRateBuckets[key] = kept
	return true
}

// RegisterGetCookieRoute 注册账号密码 -> Cookie 接口
func RegisterGetCookieRoute(api *gin.RouterGroup) {
	api.POST("/get_cookie", func(c *gin.Context) {
		if !allowGetCookieRate(c) {
			c.JSON(http.StatusTooManyRequests, GetCookieResponse{Success: false, Message: "请求过于频繁，请稍后再试"})
			return
		}

		var req GetCookieRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, GetCookieResponse{Success: false, Message: fmt.Sprintf("请求格式错误: %v", err)})
			return
		}

		authKey := getCookieAuthKey()
		if authKey != "" {
			supplied := req.APIKey
			if supplied == "" {
				supplied = strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
			}
			if supplied != authKey {
				c.JSON(http.StatusOK, GetCookieResponse{Success: false, Message: "api_key 无效"})
				return
			}
		}

		if req.Username == "" || req.Password == "" {
			c.JSON(http.StatusOK, GetCookieResponse{Success: false, Message: "账号或密码不能为空"})
			return
		}

		dev, derr := mpay.GenerateDevice(c.Request.Context())
		if derr != nil {
			c.JSON(http.StatusOK, GetCookieResponse{Success: false, Message: fmt.Sprintf("生成设备信息失败: %v", derr)})
			return
		}

		usr, lerr := dev.LoginEmail(c.Request.Context(), req.Username, req.Password)
		if lerr != nil {
			c.JSON(http.StatusOK, GetCookieResponse{Success: false, Message: fmt.Sprintf("登录失败: %v", lerr)})
			return
		}

		ck, cerr := usr.CookieString()
		if cerr != nil {
			c.JSON(http.StatusOK, GetCookieResponse{Success: false, Message: fmt.Sprintf("生成 Cookie 失败: %v", cerr)})
			return
		}

		fmt.Printf("[get_cookie] 换到 Cookie len=%d user=%s\n", len(ck), req.Username)
		c.JSON(http.StatusOK, GetCookieResponse{Success: true, Message: "ok", Cookie: ck})
	})
}
