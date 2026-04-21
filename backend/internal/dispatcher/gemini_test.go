package dispatcher

import (
	"context"
	"errors"
	"testing"
)

type fakeGenerator struct {
	response string
	err      error
}

func (f *fakeGenerator) GenerateContent(_ context.Context, _, _ string) (string, error) {
	return f.response, f.err
}

var testAllowedJobTypes = []string{"general", "email", "image_resize", "url_scrape", "fail"}

func TestDispatch_email(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{
		response: `{"job_type":"email","parameters":{"to":"bob@example.com","subject":"Hello"}}`,
	}}
	result, err := d.Dispatch(context.Background(), "send email to bob@example.com", testAllowedJobTypes)
	if err != nil {
		t.Fatal(err)
	}
	if result.JobType != "email" {
		t.Fatalf("job_type: %s", result.JobType)
	}
	if result.Parameters["to"] != "bob@example.com" {
		t.Fatalf("params: %v", result.Parameters)
	}
}

func TestDispatch_imageResize(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{
		response: `{"job_type":"image_resize","parameters":{"width":800,"height":600}}`,
	}}
	result, err := d.Dispatch(context.Background(), "resize the image to 800x600", testAllowedJobTypes)
	if err != nil {
		t.Fatal(err)
	}
	if result.JobType != "image_resize" {
		t.Fatalf("job_type: %s", result.JobType)
	}
}

func TestDispatch_urlScrape(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{
		response: `{"job_type":"url_scrape","parameters":{"url":"https://example.com"}}`,
	}}
	result, err := d.Dispatch(context.Background(), "scrape https://example.com for prices", testAllowedJobTypes)
	if err != nil {
		t.Fatal(err)
	}
	if result.JobType != "url_scrape" {
		t.Fatalf("job_type: %s", result.JobType)
	}
}

// When the user's list is just ["general"], the AI must pick that.
func TestDispatch_fallsBackToGeneric(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{
		response: `{"job_type":"general","parameters":{"message":"read a book on Sunday"}}`,
	}}
	result, err := d.Dispatch(context.Background(), "read a book on Sunday", []string{"general"})
	if err != nil {
		t.Fatal(err)
	}
	if result.JobType != "general" {
		t.Fatalf("job_type: %s", result.JobType)
	}
}

func TestDispatch_malformedJSON(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{response: `not json`}}
	_, err := d.Dispatch(context.Background(), "anything", testAllowedJobTypes)
	if !errors.Is(err, ErrAIParseFailed) {
		t.Fatalf("want ErrAIParseFailed, got %v", err)
	}
}

func TestDispatch_missingJobType(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{response: `{"parameters":{}}`}}
	_, err := d.Dispatch(context.Background(), "anything", testAllowedJobTypes)
	if !errors.Is(err, ErrAIParseFailed) {
		t.Fatalf("want ErrAIParseFailed, got %v", err)
	}
}

func TestDispatch_invalidJobType(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{response: `{"job_type":"unknown","parameters":{}}`}}
	_, err := d.Dispatch(context.Background(), "anything", testAllowedJobTypes)
	if !errors.Is(err, ErrAIParseFailed) {
		t.Fatalf("want ErrAIParseFailed, got %v", err)
	}
}

func TestDispatch_emptyJobTypes(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{response: `{"job_type":"email","parameters":{}}`}}
	_, err := d.Dispatch(context.Background(), "anything", nil)
	if !errors.Is(err, ErrAIParseFailed) {
		t.Fatalf("want ErrAIParseFailed, got %v", err)
	}
}

func TestDispatch_generatorError(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{err: errors.New("api down")}}
	_, err := d.Dispatch(context.Background(), "anything", testAllowedJobTypes)
	if !errors.Is(err, ErrAIParseFailed) {
		t.Fatalf("want ErrAIParseFailed, got %v", err)
	}
}

// Gemini sometimes appends prose/label text after the JSON object even with
// ResponseMIMEType set to application/json. json.Unmarshal rejects that with
// "invalid character X after top-level value" — we must tolerate it by
// extracting only the first balanced object.
func TestDispatch_trailingProseAfterJSON(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{
		response: `{"job_type":"general","parameters":{"message":"hi"}}  pastime is a good fit`,
	}}
	result, err := d.Dispatch(context.Background(), "anything", testAllowedJobTypes)
	if err != nil {
		t.Fatalf("want tolerant parse, got %v", err)
	}
	if result.JobType != "general" {
		t.Fatalf("job_type: %s", result.JobType)
	}
	if result.Parameters["message"] != "hi" {
		t.Fatalf("params: %v", result.Parameters)
	}
}

func TestDispatch_markdownCodeFence(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{
		response: "```json\n{\"job_type\":\"email\",\"parameters\":{\"to\":\"a@b.com\"}}\n```",
	}}
	result, err := d.Dispatch(context.Background(), "anything", testAllowedJobTypes)
	if err != nil {
		t.Fatalf("want tolerant parse, got %v", err)
	}
	if result.JobType != "email" {
		t.Fatalf("job_type: %s", result.JobType)
	}
}

func TestExtractJSONObject_bracesInsideStrings(t *testing.T) {
	in := `Some preamble. {"job_type":"general","parameters":{"msg":"a } brace \"quoted\" and {nested}"}} trailing`
	got, err := extractJSONObject(in)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"job_type":"general","parameters":{"msg":"a } brace \"quoted\" and {nested}"}}`
	if got != want {
		t.Fatalf("extracted:\n  got  %q\n  want %q", got, want)
	}
}

func TestExtractJSONObject_noObject(t *testing.T) {
	if _, err := extractJSONObject("no braces here"); err == nil {
		t.Fatal("want error for response without an object")
	}
}

func TestExtractJSONObject_unterminated(t *testing.T) {
	if _, err := extractJSONObject(`{"job_type":"general","parameters":{`); err == nil {
		t.Fatal("want error for unterminated object")
	}
}
