package dispatcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"google.golang.org/genai"
)

const systemPrompt = `You are a job dispatcher. Given a plain-English request, respond with ONLY valid JSON:
{"job_type": "<one of: echo, email, image_resize, url_scrape, fail>", "parameters": {<relevant key-value pairs>}}
No explanation. No markdown. JSON only.`

// textGenerator abstracts the Gemini model call (injectable for tests).
type textGenerator interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}

// GeminiDispatcher implements Dispatcher using the Gemini API.
type GeminiDispatcher struct {
	gen textGenerator
}

// NewGeminiDispatcher creates a GeminiDispatcher using the given API key, primary model,
// and an optional ordered list of fallback model names tried when the primary returns a
// retryable error (503 UNAVAILABLE, 429 RESOURCE_EXHAUSTED).
func NewGeminiDispatcher(apiKey, model string, fallbacks ...string) (*GeminiDispatcher, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}
	models := append([]string{model}, fallbacks...)
	return &GeminiDispatcher{gen: &geminiModel{client: client, models: models}}, nil
}

// geminiModel wraps genai.Client as textGenerator. It tries each model in order,
// falling back on retryable upstream errors.
type geminiModel struct {
	client *genai.Client
	models []string
}

func (g *geminiModel) GenerateContent(ctx context.Context, prompt string) (string, error) {
	cfg := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
		ResponseMIMEType: "application/json",
	}
	var lastErr error
	for i, m := range g.models {
		resp, err := g.client.Models.GenerateContent(ctx, m, genai.Text(prompt), cfg)
		if err != nil {
			lastErr = err
			if i < len(g.models)-1 && isRetryable(err) {
				log.Printf("gemini model %q unavailable (%v); falling back to %q", m, err, g.models[i+1])
				continue
			}
			return "", err
		}
		text := resp.Text()
		if text == "" {
			return "", errors.New("empty response from gemini")
		}
		return text, nil
	}
	return "", lastErr
}

// isRetryable reports whether err is a transient upstream condition worth retrying
// against a different model (overload / quota).
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	// 5xx gateway/server errors + rate-limit/quota errors are all worth trying
	// another model for. 4xx other than 429 (bad request, auth, not found) are not.
	return strings.Contains(s, "500") ||
		strings.Contains(s, "INTERNAL") ||
		strings.Contains(s, "502") ||
		strings.Contains(s, "503") ||
		strings.Contains(s, "UNAVAILABLE") ||
		strings.Contains(s, "504") ||
		strings.Contains(s, "DEADLINE_EXCEEDED") ||
		strings.Contains(s, "429") ||
		strings.Contains(s, "RESOURCE_EXHAUSTED") ||
		strings.Contains(s, "overloaded")
}

// Dispatch calls Gemini with the prompt and parses the JSON response into a DispatchResult.
func (d *GeminiDispatcher) Dispatch(ctx context.Context, prompt string) (DispatchResult, error) {
	raw, err := d.gen.GenerateContent(ctx, prompt)
	if err != nil {
		return DispatchResult{}, fmt.Errorf("%w: %v", ErrAIParseFailed, err)
	}
	var result DispatchResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return DispatchResult{}, fmt.Errorf("%w: %v", ErrAIParseFailed, err)
	}
	if result.JobType == "" {
		return DispatchResult{}, fmt.Errorf("%w: missing job_type", ErrAIParseFailed)
	}
	return result, nil
}
