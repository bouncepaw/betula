// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package wwwgw

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestFindTitle(t *testing.T) {
	www := New()

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "title in head",
			html:     `<html><head><title>Test Title</title></head><body></body></html>`,
			expected: "Test Title",
		},
		{
			name:     "title in root",
			html:     `<html><title>Root Title</title><head></head><body></body></html>`,
			expected: "Root Title",
		},
		{
			name:     "title with whitespace",
			html:     `<html><head><title>  Title with spaces  </title></head><body></body></html>`,
			expected: "Title with spaces",
		},
		{
			name:     "no title element",
			html:     `<html><head></head><body></body></html>`,
			expected: "",
		},
		{
			name:     "empty title",
			html:     `<html><head><title></title></head><body></body></html>`,
			expected: "",
		},
		{
			name:     "title after body should not be found",
			html:     `<html><head></head><body><title>Body Title</title></body></html>`,
			expected: "",
		},
		{
			name:     "complex html structure",
			html:     `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>Complex Page</title></head><body><div>Content</div></body></html>`,
			expected: "Complex Page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := www.findTitle(doc)
			if result != tt.expected {
				t.Errorf("Expected title '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestTitleOfPageWithSmallLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<html><head><meta charset="utf-8"><title>Late Title</title></head><body></body></html>`))
	}))
	defer server.Close()
	title, err := NewWithLimit(10).TitleOfPage(server.URL)
	if err != nil {
		t.Fatalf("TitleOfPage failed: %v", err)
	}
	expected := "Late Title"
	if title != expected {
		t.Errorf("Expected title '%s', got '%s'", expected, title)
	}
}

func TestTitleOfPageWithDefaultLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<html><head><meta charset="utf-8"><title>Late Title</title></head><body></body></html>`))
	}))
	defer server.Close()
	title, err := New().TitleOfPage(server.URL)
	if err != nil {
		t.Fatalf("TitleOfPage failed: %v", err)
	}
	expected := "Late Title"
	if title != expected {
		t.Errorf("Expected title '%s', got '%s'", expected, title)
	}
}
