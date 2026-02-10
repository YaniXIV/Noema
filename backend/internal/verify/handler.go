package verify

import (
	"net/http"
	"strings"

	"noema/internal/config"
	"noema/internal/httputil"
	"noema/internal/zk"

	"github.com/gin-gonic/gin"
)

// VerifyRequest is the JSON body for POST /api/verify.
type VerifyRequest struct {
	RunID           string `json:"run_id"`
	ProofB64        string `json:"proof_b64"`
	PublicInputsB64 string `json:"public_inputs_b64"`
}

// VerifyResponse is the JSON response for POST /api/verify.
type VerifyResponse struct {
	RunID    string `json:"run_id"`
	Verified bool   `json:"verified"`
	Message  string `json:"message,omitempty"`
}

// Handler handles POST /api/verify.
func Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, config.MaxVerifyBytes)

		var req VerifyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			if httputil.IsBodyTooLarge(err) {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
			return
		}
		runID := strings.TrimSpace(req.RunID)
		proofB64 := strings.TrimSpace(req.ProofB64)
		publicInputsB64 := strings.TrimSpace(req.PublicInputsB64)

		if runID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing run_id"})
			return
		}
		if proofB64 == "" || publicInputsB64 == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing proof or public inputs"})
			return
		}

		verified, msg, err := zk.VerifyProof(proofB64, publicInputsB64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusOK, VerifyResponse{
			RunID:    runID,
			Verified: verified,
			Message:  msg,
		})
	}
}
