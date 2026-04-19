package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type countingCategoryStore struct {
	setCalls int
}

func (c *countingCategoryStore) GetCategories(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func (c *countingCategoryStore) SetCategories(_ context.Context, _ string, _ []string) error {
	c.setCalls++
	return nil
}

func TestHandlePutDispatchCategories_tooMany(t *testing.T) {
	store := &countingCategoryStore{}
	s := &Server{Categories: store}
	rr := httptest.NewRecorder()
	cats := make([]string, 11)
	for i := range cats {
		cats[i] = string(rune('a' + i))
	}
	body := map[string][]string{"categories": cats}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/settings/dispatch-categories", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "u1")
	s.HandlePutDispatchCategories(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d: %s", rr.Code, rr.Body.String())
	}
	if store.setCalls != 0 {
		t.Fatalf("SetCategories should not run: %d", store.setCalls)
	}
}
