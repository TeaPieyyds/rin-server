package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/Yeah114/FunAuth/auth"
	"github.com/Yeah114/g79client/account/mpay"
	"github.com/gin-gonic/gin"
)

func RegisterPhoenixTanLobbyLoginRoute(api *gin.RouterGroup) {
	api.POST("/phoenix/tan_lobby_login", func(c *gin.Context) {
		var req TanLobbyLoginRequest
		bodyBytes, _ := c.GetRawData()
		fmt.Printf("[DBG-BODY] tan_lobby_login raw=%s\n", string(bodyBytes))
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, TanLobbyLoginResponse{Success: false, ErrorInfo: fmt.Sprintf("TanLobbyLogin: 绑定请求体时出现问题, 原因是 %v", err)})
			return
		}
		cookieStr := req.FBToken
		if cookieStr == "" && req.UserName != "" && req.Password != "" {
			fmt.Printf("[DBG] tan_lobby_login: 正在用账号密码换 Cookie user=%s\n", req.UserName)
			dev, derr := mpay.GenerateDevice(c.Request.Context())
			if derr != nil {
				fmt.Printf("[DBG] tan_lobby_login: GenerateDevice 失败: %v\n", derr)
			} else {
				usr, lerr := dev.LoginEmail(c.Request.Context(), req.UserName, req.Password)
				if lerr != nil {
					fmt.Printf("[DBG] tan_lobby_login: LoginEmail 失败: %v\n", lerr)
				} else {
					ck, cerr := usr.CookieString()
					if cerr != nil {
						fmt.Printf("[DBG] tan_lobby_login: CookieString 失败: %v\n", cerr)
					} else {
						cookieStr = ck
						fmt.Printf("[DBG] tan_lobby_login: 已换到 Cookie len=%d\n", len(ck))
					}
				}
			}
		}
		if cookieStr == "" {
			fmt.Printf("[DBG] tan_lobby_login: 未能换到有效 Cookie，使用 fixedCookie（将导致 code:32）\n")
			cookieStr = fixedCookie
		}

		cli, err := auth.NewG79Client(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusOK, TanLobbyLoginResponse{Success: false, ErrorInfo: fmt.Sprintf("TanLobbyLogin: 初始化客户端时出现问题, 原因是 %v", err)})
			return
		}

		if err := cli.G79AuthenticateWithCookie(cookieStr); err != nil {
			c.JSON(http.StatusOK, TanLobbyLoginResponse{Success: false, ErrorInfo: fmt.Sprintf("TanLobbyLogin: 使用 Cookie 认证时出现问题, 原因是 %v", err)})
			return
		}

		loginRes, err := auth.TanLobbyLogin(c.Request.Context(), cli, auth.TanLobbyLoginParams{
			RoomID:   req.RoomID,
			Password: req.RoomPassword,
		})
		if err != nil {
			c.JSON(http.StatusOK, TanLobbyLoginResponse{Success: false, ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", err)})
			return
		}

		enableSkin := true
		var skinInfo SkinInfo
		if enableSkin {
			authSkinInfo, err := auth.GetSkinInfo(cli)
			if err != nil {
				c.JSON(http.StatusOK, TanLobbyLoginResponse{Success: false, ErrorInfo: fmt.Sprintf("TanLobbyLogin: 获取皮肤信息时出现问题, 原因是 %v", err)})
				return
			}
			skinInfo = SkinInfo{
				ItemID:          authSkinInfo.ItemID,
				SkinDownloadURL: authSkinInfo.SkinDownloadURL,
				SkinIsSlim:      authSkinInfo.SkinIsSlim,
			}
		}

		botLevel := 0
		if cli.UserDetail != nil {
			botLevel = int(cli.UserDetail.Level.Int64())
		}
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:                true,
			ErrorInfo:              "",
			UserUniqueID:           loginRes.UserUniqueID,
			UserPlayerName:         loginRes.UserPlayerName,
			BotLevel:               botLevel,
			BotSkin:                skinInfo,
			BotComponent:           loginRes.BotComponent,
			RoomOwnerID:            loginRes.RoomOwnerID,
			RoomModDisplayName:     loginRes.RoomModDisplayName,
			RoomModDownloadURL:     loginRes.RoomModDownloadURL,
			RoomModEncryptKey:      loginRes.RoomModEncryptKey,
			RaknetServerAddress:    loginRes.RaknetServerAddress,
			RaknetRand:             loginRes.RaknetRand,
			RaknetAESRand:          loginRes.RaknetAESRand,
			EncryptKeyBytes:        loginRes.EncryptKeyBytes,
			DecryptKeyBytes:        loginRes.DecryptKeyBytes,
			SignalingServerAddress: loginRes.SignalingServerAddress,
			SignalingSeed:          loginRes.SignalingSeed,
			SignalingTicket:        loginRes.SignalingTicket,
		})
	})
}
