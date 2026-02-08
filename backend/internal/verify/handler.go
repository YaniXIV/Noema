package verify

import (
	"encoding/base64"
	"fmt"
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

// Handler handles POST /api/verify. Stub verifier for now.
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
		if req.RunID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing run_id"})
			return
		}
		if req.ProofB64 == "" || req.PublicInputsB64 == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing proof or public inputs"})
			return
		}

		verified, msg, err := zk.VerifyProof(req.ProofB64, req.PublicInputsB64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}
		if verified || !shouldTryLegacy(msg) {
			c.JSON(http.StatusOK, VerifyResponse{
				RunID:    req.RunID,
				Verified: verified,
				Message:  msg,
			})
			return
		}

		verified, msg, err = verifyLegacyStub(req.RunID, req.ProofB64, req.PublicInputsB64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusOK, VerifyResponse{
			RunID:    req.RunID,
			Verified: verified,
			Message:  msg,
		})
	}
}

func shouldTryLegacy(msg string) bool {
	switch msg {
	case "invalid public inputs format", "invalid proof format":
		return true
	default:
		return false
	}
}

func verifyLegacyStub(runID, proofB64, publicInputsB64 string) (bool, string, error) {
	if runID == "" {
		return false, "missing run_id", fmt.Errorf("missing run_id")
	}
	proofRaw, err := base64.StdEncoding.DecodeString(proofB64)
	if err != nil {
		return false, "invalid proof encoding", fmt.Errorf("invalid proof encoding")
	}
	pubRaw, err := base64.StdEncoding.DecodeString(publicInputsB64)
	if err != nil {
		return false, "invalid public inputs encoding", fmt.Errorf("invalid public inputs encoding")
	}
	expectedProof := "stub_proof_" + runID
	expectedPub := "stub_inputs_" + runID
	if strings.HasPrefix(string(proofRaw), "noema_stub_proof_v1|") {
		return false, "unexpected proof format", nil
	}
	if string(proofRaw) != expectedProof || string(pubRaw) != expectedPub {
		return false, "legacy stub proof mismatch", nil
	}
	return true, "legacy stub verifier", nil
}
