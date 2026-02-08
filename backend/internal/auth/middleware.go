package auth

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"

	"noema/internal/config"
	"noema/internal/session"

	"github.com/gin-gonic/gin"
)

// JudgeKey checks the request for a valid judge key (header or query).
// Use for API routes only. Expects X-Judge-Key header or judge_key query param to match JUDGE_KEY in .env.
func JudgeKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		expect := config.JudgeKey()
		if expect == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "server not configured"})
			return
		}
		got := c.GetHeader("X-Judge-Key")
		if got == "" {
			got = c.Query("judge_key")
		}
		if !constantTimeEqual(got, expect) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid judge key"})
			return
		}
		c.Next()
	}
}

// CookieAuth validates the noema_judge session cookie.
// For HTML routes: redirects to / with error query param if invalid.
// For /api/*: returns 401 JSON if invalid (so browser can handle without redirect).
func CookieAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		expect := config.JudgeKey()
		if expect == "" {
			abortCookieAuth(c, "Server not configured")
			return
		}
		cookie, err := c.Cookie(session.CookieName)
		if err != nil || cookie == "" {
			abortCookieAuth(c, "Session required")
			return
		}
		payload, ok := session.Verify(config.CookieSecret(), cookie)
		if !ok || !constantTimeEqual(payload, expect) {
			abortCookieAuth(c, "Invalid session")
			return
		}
		c.Next()
	}
}

func abortCookieAuth(c *gin.Context, msg string) {
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
		return
	}
	c.Redirect(http.StatusSeeOther, "/?err="+url.QueryEscape(msg))
	c.Abort()
}

func constantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
