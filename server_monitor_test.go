package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildListenAddress(t *testing.T) {
	address, err := buildListenAddress("127.0.0.1", 9000)
	if err != nil {
		t.Fatalf("buildListenAddress returned an error: %v", err)
	}
	if address != "127.0.0.1:9000" {
		t.Fatalf("unexpected address: %s", address)
	}
	if _, err := buildListenAddress("127.0.0.1", 0); err == nil {
		t.Fatal("expected an invalid port error")
	}
}

func TestAPIHandler(t *testing.T) {
	handler := newHandler(func() SystemInfo {
		return SystemInfo{Timestamp: "2026-08-04 10:00:00"}
	})

	request := httptest.NewRequest(http.MethodGet, "/api", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("API responses must not be cached")
	}
	if body := response.Body.String(); body == "" || body == "{}\n" {
		t.Fatalf("unexpected API body: %q", body)
	}
}

func TestHandlerRejectsUnknownPathsAndMethods(t *testing.T) {
	handler := newHandler(func() SystemInfo { return SystemInfo{} })

	unknown := httptest.NewRecorder()
	handler.ServeHTTP(unknown, httptest.NewRequest(http.MethodGet, "/missing", nil))
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("unexpected unknown-path status: %d", unknown.Code)
	}

	post := httptest.NewRecorder()
	handler.ServeHTTP(post, httptest.NewRequest(http.MethodPost, "/api", nil))
	if post.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected method status: %d", post.Code)
	}
}
