package evaluate

import (
	"context"
	"log"
	"mime/multipart"
	"os"
	"time"

	"noema/internal/config"
	"noema/internal/gemini"
)

const geminiEvalTimeout = 45 * time.Second

func resolveEvaluationResult(ctx context.Context, form *multipart.Form, cfg PolicyConfig, runsDir string, datasetFile *multipart.FileHeader, imageFiles []*multipart.FileHeader) (EvaluationResult, error) {
	if out, provided, err := parseEvaluationResultProvided(form, cfg); err != nil {
		return EvaluationResult{}, err
	} else if provided {
		return out, nil
	}
	return evalWithGemini(ctx, cfg, runsDir, datasetFile, imageFiles), nil
}

func evalWithGemini(ctx context.Context, cfg PolicyConfig, runsDir string, datasetFile *multipart.FileHeader, imageFiles []*multipart.FileHeader) EvaluationResult {
	if config.GeminiAPIKey() == "" {
		log.Printf("gemini disabled: missing GEMINI_API_KEY")
		return stubEvaluationResult(cfg)
	}

	rawDataset, err := readDatasetBytes(datasetFile)
	if err != nil {
		log.Printf("gemini fallback: read dataset failed: %v", err)
		return stubEvaluationResult(cfg)
	}

	sampleLimit := config.SampleItemsLimit()
	model := gemini.ModelName()
	policyJSON, err := jsonBytes(cfg)
	if err != nil {
		log.Printf("gemini fallback: marshal policy_config failed: %v", err)
		return stubEvaluationResult(cfg)
	}
	log.Printf("gemini request: model=%s sample_limit=%d", model, sampleLimit)
	key := cacheKey(rawDataset, policyJSON, model, sampleLimit)
	if cached, err := loadCache(runsDir, key); err == nil {
		if err := validateEvaluationResult(cached.Output, cfg); err == nil {
			log.Printf("gemini cache hit: %s", key)
			return cached.Output
		}
		_ = os.Remove(cachePath(runsDir, key))
	} else if !os.IsNotExist(err) {
		_ = os.Remove(cachePath(runsDir, key))
	}

	var sampledJSON []byte
	if ds, err := parseDatasetSchema(rawDataset); err == nil {
		sampled := sampleDataset(ds, sampleLimit)
		sampledJSON, err = marshalSampledDataset(sampled)
		if err != nil {
			log.Printf("gemini marshal dataset: %v", err)
			return stubEvaluationResult(cfg)
		}
	} else {
		sampledJSON = rawDataset
	}

	images, err := readImages(imageFiles)
	if err != nil {
		log.Printf("gemini fallback: read images failed: %v", err)
		return stubEvaluationResult(cfg)
	}

	prompt := buildUserPrompt(cfg, sampledJSON, images)
	req := gemini.EvalRequest{
		SystemPrompt:    buildSystemPrompt(),
		UserPrompt:      prompt,
		ResponseSchema:  evalResponseSchema(),
		Temperature:     0,
		MaxOutputTokens: 2048,
		Images:          toGeminiImages(images),
	}
	log.Printf("gemini system prompt: %s", req.SystemPrompt)
	log.Printf("gemini user prompt: %s", req.UserPrompt)

	ctx, cancel := withGeminiTimeout(ctx)
	defer cancel()
	log.Printf("gemini call: sending request")
	resp, err := gemini.Evaluate(ctx, req)
	if err != nil {
		log.Printf("gemini fallback: evaluate failed: %v", err)
		return stubEvaluationResult(cfg)
	}
	log.Printf("gemini output: %s", resp.Text)

	out, err := parseEvaluationResult(resp.Text)
	if err != nil {
		log.Printf("gemini fallback: parse output failed: %v", err)
		return stubEvaluationResult(cfg)
	}
	if err := validateEvaluationResult(out, cfg); err != nil {
		log.Printf("gemini fallback: validate output failed: %v", err)
		return stubEvaluationResult(cfg)
	}

	cacheOut := CachedGeminiOutput{
		Model:         resp.Model,
		PromptVersion: promptVersion,
		Output:        out,
		RawText:       resp.Text,
		Usage:         toGeminiUsage(resp.Usage),
	}
	if err := saveCache(runsDir, key, cacheOut); err != nil {
		log.Printf("gemini cache save: %v", err)
	}

	return out
}

func withGeminiTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, geminiEvalTimeout)
}

func toGeminiImages(images []ImageInfo) []gemini.ImageInput {
	if len(images) == 0 {
		return nil
	}
	out := make([]gemini.ImageInput, 0, len(images))
	for _, img := range images {
		out = append(out, gemini.ImageInput{
			MIMEType: img.MIMEType,
			Data:     img.Data,
		})
	}
	return out
}

func toGeminiUsage(usage *gemini.Usage) *GeminiUsage {
	if usage == nil {
		return nil
	}
	return &GeminiUsage{
		PromptTokens:     usage.PromptTokens,
		CandidateTokens:  usage.CandidateTokens,
		TotalTokens:      usage.TotalTokens,
		CachedTokenCount: usage.CachedTokenCount,
	}
}
