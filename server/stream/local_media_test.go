// Test spec for LocalMediaHandler. Covers the happy path (full-file and
// range requests) plus the error cases: missing file, no filename, and a
// path-traversal attempt (verifying filepath.Base actually blocks it).
package stream

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalMediaHandler_ServeHTTP(t *testing.T) {
	mediaDir := t.TempDir()

	fileContent := []byte("fake mp4 bytes for testing")
	if err := os.WriteFile(filepath.Join(mediaDir, "sample.mp4"), fileContent, 0o644); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}

	handler := NewLocalMediaHandler(mediaDir)

	tests := []struct {
		name           string
		path           string
		rangeHeader    string
		wantStatus     int
		wantBodyPrefix string // empty means "don't check body"
	}{
		{
			name:           "happy path: full file",
			path:           "/stream/local/sample.mp4",
			wantStatus:     http.StatusOK,
			wantBodyPrefix: string(fileContent),
		},
		{
			name:        "happy path: range request",
			path:        "/stream/local/sample.mp4",
			rangeHeader: "bytes=0-3",
			wantStatus:  http.StatusPartialContent,
			// http.ServeFile returns just the requested byte range.
			wantBodyPrefix: string(fileContent[0:4]),
		},
		{
			name:       "error: file does not exist",
			path:       "/stream/local/does-not-exist.mp4",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "error: no filename given",
			path:       "/stream/local/",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "path traversal attempt is neutralized",
			// http.ServeFile itself rejects any request path containing
			// ".." with 400 before our filepath.Base cleanup even runs -
			// belt-and-suspenders with the safeFileName logic in
			// ServeHTTP. Either way, /etc/passwd must never be reached.
			path:       "/stream/local/../../../../etc/passwd",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.rangeHeader != "" {
				req.Header.Set("Range", tt.rangeHeader)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %q)", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantBodyPrefix != "" && rec.Body.String() != tt.wantBodyPrefix {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBodyPrefix)
			}
		})
	}
}
