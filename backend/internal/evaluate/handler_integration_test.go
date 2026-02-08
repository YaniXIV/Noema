package evaluate

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestEvaluateHandler_WithEvalOutput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	runsDir := t.TempDir()
	router.POST("/api/evaluate", Handler(runsDir, 0))

	spec := Spec{
		SchemaVersion:  1,
		EvaluationName: "test",
		Policy:         Policy{Reveal: RevealPolicy{MaxSeverity: true, Commitment: true}},
		Constraints: []Constraint{
			{ID: "pii_exposure_risk", Enabled: true, AllowedMaxSeverity: 1},
		},
	}
	evalOut := EvalOutput{
		SchemaVersion: 1,
		Constraints: []EvalConstraintResult{
			{ID: "pii_exposure_risk", Severity: 2, Rationale: "clear identifiers"},
		},
		MaxSeverity: 2,
	}

	body, contentType := buildMultipartEvalRequest(t, spec, evalOut, true)
	req := httptest.NewRequest(http.MethodPost, "/api/evaluate", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp EvaluateResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != "FAIL" {
		t.Fatalf("expected status FAIL, got %s", resp.Status)
	}
	if resp.PublicOutput.MaxSeverity != 2 {
		t.Fatalf("expected max severity 2, got %d", resp.PublicOutput.MaxSeverity)
	}
	if resp.PublicOutput.PolicyThreshold != 1 {
		t.Fatalf("expected policy threshold 1, got %d", resp.PublicOutput.PolicyThreshold)
	}
	if resp.Proof.ProofB64 == "" || resp.Proof.PublicInputsB64 == "" {
		t.Fatalf("expected proof fields to be populated")
	}
}

func TestEvaluateHandler_StubEvalOutput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	runsDir := t.TempDir()
	router.POST("/api/evaluate", Handler(runsDir, 0))

	spec := Spec{
		SchemaVersion:  1,
		EvaluationName: "test",
		Policy:         Policy{Reveal: RevealPolicy{MaxSeverity: true, Commitment: true}},
		Constraints: []Constraint{
			{ID: "pii_exposure_risk", Enabled: true, AllowedMaxSeverity: 1},
			{ID: "harm_enabling_content_risk", Enabled: true, AllowedMaxSeverity: 2},
		},
	}

	body, contentType := buildMultipartEvalRequest(t, spec, EvalOutput{}, false)
	req := httptest.NewRequest(http.MethodPost, "/api/evaluate", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp EvaluateResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != "PASS" {
		t.Fatalf("expected status PASS, got %s", resp.Status)
	}
	if resp.PublicOutput.MaxSeverity != 0 {
		t.Fatalf("expected max severity 0, got %d", resp.PublicOutput.MaxSeverity)
	}
	if resp.Proof.ProofB64 == "" || resp.Proof.PublicInputsB64 == "" {
		t.Fatalf("expected proof fields to be populated")
	}
}

func TestEvaluateHandler_InvalidDataset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	runsDir := t.TempDir()
	router.POST("/api/evaluate", Handler(runsDir, 0))

	spec := Spec{
		SchemaVersion:  1,
		EvaluationName: "test",
		Constraints: []Constraint{
			{ID: "pii_exposure_risk", Enabled: true, AllowedMaxSeverity: 1},
		},
	}

	body, contentType := buildMultipartEvalRequestWithDataset(t, spec, `{"items":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/evaluate", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEvaluateHandler_InvalidEvalOutput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	runsDir := t.TempDir()
	router.POST("/api/evaluate", Handler(runsDir, 0))

	spec := Spec{
		SchemaVersion:  1,
		EvaluationName: "test",
		Constraints: []Constraint{
			{ID: "pii_exposure_risk", Enabled: true, AllowedMaxSeverity: 1},
		},
	}
	evalOut := EvalOutput{
		SchemaVersion: 1,
		Constraints: []EvalConstraintResult{
			{ID: "unknown", Severity: 2, Rationale: "bad"},
		},
		MaxSeverity: 2,
	}

	body, contentType := buildMultipartEvalRequest(t, spec, evalOut, true)
	req := httptest.NewRequest(http.MethodPost, "/api/evaluate", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEvaluateHandler_CleansUpFailedRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	runsDir := t.TempDir()
	router.POST("/api/evaluate", Handler(runsDir, 0))

	spec := Spec{
		SchemaVersion:  1,
		EvaluationName: "test",
		Constraints: []Constraint{
			{ID: "pii_exposure_risk", Enabled: true, AllowedMaxSeverity: 1},
		},
	}
	evalOut := EvalOutput{
		SchemaVersion: 1,
		Constraints: []EvalConstraintResult{
			{ID: "unknown", Severity: 2, Rationale: "bad"},
		},
		MaxSeverity: 2,
	}

	body, contentType := buildMultipartEvalRequest(t, spec, evalOut, true)
	req := httptest.NewRequest(http.MethodPost, "/api/evaluate", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}

	entries, err := os.ReadDir(runsDir)
	if err != nil {
		t.Fatalf("read runs dir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "run_") {
			t.Fatalf("expected failed runs to be cleaned up, found %s", entry.Name())
		}
	}
}

func TestEvaluateHandler_WithImages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	runsDir := t.TempDir()
	router.POST("/api/evaluate", Handler(runsDir, 0))

	spec := Spec{
		SchemaVersion:  1,
		EvaluationName: "test",
		Constraints: []Constraint{
			{ID: "pii_exposure_risk", Enabled: true, AllowedMaxSeverity: 1},
		},
	}

	body, contentType := buildMultipartEvalRequestWithImages(t, spec)
	req := httptest.NewRequest(http.MethodPost, "/api/evaluate", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func buildMultipartEvalRequest(t *testing.T, spec Spec, evalOut EvalOutput, includeEvalOutput bool) (*bytes.Buffer, string) {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	specRaw, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal spec: %v", err)
	}
	if err := writer.WriteField("spec", string(specRaw)); err != nil {
		t.Fatalf("write spec field: %v", err)
	}

	if includeEvalOutput {
		evalRaw, err := json.Marshal(evalOut)
		if err != nil {
			t.Fatalf("marshal eval_output: %v", err)
		}
		if err := writer.WriteField("eval_output", string(evalRaw)); err != nil {
			t.Fatalf("write eval_output field: %v", err)
		}
	}

	part, err := writer.CreateFormFile("dataset", "dataset.json")
	if err != nil {
		t.Fatalf("create dataset part: %v", err)
	}
	if _, err := part.Write([]byte(`{"items":[{"id":"1","text":"hello"}]}`)); err != nil {
		t.Fatalf("write dataset: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	return &buf, writer.FormDataContentType()
}

func buildMultipartEvalRequestWithDataset(t *testing.T, spec Spec, datasetJSON string) (*bytes.Buffer, string) {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	specRaw, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal spec: %v", err)
	}
	if err := writer.WriteField("spec", string(specRaw)); err != nil {
		t.Fatalf("write spec field: %v", err)
	}

	part, err := writer.CreateFormFile("dataset", "dataset.json")
	if err != nil {
		t.Fatalf("create dataset part: %v", err)
	}
	if _, err := part.Write([]byte(datasetJSON)); err != nil {
		t.Fatalf("write dataset: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	return &buf, writer.FormDataContentType()
}

func buildMultipartEvalRequestWithImages(t *testing.T, spec Spec) (*bytes.Buffer, string) {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	specRaw, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal spec: %v", err)
	}
	if err := writer.WriteField("spec", string(specRaw)); err != nil {
		t.Fatalf("write spec field: %v", err)
	}

	part, err := writer.CreateFormFile("dataset", "dataset.json")
	if err != nil {
		t.Fatalf("create dataset part: %v", err)
	}
	if _, err := part.Write([]byte(`{"items":[{"id":"1","text":"hello"}]}`)); err != nil {
		t.Fatalf("write dataset: %v", err)
	}

	img1, err := writer.CreateFormFile("images", "img1.png")
	if err != nil {
		t.Fatalf("create image part: %v", err)
	}
	if _, err := img1.Write([]byte{0x89, 0x50, 0x4e, 0x47}); err != nil {
		t.Fatalf("write image: %v", err)
	}

	img2, err := writer.CreateFormFile("images", "img2.jpg")
	if err != nil {
		t.Fatalf("create image part: %v", err)
	}
	if _, err := img2.Write([]byte{0xff, 0xd8, 0xff, 0xdb}); err != nil {
		t.Fatalf("write image: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	return &buf, writer.FormDataContentType()
}
