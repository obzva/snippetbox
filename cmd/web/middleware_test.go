package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/obzva/snippetbox/internal/assert"
)

func TestSetCommonHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	stubHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	setCommonHeaders(stubHandler).ServeHTTP(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	// check response status code
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// check if response headers are set properly
	tests := []struct {
		header string
		want   string
	}{
		{
			header: "Content-Security-Policy",
			want:   "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com",
		},
		{
			header: "Referrer-Policy",
			want:   "origin-when-cross-origin",
		},
		{
			header: "X-Content-Type-Options",
			want:   "nosniff",
		},
		{
			header: "X-Frame-Options",
			want:   "deny",
		},
		{
			header: "X-XSS-Protection",
			want:   "0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			assert.Equal(t, res.Header.Get(tt.header), tt.want)
		})
	}

	// check response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)
	assert.Equal(t, string(body), "OK")
}
