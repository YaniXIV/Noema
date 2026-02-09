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
	RunID           string       `json:"run_id"`
	Status          string       `json:"status"` // PASS or FAIL
	OverallPass     bool         `json:"overall_pass"`
	MaxSeverity     int          `json:"max_severity"`
	Commitment      string       `json:"commitment"`
	ProofB64        string       `json:"proof_b64"`
	PublicInputsB64 string       `json:"public_inputs_b64"`
	PublicOutput    PublicOutput `json:"public_output"`
	Proof           Proof        `json:"proof"`
	Verified        bool         `json:"verified"`
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

		// Parse multipart: policy_config (string), dataset (file, required), images (files, optional)
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

		policyRaw, policyProvided, err := optionalFormValue(form, "policy_config")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var policyConfig PolicyConfig
		if policyProvided {
			policyConfig, err = parsePolicyConfig(policyRaw)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := validatePolicyConfig(policyConfig); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		} else if len(form.Value["spec"]) > 0 {
			spec, err := parseSpec(form)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := validateSpec(spec); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			policyConfig = policyConfigFromSpec(spec)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing field: policy_config"})
			return
		}

		datasetFile, imageFiles, err := parseUploads(form)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		runID, runPath, err := createRunDir(runsDir)
		if err != nil {
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
			log.Printf("save run files: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save run files"})
			return
		}

		evalOut, err := resolveEvaluationResult(c.Request.Context(), form, policyConfig, runsDir, datasetFile, imageFiles)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		overallPass, maxSeverity, policyThreshold := computePolicyResult(evalOut, policyConfig)
		status := "FAIL"
		if overallPass {
			status = "PASS"
		}

		policyJSON, err := jsonBytes(policyConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode policy_config"})
			return
		}
		evalJSON, err := jsonBytes(evalOut)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode evaluation result"})
			return
		}
		datasetDigest, err := datasetDigestHex(datasetFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute dataset digest"})
			return
		}
		witness, err := buildPolicyWitness(policyConfig, evalOut)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "proof generation failed"})
			return
		}
		witness.DatasetDigestHex = datasetDigest
		commitment, err := zk.CommitmentPoseidon(datasetDigest, witness.Enabled, witness.MaxAllowed, witness.Severity)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "proof generation failed"})
			return
		}

		log.Printf("policy_config=%s", string(policyJSON))
		log.Printf("evaluation_result=%s", string(evalJSON))
		proof, err := zk.GenerateProof(zk.PublicInputs{
			PolicyThreshold: policyThreshold,
			MaxSeverity:     maxSeverity,
			OverallPass:     overallPass,
			Commitment:      commitment,
			Witness:         witness,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "proof generation failed"})
			return
		}
		verified, reason, err := zk.VerifyProof(proof.ProofB64, proof.PublicInputsB64)
		if err != nil {
			log.Printf("proof verify error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "proof verification failed"})
			return
		}
		if !verified {
			log.Printf("proof verify failed: %s", reason)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "proof verification failed"})
			return
		}

		if err := saveRunMetadata(runPath, policyConfig, evalOut); err != nil {
			log.Printf("save run metadata: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist run metadata"})
			return
		}

		cleanupRun = false
		if err := updateRunsIndex(runsDir, config.RunsIndexLimit(), RunIndexEntry{
			RunID:          runID,
			Status:         status,
			Timestamp:      time.Now().Unix(),
			EvaluationName: "",
		}); err != nil {
			log.Printf("runs index update: %v", err)
		}

		if err := pruneRuns(runsDir, maxRuns); err != nil {
			log.Printf("prune runs: %v", err)
		}

		c.JSON(http.StatusOK, EvaluateResponse{
			RunID:           runID,
			Status:          status,
			OverallPass:     overallPass,
			MaxSeverity:     maxSeverity,
			Commitment:      commitment,
			ProofB64:        proof.ProofB64,
			PublicInputsB64: proof.PublicInputsB64,
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

func parseEvaluationResultOptional(form *multipart.Form, cfg PolicyConfig) (EvaluationResult, error) {
	out, provided, err := parseEvaluationResultProvided(form, cfg)
	if err != nil {
		return EvaluationResult{}, err
	}
	if !provided {
		return stubEvaluationResult(cfg), nil
	}
	return out, nil
}

func stubEvaluationResult(cfg PolicyConfig) EvaluationResult {
	out := EvaluationResult{
		EvalVersion: "noema_eval_v1",
		Results:     make([]EvalResultItem, 0, len(cfg.Constraints)),
	}
	for _, c := range cfg.Constraints {
		out.Results = append(out.Results, EvalResultItem{
			ID:        c.ID,
			Severity:  0,
			Rationale: "stub",
		})
	}
	sort.Slice(out.Results, func(i, j int) bool {
		return out.Results[i].ID < out.Results[j].ID
	})
	return out
}

func jsonBytes(v any) ([]byte, error) {
	return json.Marshal(v)
}

func parseEvaluationResultProvided(form *multipart.Form, cfg PolicyConfig) (EvaluationResult, bool, error) {
	if form == nil {
		return EvaluationResult{}, false, nil
	}
	if len(form.Value["evaluation_result"]) > 1 || len(form.Value["eval_output"]) > 1 {
		return EvaluationResult{}, true, fmt.Errorf("only one evaluation_result value allowed")
	}
	hasEval := len(form.Value["evaluation_result"]) > 0
	hasLegacy := len(form.Value["eval_output"]) > 0
	raw := strings.TrimSpace(formValue(form, "evaluation_result"))
	if raw == "" && hasEval {
		return EvaluationResult{}, true, fmt.Errorf("evaluation_result must be non-empty")
	}
	if raw == "" {
		raw = strings.TrimSpace(formValue(form, "eval_output"))
		if raw == "" && hasLegacy {
			return EvaluationResult{}, true, fmt.Errorf("evaluation_result must be non-empty")
		}
	}
	if raw == "" {
		return EvaluationResult{}, false, nil
	}
	out, err := parseEvaluationResult(raw)
	if err != nil {
		return EvaluationResult{}, true, err
	}
	if err := validateEvaluationResult(out, cfg); err != nil {
		return EvaluationResult{}, true, err
	}
	return out, true, nil
}

func singleFormValue(form *multipart.Form, key string) (string, error) {
	if form == nil || len(form.Value[key]) == 0 {
		return "", fmt.Errorf("missing field: %s", key)
	}
	if len(form.Value[key]) > 1 {
		return "", fmt.Errorf("only one %s value allowed", key)
	}
	raw := strings.TrimSpace(form.Value[key][0])
	if raw == "" {
		return "", fmt.Errorf("%s must be non-empty", key)
	}
	return raw, nil
}

func optionalFormValue(form *multipart.Form, key string) (string, bool, error) {
	if form == nil || len(form.Value[key]) == 0 {
		return "", false, nil
	}
	if len(form.Value[key]) > 1 {
		return "", true, fmt.Errorf("only one %s value allowed", key)
	}
	raw := strings.TrimSpace(form.Value[key][0])
	if raw == "" {
		return "", true, fmt.Errorf("%s must be non-empty", key)
	}
	return raw, true, nil
}

func formValue(form *multipart.Form, key string) string {
	if form == nil || len(form.Value[key]) == 0 {
		return ""
	}
	return form.Value[key][0]
}

var policyConstraintOrder = []string{
	"pii_exposure_risk",
	"regulated_sensitive_data_presence",
	"data_provenance_or_consent_violation_risk",
	"safety_critical_advisory_presence",
	"harm_enabling_content_risk",
	"dataset_intended_use_mismatch",
}

func buildPolicyWitness(cfg PolicyConfig, out EvaluationResult) (*zk.WitnessInputs, error) {
	if len(policyConstraintOrder) != zk.PolicyGateConstraintCount {
		return nil, fmt.Errorf("policy constraint ordering mismatch")
	}
	cfgByID := make(map[string]PolicyConstraint, len(cfg.Constraints))
	for _, c := range cfg.Constraints {
		cfgByID[c.ID] = c
	}
	resultsByID := make(map[string]EvalResultItem, len(out.Results))
	for _, r := range out.Results {
		resultsByID[r.ID] = r
	}

	known := make(map[string]struct{}, len(policyConstraintOrder))
	for _, id := range policyConstraintOrder {
		known[id] = struct{}{}
	}
	for id := range cfgByID {
		if _, ok := known[id]; !ok {
			return nil, fmt.Errorf("unsupported constraint id: %s", id)
		}
	}

	var enabled [zk.PolicyGateConstraintCount]uint64
	var maxAllowed [zk.PolicyGateConstraintCount]uint64
	var severity [zk.PolicyGateConstraintCount]uint64

	for i, id := range policyConstraintOrder {
		c, ok := cfgByID[id]
		if !ok {
			enabled[i] = 0
			maxAllowed[i] = 0
			severity[i] = 0
			continue
		}
		if c.Enabled {
			enabled[i] = 1
		} else {
			enabled[i] = 0
		}
		maxAllowed[i] = uint64(c.MaxAllowed)
		r, ok := resultsByID[id]
		if !ok {
			return nil, fmt.Errorf("missing evaluation result for %s", id)
		}
		severity[i] = uint64(r.Severity)
	}

	return &zk.WitnessInputs{
		Enabled:    enabled,
		MaxAllowed: maxAllowed,
		Severity:   severity,
	}, nil
}
