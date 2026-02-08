package evaluate

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"noema/internal/config"
	"noema/internal/httputil"
	"noema/internal/zk"

	"github.com/gin-gonic/gin"
)

// EvaluateResponse is the JSON response for POST /api/evaluate.
type EvaluateResponse struct {
	RunID        string       `json:"run_id"`
	Status       string       `json:"status"` // PASS or FAIL
	PublicOutput PublicOutput `json:"public_output"`
	Proof        Proof        `json:"proof"`
	Verified     bool         `json:"verified"`
}

type PublicOutput struct {
	OverallPass     bool   `json:"overall_pass"`
	MaxSeverity     int    `json:"max_severity"`
	PolicyThreshold int    `json:"policy_threshold"`
	Commitment      string `json:"commitment"`
}

type Proof struct {
	System          string `json:"system"`
	Curve           string `json:"curve"`
	ProofB64        string `json:"proof_b64"`
	PublicInputsB64 string `json:"public_inputs_b64"`
}

// Handler handles POST /api/evaluate. Expects CookieAuth to have run first.
func Handler(runsDir string, maxRuns int) gin.HandlerFunc {
	return func(c *gin.Context) {
		const multipartOverhead = 2 << 20
		maxBody := int64(config.MaxDatasetBytes) + int64(config.MaxImages*config.MaxImageBytes) + multipartOverhead
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBody)

		// Parse multipart: spec (string), dataset (file, required), images (files, optional)
		form, err := c.MultipartForm()
		if err != nil {
			if httputil.IsBodyTooLarge(err) {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
			return
		}
		defer form.RemoveAll()

		spec, err := parseSpec(form)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := validateSpec(spec); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		datasetFile, imageFiles, err := parseUploads(form)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		runID := genRunID()
		runPath := filepath.Join(runsDir, runID)
		if err := ensureRunDir(runPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create run directory"})
			return
		}
		cleanupRun := true
		defer func() {
			if cleanupRun {
				_ = os.RemoveAll(runPath)
			}
		}()

		if err := saveRunFiles(runPath, datasetFile, imageFiles); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		enabled, err := enabledConstraints(spec)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		evalOut, err := resolveEvalOutput(c.Request.Context(), form, enabled, runsDir, spec, datasetFile, imageFiles)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		overallPass, maxSeverity, policyThreshold := computePolicyResult(evalOut, enabled)
		status := "FAIL"
		if overallPass {
			status = "PASS"
		}

		specJSON, err := jsonBytes(spec)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode spec"})
			return
		}
		evalJSON, err := jsonBytes(evalOut)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode evaluation output"})
			return
		}
		commitment := zk.CommitmentSHA256([]byte("spec"), specJSON, []byte("eval"), evalJSON)
		proof, err := zk.GenerateProof(zk.PublicInputs{
			PolicyThreshold: policyThreshold,
			MaxSeverity:     maxSeverity,
			OverallPass:     overallPass,
			Commitment:      commitment,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "proof generation failed"})
			return
		}
		verified, _, _ := zk.VerifyProof(proof.ProofB64, proof.PublicInputsB64)

		cleanupRun = false
		if err := updateRunsIndex(runsDir, config.RunsIndexLimit(), RunIndexEntry{
			RunID:          runID,
			Status:         status,
			Timestamp:      time.Now().Unix(),
			EvaluationName: spec.EvaluationName,
		}); err != nil {
			log.Printf("runs index update: %v", err)
		}

		if err := pruneRuns(runsDir, maxRuns); err != nil {
			log.Printf("prune runs: %v", err)
		}

		c.JSON(http.StatusOK, EvaluateResponse{
			RunID:  runID,
			Status: status,
			PublicOutput: PublicOutput{
				OverallPass:     overallPass,
				MaxSeverity:     maxSeverity,
				PolicyThreshold: policyThreshold,
				Commitment:      commitment,
			},
			Proof: Proof{
				System:          proof.System,
				Curve:           proof.Curve,
				ProofB64:        proof.ProofB64,
				PublicInputsB64: proof.PublicInputsB64,
			},
			Verified: verified,
		})
	}
}

type runEntry struct {
	path    string
	modTime time.Time
}

func pruneRuns(runsDir string, maxRuns int) error {
	if maxRuns <= 0 {
		return nil
	}
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var runs []runEntry
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "run_") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		runs = append(runs, runEntry{
			path:    filepath.Join(runsDir, entry.Name()),
			modTime: info.ModTime(),
		})
	}

	if len(runs) <= maxRuns {
		return nil
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i].modTime.After(runs[j].modTime)
	})

	for i := maxRuns; i < len(runs); i++ {
		if err := os.RemoveAll(runs[i].path); err != nil {
			return err
		}
	}
	return nil
}

func parseEvalOutputOptional(form *multipart.Form, enabled map[string]ConstraintRule) (EvalOutput, error) {
	out, provided, err := parseEvalOutputProvided(form, enabled)
	if err != nil {
		return EvalOutput{}, err
	}
	if !provided {
		return stubEvalOutput(enabled), nil
	}
	return out, nil
}

func stubEvalOutput(enabled map[string]ConstraintRule) EvalOutput {
	out := EvalOutput{
		SchemaVersion: 1,
		Constraints:   make([]EvalConstraintResult, 0, len(enabled)),
		MaxSeverity:   0,
	}
	for id := range enabled {
		out.Constraints = append(out.Constraints, EvalConstraintResult{
			ID:        id,
			Severity:  0,
			Rationale: "stub",
		})
	}
	sort.Slice(out.Constraints, func(i, j int) bool {
		return out.Constraints[i].ID < out.Constraints[j].ID
	})
	return out
}

func jsonBytes(v any) ([]byte, error) {
	return json.Marshal(v)
}

func parseEvalOutputProvided(form *multipart.Form, enabled map[string]ConstraintRule) (EvalOutput, bool, error) {
	if form == nil || len(form.Value["eval_output"]) == 0 {
		return EvalOutput{}, false, nil
	}
	if len(form.Value["eval_output"]) > 1 {
		return EvalOutput{}, true, fmt.Errorf("only one eval_output value allowed")
	}
	raw := strings.TrimSpace(form.Value["eval_output"][0])
	if raw == "" {
		return EvalOutput{}, false, nil
	}
	out, err := parseEvalOutput(raw)
	if err != nil {
		return EvalOutput{}, true, err
	}
	if err := validateEvalOutput(out, enabled); err != nil {
		return EvalOutput{}, true, err
	}
	return out, true, nil
}
