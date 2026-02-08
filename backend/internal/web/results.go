package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResultsData is passed to the results template.
type ResultsData struct {
	RunID string
}

// ResultsPage renders GET /app/results/:id.
func ResultsPage(c *gin.Context, tmpl string, runID string) {
	c.HTML(http.StatusOK, tmpl, ResultsData{RunID: runID})
}
