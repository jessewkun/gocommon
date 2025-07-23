package http

import (
	"net/http"
	"strings"
	"time"
)

// SetCookie go 官方的 setcookie 会默认去掉 domain 前边的 .，可能有一些浏览器的请求无法携带 cookie
func SetCookie(response http.ResponseWriter, name string, value string, maxAge time.Duration, path string, domain string, secure bool, httpOnly bool) {
	var cookieStr = (&http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().Add(maxAge),
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode, // 允许跨域请求携带cookie
	}).String()
	if len(domain) > 0 && domain[0] == '.' {
		cookieStr = strings.Replace(cookieStr, "; Domain=", "; Domain=.", 1)
	}
	response.Header().Add("Set-Cookie", cookieStr)
}
