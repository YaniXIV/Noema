package evaluate

import (
	"context"
	"log"
	"mime/multipart"
	"time"

	"noema/internal/config"
	"noema/internal/gemini"
)

const geminiEvalTimeout = 45 * time.Second

func resolveEvalOutput(ctx context.Context, form *multipart.Form, enabled map[string]ConstraintRule, runsDir string, spec Spec, datasetFile *multipart.FileHeader, imageFiles []*multipart.FileHeader) (EvalOutput, error) {
	if out, provided, err := parseEvalOutputProvided(form, enabled); err != nil {
		return EvalOutput{}, err
	} else if provided {
		return out, nil
	}
	return evalWithGemini(ctx, enabled, runsDir, spec, datasetFile, imageFiles), nil
}

func evalWithGemini(ctx context.Context, enabled map[string]ConstraintRule, runsDir string, spec Spec, datasetFile *multipart.FileHeader, imageFiles []*multipart.FileHeader) EvalOutput {
	if config.GeminiAPIKey() == "" {
		return stubEvalOutput(enabled)
	}

	rawDataset, ds, err := readDatasetFile(datasetFile)
	if err != nil {
		log.Printf("gemini read dataset: %v", err)
		return stubEvalOutput(enabled)
	}

	sampleLimit := config.SampleItemsLimit()
	model := gemini.ModelName()
	specJSON, err := jsonBytes(spec)
	if err != nil {
		log.Printf("gemini marshal spec: %v", err)
		return stubEvalOutput(enabled)
	}
	key := cacheKey(rawDataset, specJSON, model, sampleLimit)
	if cached, err := loadCache(runsDir, key); err == nil {
		if err := validateEvalOutput(cached.Output, enabled); err == nil {
			return cached.Output
		}
	}

	sampled := sampleDataset(ds, sampleLimit)
	sampledJSON, err := marshalSampledDataset(sampled)
	if err != nil {
		log.Printf("gemini marshal dataset: %v", err)
		return stubEvalOutput(enabled)
	}

	images, err := readImages(imageFiles)
	if err != nil {
		log.Printf("gemini read images: %v", err)
		return stubEvalOutput(enabled)
	}

	prompt := buildUserPrompt(spec, sampledJSON, images)
	req := gemini.EvalRequest{
		SystemPrompt:    buildSystemPrompt(),
		UserPrompt:      prompt,
		ResponseSchema:  evalResponseSchema(),
		Temperature:     0,
		MaxOutputTokens: 2048,
		Images:          toGeminiImages(images),
	}

	ctx, cancel := withGeminiTimeout(ctx)
	defer cancel()
	resp, err := gemini.Evaluate(ctx, req)
	if err != nil {
		log.Printf("gemini evaluate: %v", err)
		return stubEvalOutput(enabled)
	}

	out, err := parseEvalOutput(resp.Text)
	if err != nil {
		log.Printf("gemini parse output: %v", err)
		return stubEvalOutput(enabled)
	}
	if err := validateEvalOutput(out, enabled); err != nil {
		log.Printf("gemini validate output: %v", err)
		return stubEvalOutput(enabled)
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
