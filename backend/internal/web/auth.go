package web

import (
	"net/http"

	"noema/internal/config"
	"noema/internal/session"

	"github.com/gin-gonic/gin"
)

// IndexData is passed to the landing template.
type IndexData struct {
	Error        string
	JudgeKeyVal  string
	ServerNotCfg bool
}

// Index renders the landing page (GET /).
func Index(c *gin.Context, tmpl string, data IndexData) {
	c.HTML(http.StatusOK, tmpl, data)
}

// AuthPost handles POST /auth (form field "judge_key").
// On success: set cookie, redirect 303 to /app. On failure: re-render index with inline error.
func AuthPost(c *gin.Context, indexTmpl string) {
	key := c.PostForm("judge_key")
	expect := config.JudgeKey()
	if expect == "" {
		Index(c, indexTmpl, IndexData{Error: "Server not configured.", ServerNotCfg: true})
		return
	}
	if key == "" {
		Index(c, indexTmpl, IndexData{Error: "Please enter a judge key.", JudgeKeyVal: key})
		return
	}
	if key != expect {
		Index(c, indexTmpl, IndexData{Error: "Invalid judge key.", JudgeKeyVal: key})
		return
	}
	signed := session.Sign(config.CookieSecret(), expect)
	opts := session.CookieOptions{Secure: config.SecureCookies()}
	c.Header("Set-Cookie", session.SetCookieHeader(signed, opts))
	c.Redirect(http.StatusSeeOther, "/app/new")
}

// Logout clears the session cookie and redirects to /.
func Logout(c *gin.Context) {
	c.SetCookie(session.CookieName, "", -1, "/", "", config.SecureCookies(), true)
	c.Redirect(http.StatusSeeOther, "/")
}
