package verify

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"noema/internal/config"

	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Error string `json:"error"`
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/verify", Handler())
	return r
}

func TestVerifyHandlerLegacyFallback(t *testing.T) {
	r := setupRouter()
	body := `{"run_id":"abc","proof_b64":"` + base64.StdEncoding.EncodeToString([]byte("stub_proof_abc")) + `","public_inputs_b64":"` + base64.StdEncoding.EncodeToString([]byte("stub_inputs_abc")) + `"}`

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/verify", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp VerifyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Verified {
		t.Fatalf("expected legacy proof to verify")
	}
	if resp.Message != "legacy stub verifier" {
		t.Fatalf("expected legacy message, got %q", resp.Message)
	}
}

func TestVerifyHandlerInvalidBase64Returns400(t *testing.T) {
	r := setupRouter()
	body := `{"run_id":"abc","proof_b64":"%%%","public_inputs_b64":"` + base64.StdEncoding.EncodeToString([]byte("stub_inputs_abc")) + `"}`

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/verify", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp errorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error != "invalid proof encoding" {
		t.Fatalf("expected invalid proof encoding, got %q", resp.Error)
	}
}

func TestVerifyHandlerLegacyMissingRunIDReturns400(t *testing.T) {
	r := setupRouter()
	body := `{"run_id":"","proof_b64":"` + base64.StdEncoding.EncodeToString([]byte("stub_proof_")) + `","public_inputs_b64":"` + base64.StdEncoding.EncodeToString([]byte("stub_inputs_")) + `"}`

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/verify", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp errorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error != "missing run_id" {
		t.Fatalf("expected missing run_id, got %q", resp.Error)
	}
}

func TestVerifyHandlerMissingRunIDReturns400(t *testing.T) {
	r := setupRouter()
	body := `{"run_id":"","proof_b64":"` + base64.StdEncoding.EncodeToString([]byte("proof")) + `","public_inputs_b64":"` + base64.StdEncoding.EncodeToString([]byte("inputs")) + `"}`

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/verify", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp errorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error != "missing run_id" {
		t.Fatalf("expected missing run_id, got %q", resp.Error)
	}
}

func TestVerifyHandlerWhitespaceRunIDReturns400(t *testing.T) {
	r := setupRouter()
	body := `{"run_id":"   ","proof_b64":"` + base64.StdEncoding.EncodeToString([]byte("proof")) + `","public_inputs_b64":"` + base64.StdEncoding.EncodeToString([]byte("inputs")) + `"}`

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/verify", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp errorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error != "missing run_id" {
		t.Fatalf("expected missing run_id, got %q", resp.Error)
	}
}

func TestVerifyHandlerBodyTooLargeReturns413(t *testing.T) {
	r := setupRouter()
	largeRunID := strings.Repeat("a", config.MaxVerifyBytes)
	body := `{"run_id":"` + largeRunID + `","proof_b64":"` + base64.StdEncoding.EncodeToString([]byte("proof")) + `","public_inputs_b64":"` + base64.StdEncoding.EncodeToString([]byte("inputs")) + `"}`

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/verify", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413, got %d", w.Code)
	}

	var resp errorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error != "request body too large" {
		t.Fatalf("expected request body too large, got %q", resp.Error)
	}
}
