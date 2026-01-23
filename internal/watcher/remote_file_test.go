//go:build small

package watcher

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCreateRemoteFileRequest(t *testing.T) {
	testCases := []struct {
		name           string
		url            string
		headers        map[string]string
		expectError    bool
		expectedMethod string
	}{
		{
			name:           "valid request",
			url:            "http://example.com/test.txt",
			headers:        map[string]string{"User-Agent": "Wampa/1.0", "Accept": "text/plain"},
			expectError:    false,
			expectedMethod: http.MethodGet,
		},
		{
			name:           "invalid url",
			url:            "://invalid",
			headers:        nil,
			expectError:    true,
			expectedMethod: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			req, err := CreateRemoteFileRequest(ctx, tc.url, tc.headers)

			// Check error expectation
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got nil")
				return
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Skip further checks if we expected an error
			if tc.expectError {
				return
			}

			// Verify request properties
			if req.Method != tc.expectedMethod {
				t.Errorf("Expected method %s, got %s", tc.expectedMethod, req.Method)
			}

			if req.URL.String() != tc.url {
				t.Errorf("Expected URL %s, got %s", tc.url, req.URL.String())
			}

			// Check headers
			for key, value := range tc.headers {
				if req.Header.Get(key) != value {
					t.Errorf("Expected header %s=%s, got %s", key, value, req.Header.Get(key))
				}
			}
		})
	}
}

// createTestResponse creates a mock HTTP response for testing
func createTestResponse(statusCode int, body string, headers map[string]string) *http.Response {
	header := http.Header{}
	for key, value := range headers {
		header.Set(key, value)
	}

	return &http.Response{
		StatusCode:    statusCode,
		Body:          io.NopCloser(strings.NewReader(body)),
		Header:        header,
		ContentLength: int64(len(body)),
	}
}

func TestProcessRemoteFileResponse(t *testing.T) {
	const testURL = "http://example.com/test.txt"

	testCases := []struct {
		name         string
		response     *http.Response
		maxSize      int64
		expectError  bool
		expectedData string
		lastModified string // 各テストケースで個別に指定
	}{
		{
			name: "successful response",
			response: createTestResponse(
				http.StatusOK,
				"Hello, World!",
				map[string]string{
					"Content-Type":  "text/plain",
					"ETag":          "\"abc123\"",
					"Last-Modified": time.Now().AddDate(0, 0, -2).Format(time.RFC1123),
				},
			),
			maxSize:      1024,
			expectError:  false,
			expectedData: "Hello, World!",
			lastModified: time.Now().AddDate(0, 0, -2).Format(time.RFC1123),
		},
		{
			name: "error status code",
			response: createTestResponse(
				http.StatusNotFound,
				"Not Found",
				map[string]string{},
			),
			maxSize:      1024,
			expectError:  true,
			expectedData: "",
		},
		{
			name: "exceeds max size",
			response: createTestResponse(
				http.StatusOK,
				"This is too much data",
				map[string]string{},
			),
			maxSize:      10, // Less than the content length
			expectError:  true,
			expectedData: "",
		},
		{
			name:        "nil response",
			response:    nil,
			maxSize:     1024,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, state, err := ProcessRemoteFileResponse(tc.response, testURL, tc.maxSize)

			// Check error expectation
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// For successful cases, verify data
			if !tc.expectError {
				if string(data) != tc.expectedData {
					t.Errorf("Expected data %q, got %q", tc.expectedData, string(data))
				}
			}

			// Verify state
			// For URL and fields that we explicitly set
			if state.URL != testURL {
				t.Errorf("Expected URL %s, got %s", testURL, state.URL)
			}

			// Only check these fields if we're not testing a nil response and we have a successful response
			if tc.response != nil && tc.response.StatusCode == http.StatusOK && !tc.expectError {
				if state.ETag != "\"abc123\"" {
					t.Errorf("Expected ETag %s, got %s", "\"abc123\"", state.ETag)
				}

				if state.ContentType != "text/plain" {
					t.Errorf("Expected Content-Type %s, got %s", "text/plain", state.ContentType)
				}

				if tc.lastModified != "" && state.LastModified != tc.lastModified {
					t.Errorf("Expected LastModified %s, got %s", tc.lastModified, state.LastModified)
				}
			}

			// Check content length for non-nil responses with expected size
			if tc.response != nil && tc.response.ContentLength > 0 {
				if state.Size != tc.response.ContentLength {
					t.Errorf("Expected Size %d, got %d", tc.response.ContentLength, state.Size)
				}
			}
		})
	}
}

func TestProcessRemoteFileResponseStreaming(t *testing.T) {
	const testURL = "http://example.com/test.txt"
	const testContent = "This is a streaming test with multiple chunks of data."

	testCases := []struct {
		name         string
		response     *http.Response
		maxSize      int64
		chunkSize    int
		expectError  bool
		lastModified string // 各テストケースで個別に指定
	}{
		{
			name: "successful streaming",
			response: createTestResponse(
				http.StatusOK,
				testContent,
				map[string]string{
					"Content-Type":  "text/plain",
					"ETag":          "\"xyz789\"",
					"Last-Modified": time.Now().AddDate(0, 0, -1).Format(time.RFC1123),
				},
			),
			maxSize:      1024,
			chunkSize:    10, // Small chunks to test multiple reads
			expectError:  false,
			lastModified: time.Now().AddDate(0, 0, -1).Format(time.RFC1123),
		},
		{
			name: "error status code",
			response: createTestResponse(
				http.StatusNotFound,
				"Not Found",
				map[string]string{},
			),
			maxSize:     1024,
			chunkSize:   10,
			expectError: true,
		},
		{
			name: "exceeds max size",
			response: createTestResponse(
				http.StatusOK,
				testContent,
				map[string]string{},
			),
			maxSize:     10, // Less than the content length
			chunkSize:   5,
			expectError: true,
		},
		{
			name:        "nil response",
			response:    nil,
			maxSize:     1024,
			chunkSize:   10,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use a buffer to collect processed chunks
			var receivedData bytes.Buffer

			state, err := ProcessRemoteFileResponseStreaming(
				tc.response,
				testURL,
				tc.maxSize,
				tc.chunkSize,
				func(chunk []byte) error {
					// Simply collect chunks for validation
					receivedData.Write(chunk)
					return nil
				},
			)

			// Check error expectation
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// For successful cases, verify data
			if !tc.expectError && tc.response != nil {
				if receivedData.String() != testContent {
					t.Errorf("Expected data %q, got %q", testContent, receivedData.String())
				}
			}

			// Verify state
			if state.URL != testURL {
				t.Errorf("Expected URL %s, got %s", testURL, state.URL)
			}

			// Only check these fields if we're not testing a nil response and have a successful response
			if tc.response != nil && tc.response.StatusCode == http.StatusOK && !tc.expectError {
				if state.ETag != "\"xyz789\"" {
					t.Errorf("Expected ETag %s, got %s", "\"xyz789\"", state.ETag)
				}

				if state.ContentType != "text/plain" {
					t.Errorf("Expected Content-Type %s, got %s", "text/plain", state.ContentType)
				}

				if tc.lastModified != "" && state.LastModified != tc.lastModified {
					t.Errorf("Expected LastModified %s, got %s", tc.lastModified, state.LastModified)
				}
			}

			// Check content length for non-nil responses with expected size
			if tc.response != nil && tc.response.ContentLength > 0 {
				if state.Size != tc.response.ContentLength {
					t.Errorf("Expected Size %d, got %d", tc.response.ContentLength, state.Size)
				}
			}
		})
	}
}

func TestProcessRemoteFileResponseStreaming_ChunkProcessingError(t *testing.T) {
	const testURL = "http://example.com/test.txt"
	const errorMessage = "simulated processing error"

	// Create test response
	response := createTestResponse(
		http.StatusOK,
		"This should cause an error during processing",
		map[string]string{},
	)

	// Process with a function that returns an error
	_, err := ProcessRemoteFileResponseStreaming(
		response,
		testURL,
		1024,
		10,
		func(chunk []byte) error {
			return errors.New(errorMessage)
		},
	)

	// Verify error was propagated
	if err == nil {
		t.Error("Expected error but got nil")
	}

	if err != nil && !strings.Contains(err.Error(), errorMessage) {
		t.Errorf("Expected error containing %q, got %v", errorMessage, err)
	}
}
