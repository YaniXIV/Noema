package evaluate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type CachedGeminiOutput struct {
	Model         string       `json:"model"`
	PromptVersion string       `json:"prompt_version"`
	Output        EvalOutput   `json:"output"`
	RawText       string       `json:"raw_text"`
	Usage         *GeminiUsage `json:"usage,omitempty"`
	CachedAt      string       `json:"cached_at"`
}

type GeminiUsage struct {
	PromptTokens     int32 `json:"prompt_tokens"`
	CandidateTokens  int32 `json:"candidate_tokens"`
	TotalTokens      int32 `json:"total_tokens"`
	CachedTokenCount int32 `json:"cached_token_count"`
}

func cacheKey(dataset []byte, spec []byte, model string, sampleLimit int) string {
	h := sha256.New()
	h.Write(dataset)
	h.Write(spec)
	h.Write([]byte(model))
	h.Write([]byte(promptVersion))
	h.Write([]byte(fmt.Sprintf("sample:%d", sampleLimit)))
	return hex.EncodeToString(h.Sum(nil))
}

func cachePath(runsDir, key string) string {
	return filepath.Join(runsDir, "cache", key, "gemini_output.json")
}

func loadCache(runsDir, key string) (*CachedGeminiOutput, error) {
	path := cachePath(runsDir, key)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out CachedGeminiOutput
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func saveCache(runsDir, key string, out CachedGeminiOutput) error {
	path := cachePath(runsDir, key)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if out.CachedAt == "" {
		out.CachedAt = time.Now().UTC().Format(time.RFC3339)
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
