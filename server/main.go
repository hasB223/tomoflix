// TomoFlix server entrypoint.
//
// For now this just wires up the local media streaming route. As we build
// more pieces (WebSocket sync hub, Google Drive proxy, SQLite storage)
// they'll get added here too.
package main

import (
	"log"
	"net/http"
	"os"

	"tomoflix/stream"
)

func main() {
	// Where our local media files (video or audio) live. In a real
	// deployment you'd load this from a config file or environment
	// variable - using an env var with a sensible default for now, so
	// it's easy to override without editing code.
	mediaDir := os.Getenv("TOMOFLIX_MEDIA_DIR")
	if mediaDir == "" {
		mediaDir = "./videos"
	}

	localMediaHandler := stream.NewLocalMediaHandler(mediaDir)

	// http.NewServeMux is Go's built-in router. It's basic (no wildcard
	// params like Express has), but it's enough for what we need right now.
	mux := http.NewServeMux()
	mux.Handle("/stream/local/", localMediaHandler)

	// Simple health check - useful later when this runs as a systemd
	// service or behind Cloudflare Tunnel, so we can confirm it's alive.
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	port := os.Getenv("TOMOFLIX_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("TomoFlix server starting on :%s (serving media from %s)", port, mediaDir)

	// http.ListenAndServe blocks here, handling requests until the process
	// is killed. log.Fatal will print the error and exit if the server
	// can't start (e.g. port already in use).
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
