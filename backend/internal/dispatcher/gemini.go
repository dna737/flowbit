package dispatcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
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

// NewGeminiDispatcher creates a GeminiDispatcher using the given API key and model name.
func NewGeminiDispatcher(apiKey, model string) (*GeminiDispatcher, error) {
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}
	m := client.GenerativeModel(model)
	m.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}
	m.ResponseMIMEType = "application/json"
	return &GeminiDispatcher{gen: &geminiModel{model: m}}, nil
}

// geminiModel wraps genai.GenerativeModel as textGenerator.
type geminiModel struct {
	model *genai.GenerativeModel
}

func (g *geminiModel) GenerateContent(ctx context.Context, prompt string) (string, error) {
	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}
	if len(resp.Candidates) == 0 ||
		resp.Candidates[0].Content == nil ||
		len(resp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty response from gemini")
	}
	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", errors.New("unexpected part type from gemini")
	}
	return string(text), nil
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
