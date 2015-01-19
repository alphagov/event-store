package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNonJsonBodyReturnsBadRequest(t *testing.T) {
	request, _ := http.NewRequest("POST", "/r", bytes.NewBufferString("Hello"))
	response := httptest.NewRecorder()

	ReportHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("Expected bad request status code, received %v", response.Code)
	}
}

func TestDocumentUriMustBeOnGovuk(t *testing.T) {
	payload := `{"csp-report": {
		"document-uri": "https://www.example.com/",
		"blocked-uri": "https://evil.example.com/",
		"violated-directive": "directive",
		"original-policy": "policy"
	}}`

	request, _ := http.NewRequest("POST", "/r", bytes.NewBufferString(payload))
	response := httptest.NewRecorder()

	ReportHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("Expected bad request status code, received %v", response.Code)
	}
}
