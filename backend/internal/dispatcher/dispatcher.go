package dispatcher

import (
	"context"
	"errors"
)

// ErrAIParseFailed is returned when the AI response cannot be parsed as a valid DispatchResult.
var ErrAIParseFailed = errors.New("ai response could not be parsed as job payload")

// DispatchResult is the structured output from the AI dispatcher.
type DispatchResult struct {
	JobType    string         `json:"job_type"`
	Parameters map[string]any `json:"parameters"`
}

// Dispatcher parses a plain-English prompt into a structured job payload.
// jobTypes is the user's single list of allowed job_type labels — the AI must
// pick exactly one. Callers are expected to load this list from the user's
// Settings (users.dispatch_categories) before calling Dispatch.
type Dispatcher interface {
	Dispatch(ctx context.Context, prompt string, jobTypes []string) (DispatchResult, error)
}
