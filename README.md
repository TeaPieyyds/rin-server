# rin-server

网易《我的世界》账号密码 → Cookie 转换服务。

基于 [FunAuth](https://github.com/Yeah114/FunAuth) 精简而来，**只保留换 Cookie 一个功能**。

## 在线网页

- 页面：https://teapieyyds.github.io/rin-server/
- 接口：https://rin-server.orkestr.run/api/get_cookie

## 功能

- 浏览器打开网页，输入网易通行证账号密码，直接拿到登录 Cookie
- 提供 REST API：`POST /api/get_cookie`

## 使用

### 网页版

打开在线页面，填写账号密码即可。

### API

```bash
curl -X POST https://rin-server.orkestr.run/api/get_cookie   -H 'Content-Type: application/json'   -d '{"username":"账号@163.com","password":"密码","api_key":"密钥"}'
```

响应：

```json
{ "success": true, "message": "ok", "cookie": "{"sauth_json":...}" }
```

失败时：

```json
{ "success": false, "message": "登录失败: ..." }
```

## 环境变量

| 变量 | 说明 |
|---|---|
| `FUNAUTH_ADDR` | 监听地址，默认 `:8080` |
| `FUNAUTH_COOKIE_KEY` | `/api/get_cookie` 接口密钥；为空时不校验（**公网部署务必设置**） |

## 构建

```bash
go build -o rin-server ./cmd/funauth
```

## 运行

```bash
# 建议设置密钥
FUNAUTH_COOKIE_KEY=your-secret ./rin-server
```

## 关于 GitHub Actions

- GitHub Actions 只能做构建、测试、发布静态资源，不能替代常驻服务。
- 在线主页放 GitHub Pages；Cookie API 放 Orkestr。

## 注意

- 在线页面为静态托管；真正的换 Cookie 接口运行在 Orkestr
- GitHub Actions 不能替代常驻服务，只能做构建/测试/发布静态资源
- Cookie 等同登录凭证，请勿泄露
- 服务端不存储任何账号密码
- 请遵守网易用户协议，仅用于个人学习研究
