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

func buildSystemPrompt(jobTypes []string) (string, error) {
	if len(jobTypes) == 0 {
		return "", errors.New("jobTypes required for system prompt")
	}
	typeList := strings.Join(jobTypes, ", ")
	return fmt.Sprintf(`You are a job dispatcher. Given a plain-English request, respond with ONLY valid JSON:
{"job_type": "<one of: %s>", "parameters": {<relevant key-value pairs>}}
The job_type MUST be exactly one of the listed labels (verbatim spelling). Pick the best fit.
No explanation. No markdown. JSON only.`, typeList), nil
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

// extractJSONObject pulls the first top-level {...} object out of raw Gemini
// text. Even with ResponseMIMEType "application/json", Gemini occasionally
// decorates the response with ```json fences, leading prose, or trailing text
// after the closing brace — any of which make json.Unmarshal fail with
// "invalid character X after top-level value". This walks the first balanced
// object (string- and escape-aware) and returns just that substring.
func extractJSONObject(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	// Strip common markdown code fences Gemini sometimes emits around JSON.
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```JSON")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	start := strings.IndexByte(s, '{')
	if start < 0 {
		return "", errors.New("no JSON object found in response")
	}

	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1], nil
			}
		}
	}
	return "", errors.New("unterminated JSON object in response")
}

// Dispatch calls Gemini with the prompt and parses the JSON response into a DispatchResult.
// jobTypes is the user's configured list of allowed job_type labels (the single source of truth).
func (d *GeminiDispatcher) Dispatch(ctx context.Context, prompt string, jobTypes []string) (DispatchResult, error) {
	if len(jobTypes) == 0 {
		return DispatchResult{}, fmt.Errorf("%w: no allowed job types", ErrAIParseFailed)
	}
	system, err := buildSystemPrompt(jobTypes)
	if err != nil {
		return DispatchResult{}, fmt.Errorf("%w: %v", ErrAIParseFailed, err)
	}
	raw, err := d.gen.GenerateContent(ctx, system, prompt)
	if err != nil {
		return DispatchResult{}, fmt.Errorf("%w: %v", ErrAIParseFailed, err)
	}
	payload, err := extractJSONObject(raw)
	if err != nil {
		return DispatchResult{}, fmt.Errorf("%w: %v", ErrAIParseFailed, err)
	}
	var result DispatchResult
	if err := json.Unmarshal([]byte(payload), &result); err != nil {
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
	return result, nil
}
