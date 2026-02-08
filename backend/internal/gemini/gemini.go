package gemini

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/genai"
)

const defaultModel = "gemini-3-pro"

type ImageInput struct {
	MIMEType string
	Data     []byte
}

type EvalRequest struct {
	SystemPrompt    string
	UserPrompt      string
	ResponseSchema  any
	Images          []ImageInput
	Temperature     float32
	MaxOutputTokens int32
}

type Usage struct {
	PromptTokens     int32 `json:"prompt_tokens"`
	CandidateTokens  int32 `json:"candidate_tokens"`
	TotalTokens      int32 `json:"total_tokens"`
	CachedTokenCount int32 `json:"cached_token_count"`
}

type EvalResponse struct {
	Text  string
	Usage *Usage
	Model string
}

func modelName() string {
	if m := strings.TrimSpace(os.Getenv("GEMINI_MODEL")); m != "" {
		return m
	}
	return defaultModel
}

// ModelName returns the resolved Gemini model name.
func ModelName() string {
	return modelName()
}

func newClient(ctx context.Context) (*genai.Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not set")
	}
	return genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
}

func buildConfig(req EvalRequest) *genai.GenerateContentConfig {
	cfg := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: req.SystemPrompt}},
		},
		Temperature:     &req.Temperature,
		MaxOutputTokens: req.MaxOutputTokens,
	}
	if req.ResponseSchema != nil {
		cfg.ResponseMIMEType = "application/json"
		cfg.ResponseJsonSchema = req.ResponseSchema
	}
	return cfg
}

func buildContents(req EvalRequest) []*genai.Content {
	parts := []*genai.Part{{Text: req.UserPrompt}}
	for _, img := range req.Images {
		if len(img.Data) == 0 {
			continue
		}
		parts = append(parts, genai.NewPartFromBytes(img.Data, img.MIMEType))
	}
	return []*genai.Content{{
		Role:  genai.RoleUser,
		Parts: parts,
	}}
}

func extractUsage(meta *genai.GenerateContentResponseUsageMetadata) *Usage {
	if meta == nil {
		return nil
	}
	return &Usage{
		PromptTokens:     meta.PromptTokenCount,
		CandidateTokens:  meta.CandidatesTokenCount,
		TotalTokens:      meta.TotalTokenCount,
		CachedTokenCount: meta.CachedContentTokenCount,
	}
}

// SendText sends the given text to Gemini, prints the response to the console, and returns it.
// Uses GEMINI_API_KEY from the environment (e.g. loaded from .env).
func SendText(ctx context.Context, text string) (string, error) {
	client, err := newClient(ctx)
	if err != nil {
		return "", err
	}

	model := modelName()
	result, err := client.Models.GenerateContent(ctx, model, genai.Text(text), nil)
	if err != nil {
		return "", fmt.Errorf("generate content: %w", err)
	}

	out := result.Text()
	log.Println("[gemini]", out)
	return out, nil
}

// Evaluate runs a structured evaluation prompt and returns the raw text response.
func Evaluate(ctx context.Context, req EvalRequest) (EvalResponse, error) {
	client, err := newClient(ctx)
	if err != nil {
		return EvalResponse{}, err
	}
	model := modelName()
	result, err := client.Models.GenerateContent(ctx, model, buildContents(req), buildConfig(req))
	if err != nil {
		return EvalResponse{}, fmt.Errorf("generate content: %w", err)
	}
	return EvalResponse{
		Text:  result.Text(),
		Usage: extractUsage(result.UsageMetadata),
		Model: model,
	}, nil
}

// EvaluateStream runs a streaming evaluation prompt. onChunk is called with text deltas.
func EvaluateStream(ctx context.Context, req EvalRequest, onChunk func(string)) (EvalResponse, error) {
	client, err := newClient(ctx)
	if err != nil {
		return EvalResponse{}, err
	}
	model := modelName()
	var sb strings.Builder
	var usage *Usage
	for result, err := range client.Models.GenerateContentStream(ctx, model, buildContents(req), buildConfig(req)) {
		if err != nil {
			return EvalResponse{}, fmt.Errorf("generate content stream: %w", err)
		}
		if usage == nil {
			usage = extractUsage(result.UsageMetadata)
		}
		chunk := result.Text()
		if chunk == "" {
			continue
		}
		sb.WriteString(chunk)
		if onChunk != nil {
			onChunk(chunk)
		}
	}
	return EvalResponse{
		Text:  sb.String(),
		Usage: usage,
		Model: model,
	}, nil
}
