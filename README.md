# rin-server

基于 [FunAuth](https://github.com/Yeah114/FunAuth) 的网易《我的世界》认证服务。

## 功能

- 账号密码换 Cookie (`POST /api/get_cookie`)
- 完整的 Phoenix 登录 / 联机大厅（Tan Lobby）服务

## 主要接口

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/get_cookie` | 账号密码 → 网易 Cookie |
| POST | `/api/phoenix/login` | 完整登录（账号密码或 FBToken）|
| POST | `/api/phoenix/tan_lobby_login` | 联机大厅登录 |

### `POST /api/get_cookie`

请求：
```json
{ "username": "账号", "password": "密码", "api_key": "密钥" }
```

响应：
```json
{ "success": true, "message": "ok", "cookie": "..." }
```

## 环境变量

| 变量 | 说明 |
|---|---|
| `FUNAUTH_ADDR` | 监听地址，默认 `:8080` |
| `FUNAUTH_COOKIE_KEY` | `get_cookie` 接口密钥；为空时不校验 |
| `FUNAUTH_PROXY_API_URL` | 代理池 HTTP API（可选，用于换 IP 规避限流）|
| `HTTPS_PROXY` / `HTTP_PROXY` | 标准代理环境变量（可选）|

## 构建

```bash
go build -o rin-server ./cmd/funauth
```

## 运行

```bash
FUNAUTH_COOKIE_KEY=your-secret ./rin-server
```
