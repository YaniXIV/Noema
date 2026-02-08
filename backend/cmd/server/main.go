package main

import (
	"context"
	"log"
	"os"
	"strings"

	"noema/internal/auth"
	"noema/internal/config"
	"noema/internal/evaluate"
	"noema/internal/gemini"
	"noema/internal/verify"
	"noema/internal/web"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load(); err != nil {
		log.Println("no .env loaded:", err)
	}
	if config.JudgeKey() == "" {
		log.Println("warning: JUDGE_KEY not set; gated routes will reject all requests")
	}
	if os.Getenv("NOEMA_COOKIE_SECRET") == "" {
		log.Println("warning: NOEMA_COOKIE_SECRET not set; using dev default")
	}

	// Paths relative to working directory — run from backend/
	r := gin.Default()
	r.MaxMultipartMemory = config.MaxMultipartMemory
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "web/static")

	// ----- Public JSON API (unchanged) -----
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/ready", func(c *gin.Context) {
		if err := os.MkdirAll(config.UploadsDir(), 0o755); err != nil {
			c.JSON(500, gin.H{"status": "error", "error": "uploads dir not writable"})
			return
		}
		if err := os.MkdirAll(config.RunsDir(), 0o755); err != nil {
			c.JSON(500, gin.H{"status": "error", "error": "runs dir not writable"})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ----- Public HTML -----
	r.GET("/", func(c *gin.Context) {
		errMsg := c.Query("err")
		web.Index(c, "index", web.IndexData{Error: errMsg})
	})
	r.GET("/verify", func(c *gin.Context) {
		c.HTML(200, "verify", nil)
	})
	r.GET("/verify/:id", func(c *gin.Context) {
		web.ResultsPage(c, "verify_results", c.Param("id"))
	})
	r.POST("/auth", func(c *gin.Context) {
		web.AuthPost(c, "index")
	})
	r.GET("/logout", web.Logout)

	// ----- Cookie-gated HTML -----
	htmlGated := r.Group("/")
	htmlGated.Use(auth.CookieAuth())
	{
		htmlGated.GET("/app", func(c *gin.Context) {
			c.HTML(200, "app", nil)
		})
		htmlGated.GET("/app/new", func(c *gin.Context) {
			c.HTML(200, "app_new", nil)
		})
		htmlGated.GET("/app/results/:id", func(c *gin.Context) {
			web.ResultsPage(c, "app_results", c.Param("id"))
		})
		htmlGated.GET("/upload", func(c *gin.Context) {
			web.UploadGet(c, "upload", web.UploadData{})
		})
		htmlGated.POST("/upload", func(c *gin.Context) {
			web.UploadPost(c, "upload", config.UploadsDir())
		})
	}

	// ----- API gated by cookie (browser session) -----
	apiCookie := r.Group("/api")
	apiCookie.Use(auth.CookieAuth())
	{
		apiCookie.POST("/evaluate", evaluate.Handler(config.RunsDir(), config.RunsMax()))
	}

	// ----- Public verify API -----
	r.POST("/api/verify", verify.Handler())

	// ----- API gated by JudgeKey (X-Judge-Key or judge_key query) — unchanged -----
	apiGated := r.Group("/")
	apiGated.Use(auth.JudgeKey())
	{
		apiGated.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})
	}

	// Optional: run a one-off Gemini test if GEMINI_TEST=1
	if os.Getenv("GEMINI_TEST") == "1" {
		ctx := context.Background()
		_, err := gemini.SendText(ctx, "Say hello in one sentence.")
		if err != nil {
			log.Println("gemini test:", err)
		}
	}

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		if strings.HasPrefix(port, ":") {
			addr = port
		} else {
			addr = ":" + port
		}
	}

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
