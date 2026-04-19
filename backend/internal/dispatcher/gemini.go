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

// textGenerator abstracts the Gemini model call (injectable for tests).
type textGenerator interface {
	GenerateContent(ctx context.Context, systemInstruction, userPrompt string) (string, error)
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

func (g *geminiModel) GenerateContent(ctx context.Context, systemInstruction, userPrompt string) (string, error) {
	cfg := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemInstruction}},
		},
		ResponseMIMEType: "application/json",
	}
	var lastErr error
	for i, m := range g.models {
		resp, err := g.client.Models.GenerateContent(ctx, m, genai.Text(userPrompt), cfg)
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

func buildSystemPrompt(jobTypes []string, categories []string) (string, error) {
	if len(jobTypes) == 0 {
		return "", errors.New("jobTypes required for system prompt")
	}
	typeList := strings.Join(jobTypes, ", ")
	base := fmt.Sprintf(`You are a job dispatcher. Given a plain-English request, respond with ONLY valid JSON:
{"job_type": "<one of: %s>", "parameters": {<relevant key-value pairs>}}
No explanation. No markdown. JSON only.`, typeList)
	if len(categories) == 0 {
		return base, nil
	}
	var b strings.Builder
	b.WriteString(base)
	b.WriteString("\n\nThe user has configured the following category labels. You MUST set parameters[\"category\"] to exactly one string from this list (verbatim spelling as shown):\n")
	for _, c := range categories {
		b.WriteString("- ")
		b.WriteString(c)
		b.WriteByte('\n')
	}
	return b.String(), nil
}

func normalizeJobType(raw string, allowed []string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("%w: missing job_type", ErrAIParseFailed)
	}
	for _, a := range allowed {
		if strings.EqualFold(s, strings.TrimSpace(a)) {
			return a, nil
		}
	}
	return "", fmt.Errorf("%w: job_type must be one of the allowed values", ErrAIParseFailed)
}

func normalizeCategoryInParameters(params map[string]any, allowed []string) error {
	if len(allowed) == 0 {
		return nil
	}
	v, ok := params["category"]
	if !ok || v == nil {
		return fmt.Errorf("%w: missing parameters[\"category\"]", ErrAIParseFailed)
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w: parameters[\"category\"] must be a string", ErrAIParseFailed)
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("%w: parameters[\"category\"] must be non-empty", ErrAIParseFailed)
	}
	for _, a := range allowed {
		if strings.EqualFold(s, strings.TrimSpace(a)) {
			params["category"] = a
			return nil
		}
	}
	return fmt.Errorf("%w: parameters[\"category\"] must be one of the configured labels", ErrAIParseFailed)
}

// Dispatch calls Gemini with the prompt and parses the JSON response into a DispatchResult.
func (d *GeminiDispatcher) Dispatch(ctx context.Context, prompt string, categories []string, jobTypes []string) (DispatchResult, error) {
	if len(jobTypes) == 0 {
		return DispatchResult{}, fmt.Errorf("%w: no allowed job types", ErrAIParseFailed)
	}
	system, err := buildSystemPrompt(jobTypes, categories)
	if err != nil {
		return DispatchResult{}, fmt.Errorf("%w: %v", ErrAIParseFailed, err)
	}
	raw, err := d.gen.GenerateContent(ctx, system, prompt)
	if err != nil {
		return DispatchResult{}, fmt.Errorf("%w: %v", ErrAIParseFailed, err)
	}
	var result DispatchResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return DispatchResult{}, fmt.Errorf("%w: %v", ErrAIParseFailed, err)
	}
	canonical, err := normalizeJobType(result.JobType, jobTypes)
	if err != nil {
		return DispatchResult{}, err
	}
	result.JobType = canonical
	if result.Parameters == nil {
		result.Parameters = map[string]any{}
	}
	if err := normalizeCategoryInParameters(result.Parameters, categories); err != nil {
		return DispatchResult{}, err
	}
	return result, nil
}
