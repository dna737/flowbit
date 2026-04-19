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

var testAllowedJobTypes = []string{"echo", "email", "image_resize", "url_scrape", "fail"}

func TestDispatch_email(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{
		response: `{"job_type":"email","parameters":{"to":"bob@example.com","subject":"Hello"}}`,
	}}
	result, err := d.Dispatch(context.Background(), "send email to bob@example.com", nil, testAllowedJobTypes)
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
	result, err := d.Dispatch(context.Background(), "resize the image to 800x600", nil, testAllowedJobTypes)
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
	result, err := d.Dispatch(context.Background(), "scrape https://example.com for prices", nil, testAllowedJobTypes)
	if err != nil {
		t.Fatal(err)
	}
	if result.JobType != "url_scrape" {
		t.Fatalf("job_type: %s", result.JobType)
	}
}

func TestDispatch_malformedJSON(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{response: `not json`}}
	_, err := d.Dispatch(context.Background(), "anything", nil, testAllowedJobTypes)
	if !errors.Is(err, ErrAIParseFailed) {
		t.Fatalf("want ErrAIParseFailed, got %v", err)
	}
}

func TestDispatch_missingJobType(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{response: `{"parameters":{}}`}}
	_, err := d.Dispatch(context.Background(), "anything", nil, testAllowedJobTypes)
	if !errors.Is(err, ErrAIParseFailed) {
		t.Fatalf("want ErrAIParseFailed, got %v", err)
	}
}

func TestDispatch_invalidJobType(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{response: `{"job_type":"unknown","parameters":{}}`}}
	_, err := d.Dispatch(context.Background(), "anything", nil, testAllowedJobTypes)
	if !errors.Is(err, ErrAIParseFailed) {
		t.Fatalf("want ErrAIParseFailed, got %v", err)
	}
}

func TestDispatch_emptyJobTypes(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{response: `{"job_type":"email","parameters":{}}`}}
	_, err := d.Dispatch(context.Background(), "anything", nil, nil)
	if !errors.Is(err, ErrAIParseFailed) {
		t.Fatalf("want ErrAIParseFailed, got %v", err)
	}
}

func TestDispatch_generatorError(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{err: errors.New("api down")}}
	_, err := d.Dispatch(context.Background(), "anything", nil, testAllowedJobTypes)
	if !errors.Is(err, ErrAIParseFailed) {
		t.Fatalf("want ErrAIParseFailed, got %v", err)
	}
}

func TestDispatch_categoriesRequired_missing(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{
		response: `{"job_type":"email","parameters":{"to":"bob@example.com","subject":"Hello"}}`,
	}}
	_, err := d.Dispatch(context.Background(), "send email", []string{"react"}, testAllowedJobTypes)
	if err == nil {
		t.Fatal("expected error when category missing")
	}
	if !errors.Is(err, ErrAIParseFailed) {
		t.Fatalf("want ErrAIParseFailed, got %v", err)
	}
}

func TestDispatch_categoriesRequired_ok(t *testing.T) {
	d := &GeminiDispatcher{gen: &fakeGenerator{
		response: `{"job_type":"email","parameters":{"to":"bob@example.com","subject":"Hello","category":"react"}}`,
	}}
	result, err := d.Dispatch(context.Background(), "send email", []string{"react"}, testAllowedJobTypes)
	if err != nil {
		t.Fatal(err)
	}
	if result.Parameters["category"] != "react" {
		t.Fatalf("category: %v", result.Parameters["category"])
	}
}
