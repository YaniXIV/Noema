package evaluate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

type formFile struct {
	field       string
	filename    string
	contentType string
	content     []byte
}

func buildMultipartForm(t *testing.T, files []formFile) *multipart.Form {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for _, f := range files {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, f.field, f.filename))
		if f.contentType != "" {
			header.Set("Content-Type", f.contentType)
		}
		part, err := writer.CreatePart(header)
		if err != nil {
			t.Fatalf("create part: %v", err)
		}
		if _, err := part.Write(f.content); err != nil {
			t.Fatalf("write part: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	reader := multipart.NewReader(&buf, writer.Boundary())
	form, err := reader.ReadForm(10 << 20)
	if err != nil {
		t.Fatalf("read form: %v", err)
	}
	t.Cleanup(func() {
		form.RemoveAll()
	})
	return form
}

func TestParseEvaluationResultOptional_DefaultsToStub(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "pii_exposure_risk", Enabled: true, MaxAllowed: 1},
			{ID: "harm_enabling_content_risk", Enabled: true, MaxAllowed: 2},
		},
	}

	form := &multipart.Form{Value: map[string][]string{}}
	out, err := parseEvaluationResultOptional(form, cfg)
	if err != nil {
		t.Fatalf("parseEvaluationResultOptional error: %v", err)
	}
	if out.EvalVersion != "noema_eval_v1" {
		t.Fatalf("expected eval_version noema_eval_v1, got %s", out.EvalVersion)
	}
	if len(out.Results) != len(cfg.Constraints) {
		t.Fatalf("expected %d results, got %d", len(cfg.Constraints), len(out.Results))
	}

	ids := make([]string, 0, len(out.Results))
	for _, r := range out.Results {
		if r.Severity != 0 {
			t.Fatalf("expected severity 0 for %s, got %d", r.ID, r.Severity)
		}
		if r.Rationale == "" {
			t.Fatalf("expected rationale for %s", r.ID)
		}
		ids = append(ids, r.ID)
	}
	if !sort.StringsAreSorted(ids) {
		t.Fatalf("expected results sorted by id, got %v", ids)
	}
}

func TestParseUploads_RejectsImageFilenameWhitespace(t *testing.T) {
	dataset := `{"items":[{"id":"item-1","text":"hello"}]}`
	form := buildMultipartForm(t, []formFile{
		{
			field:       "dataset",
			filename:    "dataset.json",
			contentType: "application/json",
			content:     []byte(dataset),
		},
		{
			field:       "images",
			filename:    " bad.png ",
			contentType: "image/png",
			content:     []byte("fake"),
		},
	})

	if _, _, err := parseUploads(form); err == nil {
		t.Fatalf("expected error for image filename with whitespace")
	} else if !strings.Contains(err.Error(), "leading/trailing whitespace") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseEvaluationResultOptional_UsesProvidedOutput(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "pii_exposure_risk", Enabled: true, MaxAllowed: 1},
			{ID: "harm_enabling_content_risk", Enabled: true, MaxAllowed: 2},
		},
	}

	payload := EvaluationResult{
		EvalVersion: "noema_eval_v1",
		Results: []EvalResultItem{
			{ID: "pii_exposure_risk", Severity: 1, Rationale: "limited"},
			{ID: "harm_enabling_content_risk", Severity: 0, Rationale: "none"},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal evaluation_result: %v", err)
	}

	form := &multipart.Form{Value: map[string][]string{"evaluation_result": {string(raw)}}}
	out, err := parseEvaluationResultOptional(form, cfg)
	if err != nil {
		t.Fatalf("parseEvaluationResultOptional error: %v", err)
	}
	if len(out.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out.Results))
	}
}

func TestParseEvaluationResultOptional_RejectsInvalidOutput(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "pii_exposure_risk", Enabled: true, MaxAllowed: 1},
		},
	}

	payload := EvaluationResult{
		EvalVersion: "noema_eval_v1",
		Results: []EvalResultItem{
			{ID: "unknown_constraint", Severity: 2, Rationale: "bad"},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal evaluation_result: %v", err)
	}

	form := &multipart.Form{Value: map[string][]string{"evaluation_result": {string(raw)}}}
	if _, err := parseEvaluationResultOptional(form, cfg); err == nil {
		t.Fatalf("expected validation error for unknown constraint")
	}
}

func TestParseEvaluationResultOptional_RejectsTrailingGarbage(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "pii_exposure_risk", Enabled: true, MaxAllowed: 1},
		},
	}

	payload := EvaluationResult{
		EvalVersion: "noema_eval_v1",
		Results: []EvalResultItem{
			{ID: "pii_exposure_risk", Severity: 0, Rationale: "ok"},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal evaluation_result: %v", err)
	}

	form := &multipart.Form{Value: map[string][]string{"evaluation_result": {string(raw) + " trailing"}}}
	if _, err := parseEvaluationResultOptional(form, cfg); err == nil {
		t.Fatalf("expected error for trailing garbage")
	}
}

func TestParseEvaluationResultOptional_RejectsMultipleValues(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "pii_exposure_risk", Enabled: true, MaxAllowed: 1},
		},
	}

	payload := EvaluationResult{
		EvalVersion: "noema_eval_v1",
		Results: []EvalResultItem{
			{ID: "pii_exposure_risk", Severity: 0, Rationale: "ok"},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal evaluation_result: %v", err)
	}

	form := &multipart.Form{Value: map[string][]string{"evaluation_result": {string(raw), string(raw)}}}
	if _, err := parseEvaluationResultOptional(form, cfg); err == nil {
		t.Fatalf("expected error for multiple evaluation_result values")
	}
}

func TestParseEvaluationResultOptional_RejectsWhitespace(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "pii_exposure_risk", Enabled: true, MaxAllowed: 1},
			{ID: "harm_enabling_content_risk", Enabled: true, MaxAllowed: 2},
		},
	}

	form := &multipart.Form{Value: map[string][]string{"evaluation_result": {" \n\t "}}}
	if _, err := parseEvaluationResultOptional(form, cfg); err == nil {
		t.Fatalf("expected error for blank evaluation_result")
	}
}

func TestComputePolicyResult(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "pii", Enabled: true, MaxAllowed: 1},
			{ID: "safety", Enabled: true, MaxAllowed: 2},
		},
	}
	out := EvaluationResult{
		EvalVersion: "noema_eval_v1",
		Results: []EvalResultItem{
			{ID: "pii", Severity: 2, Rationale: "ok"},
			{ID: "safety", Severity: 2, Rationale: "ok"},
		},
	}
	overall, maxSeverity, threshold := computePolicyResult(out, cfg)
	if threshold != 1 {
		t.Fatalf("expected threshold 1, got %d", threshold)
	}
	if maxSeverity != 2 {
		t.Fatalf("expected max severity 2, got %d", maxSeverity)
	}
	if overall {
		t.Fatalf("expected overall pass false due to pii threshold")
	}
}

func TestValidateDatasetJSON_RejectsEmptyFile(t *testing.T) {
	form := buildMultipartForm(t, []formFile{
		{field: "dataset", filename: "dataset.json", contentType: "application/json", content: []byte{}},
	})
	datasetFile := form.File["dataset"][0]
	if err := validateDatasetJSON(datasetFile, nil); err == nil {
		t.Fatalf("expected error for empty dataset file")
	}
}

func TestParseUploads_RejectsMultipleDatasetFiles(t *testing.T) {
	dataset := `{"items":[{"id":"1","text":"hello"}]}`
	form := buildMultipartForm(t, []formFile{
		{field: "dataset", filename: "dataset.json", contentType: "application/json", content: []byte(dataset)},
		{field: "dataset", filename: "dataset2.json", contentType: "application/json", content: []byte(dataset)},
	})
	if _, _, err := parseUploads(form); err == nil {
		t.Fatalf("expected error for multiple dataset files")
	}
}

func TestParseUploads_RejectsDuplicateImageNames(t *testing.T) {
	dataset := `{"items":[{"id":"1","text":"hello","image_ref":"img.png"}]}`
	form := buildMultipartForm(t, []formFile{
		{field: "dataset", filename: "dataset.json", contentType: "application/json", content: []byte(dataset)},
		{field: "images", filename: "img.png", contentType: "image/png", content: []byte("png")},
		{field: "images", filename: "img.png", contentType: "image/png", content: []byte("png")},
	})
	if _, _, err := parseUploads(form); err == nil {
		t.Fatalf("expected error for duplicate image filenames")
	}
}

func TestValidatePolicyConfig_RejectsEmptyConstraintIDs(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "pii_exposure_risk", Enabled: true, MaxAllowed: 1},
			{ID: "   ", Enabled: true, MaxAllowed: 1},
		},
	}
	if err := validatePolicyConfig(cfg); err == nil {
		t.Fatalf("expected error for empty constraint id")
	}
}

func TestValidatePolicyConfig_RejectsDuplicateConstraintIDs(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "pii_exposure_risk", Enabled: true, MaxAllowed: 1},
			{ID: "pii_exposure_risk", Enabled: false, MaxAllowed: 2},
		},
	}
	if err := validatePolicyConfig(cfg); err == nil {
		t.Fatalf("expected error for duplicate constraint id")
	}
}

func TestValidatePolicyConfig_RejectsWhitespaceConstraintIDs(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: " pii_exposure_risk ", Enabled: true, MaxAllowed: 1},
		},
	}
	if err := validatePolicyConfig(cfg); err == nil {
		t.Fatalf("expected error for whitespace in constraint id")
	}
}

func TestValidatePolicyConfig_RejectsInvalidVersion(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v2",
		Constraints: []PolicyConstraint{
			{ID: "pii_exposure_risk", Enabled: true, MaxAllowed: 1},
		},
	}
	if err := validatePolicyConfig(cfg); err == nil {
		t.Fatalf("expected error for invalid policy_version")
	}
}

func TestParsePolicyConfig_RejectsUnknownFields(t *testing.T) {
	raw := `{"policy_version":"noema_policy_v1","constraints":[{"id":"pii_exposure_risk","enabled":true,"max_allowed":1}],"extra":true}`
	if _, err := parsePolicyConfig(raw); err == nil {
		t.Fatalf("expected error for unknown policy_config fields")
	}
}

func TestPruneRuns_IgnoresNonRunDirectories(t *testing.T) {
	base := t.TempDir()

	runOld := filepath.Join(base, "run_old")
	runNew := filepath.Join(base, "run_new")
	cacheDir := filepath.Join(base, "cache")
	otherDir := filepath.Join(base, "misc")

	for _, dir := range []string{runOld, runNew, cacheDir, otherDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	oldTime := time.Now().Add(-2 * time.Hour)
	newTime := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(runOld, oldTime, oldTime); err != nil {
		t.Fatalf("chtimes run_old: %v", err)
	}
	if err := os.Chtimes(runNew, newTime, newTime); err != nil {
		t.Fatalf("chtimes run_new: %v", err)
	}

	if err := pruneRuns(base, 1); err != nil {
		t.Fatalf("pruneRuns error: %v", err)
	}

	if _, err := os.Stat(runOld); err == nil {
		t.Fatalf("expected run_old to be pruned")
	}
	if _, err := os.Stat(runNew); err != nil {
		t.Fatalf("expected run_new to remain, got err: %v", err)
	}
	if _, err := os.Stat(cacheDir); err != nil {
		t.Fatalf("expected cache dir to remain, got err: %v", err)
	}
	if _, err := os.Stat(otherDir); err != nil {
		t.Fatalf("expected misc dir to remain, got err: %v", err)
	}
}

func TestPruneRuns_MissingDirNoError(t *testing.T) {
	base := filepath.Join(t.TempDir(), "missing")
	if err := pruneRuns(base, 1); err != nil {
		t.Fatalf("expected no error for missing dir, got %v", err)
	}
}
