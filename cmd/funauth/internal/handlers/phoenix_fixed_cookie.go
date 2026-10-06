package handlers

// fixedCookie 作为「未提供账号密码且未提供 FBToken」时的兜底。
// 出于安全考虑，默认留空（不回退到任何固定账号凭据）。
// 如确需兜底，可通过环境变量 FUNAUTH_FIXED_COOKIE 注入。
var fixedCookie = ""
