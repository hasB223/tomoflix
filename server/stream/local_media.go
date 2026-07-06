// Package stream contains HTTP handlers that serve media content to
// clients. This file handles video and audio files stored locally on disk.
package stream

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// LocalMediaHandler serves media files (video or audio) from a single
// directory on disk, using Go's standard library support for HTTP "Range"
// requests.
//
// Why range requests matter: when a browser's <video> or <audio> element
// wants to seek to the middle of a file, or just wants to start buffering,
// it sends a request like "Range: bytes=1000000-2000000" instead of asking
// for the whole file. Go's http.ServeContent understands these headers
// and automatically returns just the requested byte range (HTTP 206
// Partial Content) - we don't have to parse Range headers ourselves. This
// works identically for video and audio files, so one handler covers both.
type LocalMediaHandler struct {
	// MediaDir is the folder on disk where our local media files live.
	// Example: "/home/has/tomoflix-media"
	MediaDir string
}

// NewLocalMediaHandler builds a handler that serves files out of mediaDir.
func NewLocalMediaHandler(mediaDir string) *LocalMediaHandler {
	return &LocalMediaHandler{MediaDir: mediaDir}
}

// ServeHTTP makes LocalMediaHandler satisfy the http.Handler interface,
// so we can mount it directly on a route like:
//
//	mux.Handle("/stream/local/", localMediaHandler)
//
// Expected URL shape: /stream/local/somefile.mp4 (or .mp3, .flac, etc.)
func (h *LocalMediaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Pull just the filename portion off the URL path.
	// e.g. "/stream/local/somefile.mp4" -> "somefile.mp4"
	requestedFile := strings.TrimPrefix(r.URL.Path, "/stream/local/")

	if requestedFile == "" {
		http.Error(w, "no file specified", http.StatusBadRequest)
		return
	}

	// SECURITY: filepath.Base strips out any directory components like
	// "../../etc/passwd", so a malicious request can't escape MediaDir
	// and read arbitrary files off the server's disk. Always do this
	// when building a file path from user/request input.
	safeFileName := filepath.Base(requestedFile)
	fullPath := filepath.Join(h.MediaDir, safeFileName)

	log.Printf("serving local media: %s", fullPath)

	// http.ServeFile handles range requests, sets the correct
	// Content-Type based on the file extension, and returns 404 if the
	// file doesn't exist - all standard library, no extra packages needed.
	http.ServeFile(w, r, fullPath)
}
