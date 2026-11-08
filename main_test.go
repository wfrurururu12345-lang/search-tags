package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Платье ", "платье"},
		{"  КРАСНОЕ   ПЛАТЬЕ  ", "красное платье"},
		{"Джинсы", "джинсы"},
	}

	for _, tc := range cases {
		result := normalize(tc.input)

		if result != tc.expected {
			t.Errorf(
				"normalize(%q) = %q, expected %q",
				tc.input,
				result,
				tc.expected,
			)
		}
	}
}

func TestTagsHandlerEmptyQuery(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/tags?q=",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler := tagsHandler(nil)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(
			"Expected status 400, got %d",
			recorder.Code,
		)
	}
}
