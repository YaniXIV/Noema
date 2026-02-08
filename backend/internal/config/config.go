package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const devCookieSecret = "dev-secret-change-in-production"

// Load reads .env from the current directory and sets env vars.
// Safe to call multiple times; existing env vars are not overwritten.
func Load() error {
	return godotenv.Load()
}

// JudgeKey returns the judge key used to gate protected routes.
func JudgeKey() string {
	return os.Getenv("JUDGE_KEY")
}

// GeminiAPIKey returns the Google Gemini API key.
func GeminiAPIKey() string {
	return os.Getenv("GEMINI_API_KEY")
}

// CookieSecret returns the secret for signing session cookies (NOEMA_COOKIE_SECRET).
// If unset, returns a dev default and callers should log a warning.
func CookieSecret() string {
	s := os.Getenv("NOEMA_COOKIE_SECRET")
	if s == "" {
		return devCookieSecret
	}
	return s
}

// SecureCookies returns true if cookies should use the Secure flag (e.g. behind HTTPS).
func SecureCookies() bool {
	return os.Getenv("NOEMA_SECURE_COOKIES") == "1" || os.Getenv("HTTPS") == "1"
}

// UploadsDir returns the directory for uploaded files.
func UploadsDir() string {
	if v := os.Getenv("NOEMA_UPLOADS_DIR"); v != "" {
		return v
	}
	return "data/uploads"
}

// RunsDir returns the directory for evaluation runs.
func RunsDir() string {
	if v := os.Getenv("NOEMA_RUNS_DIR"); v != "" {
		return v
	}
	return "data/runs"
}

// SampleItemsLimit returns the max number of dataset items sent to Gemini.
func SampleItemsLimit() int {
	if v := os.Getenv("NOEMA_SAMPLE_ITEMS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return 100
}

// RunsIndexLimit returns the max number of runs kept in index.json.
func RunsIndexLimit() int {
	if v := os.Getenv("NOEMA_RUNS_INDEX_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return 50
}

// RunsMax returns the maximum number of run artifacts to retain.
// If unset or invalid, defaults to 50. Set to 0 to disable pruning.
func RunsMax() int {
	if v := os.Getenv("NOEMA_RUNS_MAX"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return 50
}
