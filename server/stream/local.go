// Package stream contains HTTP handlers that serve video content to
// clients. This file handles videos stored as local files on disk.
package stream

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// LocalVideoHandler serves .mp4 files from a single directory on disk,
// using Go's standard library support for HTTP "Range" requests.
//
// Why range requests matter: when a browser's <video> element wants to
// seek to the middle of a video, or just wants to start buffering, it
// sends a request like "Range: bytes=1000000-2000000" instead of asking
// for the whole file. Go's http.ServeContent understands these headers
// and automatically returns just the requested byte range (HTTP 206
// Partial Content) - we don't have to parse Range headers ourselves.
type LocalVideoHandler struct {
	// VideoDir is the folder on disk where our .mp4 files live.
	// Example: "/home/has/tomoflix-videos"
	VideoDir string
}

// NewLocalVideoHandler builds a handler that serves files out of videoDir.
func NewLocalVideoHandler(videoDir string) *LocalVideoHandler {
	return &LocalVideoHandler{VideoDir: videoDir}
}

// ServeHTTP makes LocalVideoHandler satisfy the http.Handler interface,
// so we can mount it directly on a route like:
//
//	mux.Handle("/stream/local/", localVideoHandler)
//
// Expected URL shape: /stream/local/somefile.mp4
func (h *LocalVideoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Pull just the filename portion off the URL path.
	// e.g. "/stream/local/somefile.mp4" -> "somefile.mp4"
	requestedFile := strings.TrimPrefix(r.URL.Path, "/stream/local/")

	if requestedFile == "" {
		http.Error(w, "no file specified", http.StatusBadRequest)
		return
	}

	// SECURITY: filepath.Base strips out any directory components like
	// "../../etc/passwd", so a malicious request can't escape VideoDir
	// and read arbitrary files off the server's disk. Always do this
	// when building a file path from user/request input.
	safeFileName := filepath.Base(requestedFile)
	fullPath := filepath.Join(h.VideoDir, safeFileName)

	log.Printf("serving local video: %s", fullPath)

	// http.ServeFile handles range requests, sets the correct
	// Content-Type based on the file extension, and returns 404 if the
	// file doesn't exist - all standard library, no extra packages needed.
	http.ServeFile(w, r, fullPath)
}
