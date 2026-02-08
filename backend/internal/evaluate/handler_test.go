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

func TestParseEvalOutputOptional_DefaultsToStub(t *testing.T) {
	spec := Spec{
		SchemaVersion: 1,
		Constraints: []Constraint{
			{ID: "pii_exposure_risk", Enabled: true, AllowedMaxSeverity: 1},
			{ID: "harm_enabling_content_risk", Enabled: true, AllowedMaxSeverity: 2},
		},
	}
	enabled, err := enabledConstraints(spec)
	if err != nil {
		t.Fatalf("enabledConstraints error: %v", err)
	}

	form := &multipart.Form{Value: map[string][]string{}}
	out, err := parseEvalOutputOptional(form, enabled)
	if err != nil {
		t.Fatalf("parseEvalOutputOptional error: %v", err)
	}
	if out.SchemaVersion != 1 {
		t.Fatalf("expected schema_version 1, got %d", out.SchemaVersion)
	}
	if out.MaxSeverity != 0 {
		t.Fatalf("expected max_severity 0, got %d", out.MaxSeverity)
	}
	if len(out.Constraints) != len(enabled) {
		t.Fatalf("expected %d constraints, got %d", len(enabled), len(out.Constraints))
	}

	ids := make([]string, 0, len(out.Constraints))
	for _, c := range out.Constraints {
		if c.Severity != 0 {
			t.Fatalf("expected severity 0 for %s, got %d", c.ID, c.Severity)
		}
		if c.Rationale == "" {
			t.Fatalf("expected rationale for %s", c.ID)
		}
		ids = append(ids, c.ID)
	}
	if !sort.StringsAreSorted(ids) {
		t.Fatalf("expected constraints sorted by id, got %v", ids)
	}
}

func TestParseEvalOutputOptional_UsesProvidedOutput(t *testing.T) {
	spec := Spec{
		SchemaVersion: 1,
		Constraints: []Constraint{
			{ID: "pii_exposure_risk", Enabled: true, AllowedMaxSeverity: 1},
			{ID: "harm_enabling_content_risk", Enabled: true, AllowedMaxSeverity: 2},
		},
	}
	enabled, err := enabledConstraints(spec)
	if err != nil {
		t.Fatalf("enabledConstraints error: %v", err)
	}

	payload := EvalOutput{
		SchemaVersion: 1,
		Constraints: []EvalConstraintResult{
			{ID: "pii_exposure_risk", Severity: 1, Rationale: "limited"},
			{ID: "harm_enabling_content_risk", Severity: 0, Rationale: "none"},
		},
		MaxSeverity: 1,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal eval_output: %v", err)
	}

	form := &multipart.Form{Value: map[string][]string{"eval_output": {string(raw)}}}
	out, err := parseEvalOutputOptional(form, enabled)
	if err != nil {
		t.Fatalf("parseEvalOutputOptional error: %v", err)
	}
	if out.MaxSeverity != 1 {
		t.Fatalf("expected max_severity 1, got %d", out.MaxSeverity)
	}
	if len(out.Constraints) != 2 {
		t.Fatalf("expected 2 constraints, got %d", len(out.Constraints))
	}
}

func TestParseEvalOutputOptional_RejectsInvalidOutput(t *testing.T) {
	spec := Spec{
		SchemaVersion: 1,
		Constraints: []Constraint{
			{ID: "pii_exposure_risk", Enabled: true, AllowedMaxSeverity: 1},
		},
	}
	enabled, err := enabledConstraints(spec)
	if err != nil {
		t.Fatalf("enabledConstraints error: %v", err)
	}

	payload := EvalOutput{
		SchemaVersion: 1,
		Constraints: []EvalConstraintResult{
			{ID: "unknown_constraint", Severity: 2, Rationale: "bad"},
		},
		MaxSeverity: 2,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal eval_output: %v", err)
	}

	form := &multipart.Form{Value: map[string][]string{"eval_output": {string(raw)}}}
	if _, err := parseEvalOutputOptional(form, enabled); err == nil {
		t.Fatalf("expected validation error for unknown constraint")
	}
}

func TestComputePolicyResult(t *testing.T) {
	enabled := map[string]ConstraintRule{
		"pii":    {ID: "pii", AllowedMaxSeverity: 1},
		"safety": {ID: "safety", AllowedMaxSeverity: 2},
	}
	out := EvalOutput{
		SchemaVersion: 1,
		Constraints: []EvalConstraintResult{
			{ID: "pii", Severity: 2, Rationale: "ok"},
			{ID: "safety", Severity: 2, Rationale: "ok"},
		},
		MaxSeverity: 2,
	}
	overall, maxSeverity, threshold := computePolicyResult(out, enabled)
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

func TestValidateDatasetJSON_ImageRefRequiresImages(t *testing.T) {
	dataset := `{"items":[{"id":"1","text":"hello","image_ref":"img1.png"}]}`
	form := buildMultipartForm(t, []formFile{
		{field: "dataset", filename: "dataset.json", contentType: "application/json", content: []byte(dataset)},
	})
	datasetFile := form.File["dataset"][0]
	if err := validateDatasetJSON(datasetFile, nil); err == nil {
		t.Fatalf("expected error for image_ref without images")
	}
}

func TestValidateDatasetJSON_ImageRefMatchesUpload(t *testing.T) {
	dataset := `{"items":[{"id":"1","text":"hello","image_ref":"img1.png"}]}`
	form := buildMultipartForm(t, []formFile{
		{field: "dataset", filename: "dataset.json", contentType: "application/json", content: []byte(dataset)},
		{field: "images", filename: "img1.png", contentType: "image/png", content: []byte("png")},
	})
	datasetFile := form.File["dataset"][0]
	imageFiles := form.File["images"]
	if err := validateDatasetJSON(datasetFile, imageFiles); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateDatasetJSON_RejectsDuplicateIDs(t *testing.T) {
	dataset := `{"items":[{"id":"1","text":"hello"},{"id":"1","text":"world"}]}`
	form := buildMultipartForm(t, []formFile{
		{field: "dataset", filename: "dataset.json", contentType: "application/json", content: []byte(dataset)},
	})
	datasetFile := form.File["dataset"][0]
	if err := validateDatasetJSON(datasetFile, nil); err == nil {
		t.Fatalf("expected error for duplicate dataset ids")
	}
}

func TestValidateDatasetJSON_RejectsUnknownFields(t *testing.T) {
	dataset := `{"items":[{"id":"1","text":"hello","extra":"nope"}]}`
	form := buildMultipartForm(t, []formFile{
		{field: "dataset", filename: "dataset.json", contentType: "application/json", content: []byte(dataset)},
	})
	datasetFile := form.File["dataset"][0]
	if err := validateDatasetJSON(datasetFile, nil); err == nil {
		t.Fatalf("expected error for unknown dataset fields")
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

func TestValidateSpec_RejectsEmptyConstraintIDs(t *testing.T) {
	spec := Spec{
		SchemaVersion: 1,
		Constraints: []Constraint{
			{ID: "pii_exposure_risk", Enabled: true, AllowedMaxSeverity: 1},
			{ID: "   ", Enabled: true, AllowedMaxSeverity: 1},
		},
	}
	if err := validateSpec(spec); err == nil {
		t.Fatalf("expected error for empty constraint id")
	}
}

func TestValidateSpec_RejectsEmptyCustomConstraintIDs(t *testing.T) {
	spec := Spec{
		SchemaVersion: 1,
		CustomConstraints: []CustomConstraint{
			{ID: "", Title: "Custom", Description: "desc", Enabled: true, AllowedMaxSeverity: 1},
		},
	}
	if err := validateSpec(spec); err == nil {
		t.Fatalf("expected error for empty custom constraint id")
	}
}

func TestParseSpec_RejectsUnknownFields(t *testing.T) {
	form := &multipart.Form{
		Value: map[string][]string{
			"spec": {`{"schema_version":1,"evaluation_name":"eval","policy":{"reveal":{"max_severity":true,"commitment":true}},"constraints":[],"custom_constraints":[],"extra":true}`},
		},
	}
	if _, err := parseSpec(form); err == nil {
		t.Fatalf("expected error for unknown spec fields")
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
