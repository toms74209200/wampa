// Package watcher provides file monitoring capabilities.
package watcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// RemoteFileState represents metadata about a remote file
type RemoteFileState struct {
	URL          string
	LastModified string
	ETag         string
	ContentType  string
	Size         int64
}

// CreateRemoteFileRequest creates an HTTP request for a remote file.
// This is a pure function with no side effects.
// ctx: context for cancellation
// url: the URL of the remote file
// headers: optional HTTP headers (e.g., User-Agent)
func CreateRemoteFileRequest(ctx context.Context, url string, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request for %s: %w", url, err)
	}

	// Add custom headers if provided
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

// ProcessRemoteFileResponse processes an HTTP response to extract file content and metadata.
// This is a pure function with no side effects.
// resp: HTTP response to process
// url: original request URL (included in metadata)
// maxSize: maximum allowed file size in bytes
// Note: Caller is responsible for closing resp.Body
func ProcessRemoteFileResponse(resp *http.Response, url string, maxSize int64) ([]byte, RemoteFileState, error) {
	if resp == nil {
		return nil, RemoteFileState{URL: url}, errors.New("nil HTTP response")
	}

	// Initialize state with available metadata
	state := RemoteFileState{
		URL:          url,
		LastModified: resp.Header.Get("Last-Modified"),
		ETag:         resp.Header.Get("ETag"),
		ContentType:  resp.Header.Get("Content-Type"),
		Size:         resp.ContentLength,
	}

	// Check for non-success status code
	if resp.StatusCode != http.StatusOK {
		return nil, state, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url)
	}

	// Check content size against maximum allowed
	if resp.ContentLength > maxSize {
		return nil, state, fmt.Errorf("file size %d exceeds maximum allowed size %d", resp.ContentLength, maxSize)
	}

	// Read response body with size limit
	limitReader := io.LimitReader(resp.Body, maxSize)
	content, err := io.ReadAll(limitReader)
	if err != nil {
		return nil, state, fmt.Errorf("reading response body from %s: %w", url, err)
	}

	return content, state, nil
}

// ProcessRemoteFileResponseStreaming processes an HTTP response as a stream, calling the provided
// function for each chunk of data read.
// This is a pure function with no side effects.
// resp: HTTP response to process
// url: original request URL (included in metadata)
// maxSize: maximum allowed file size in bytes
// chunkSize: size of chunks to process at once
// processChunk: function to process each chunk of data
// Note: Caller is responsible for closing resp.Body
func ProcessRemoteFileResponseStreaming(
	resp *http.Response,
	url string,
	maxSize int64,
	chunkSize int,
	processChunk func([]byte) error,
) (RemoteFileState, error) {
	if resp == nil {
		return RemoteFileState{URL: url}, errors.New("nil HTTP response")
	}

	// Initialize state with available metadata
	state := RemoteFileState{
		URL:          url,
		LastModified: resp.Header.Get("Last-Modified"),
		ETag:         resp.Header.Get("ETag"),
		ContentType:  resp.Header.Get("Content-Type"),
		Size:         resp.ContentLength,
	}

	// Check for non-success status code
	if resp.StatusCode != http.StatusOK {
		return state, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url)
	}

	// Check content size against maximum allowed
	if resp.ContentLength > maxSize {
		return state, fmt.Errorf("file size %d exceeds maximum allowed size %d", resp.ContentLength, maxSize)
	}

	// Read and process response in chunks
	buffer := make([]byte, chunkSize)
	var totalBytes int64

	for {
		n, err := resp.Body.Read(buffer)

		// Process any data that was read
		if n > 0 {
			// Check total size against maximum
			totalBytes += int64(n)
			if totalBytes > maxSize {
				return state, fmt.Errorf("file size exceeds maximum allowed size %d", maxSize)
			}

			// Process this chunk
			if err := processChunk(buffer[:n]); err != nil {
				return state, fmt.Errorf("processing chunk: %w", err)
			}
		}

		// Handle end of file or other errors
		if err == io.EOF {
			break
		}

		if err != nil {
			return state, fmt.Errorf("reading response body from %s: %w", url, err)
		}
	}

	return state, nil
}
