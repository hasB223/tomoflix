# TomoFlix Architecture

This is a snapshot of the system's actual shape. Unlike `context.md` (which
is an append-only decision log), this file gets **rewritten** as the system
evolves, so it should always reflect current reality, not history.

## High-level diagram

```
┌─────────────┐        WebSocket (sync)        ┌─────────────────┐
│  Browser A  │◄──────────────────────────────►│                  │
│ (React+Vite)│                                 │   Go Server      │
└─────────────┘        HTTP (video bytes)       │                  │
       ▲        ◄──────────────────────────────►│  ┌────────────┐  │
       │                                         │  │ sync.Hub   │  │
┌─────────────┐        WebSocket (sync)          │  │(in-memory) │  │
│  Browser B  │◄──────────────────────────────►  │  └────────────┘  │
│ (React+Vite)│                                  │  ┌────────────┐  │
└─────────────┘        HTTP (video bytes)        │  │stream.Local│  │
       ▲        ◄──────────────────────────────  │  │stream.Drive│──┼──► Google Drive API
       │                                         │  └────────────┘  │
       └── (more friends' browsers) ──►          │  ┌────────────┐  │
                                                  │  │  SQLite DB │  │
                                                  │  └────────────┘  │
                                                  └─────────────────┘
```

## Components

### Go server (`server/`)
Single binary, single process (for now). Runs two kinds of traffic on the
same HTTP server:
- **Plain HTTP routes** — serve video bytes (range-request aware) and
  eventually REST-ish endpoints for room/video management.
- **WebSocket connections** — one persistent connection per connected
  browser, used only for small sync control messages (play/pause/seek/
  drift-correction), never for video data itself.

**Current packages:**
- `stream/` — HTTP handlers that serve media bytes (video AND audio;
  `http.ServeFile` handles range requests for any file type, so no
  separate code path is needed for audio).
  - `local_media.go` — `LocalMediaHandler` serves local files with
    range-request support (implemented ✅; media-neutral naming as of
    2026-07-06 — works identically for video `.mp4` and audio
    `.mp3`/`.flac`/etc., same handler/code path for both)
  - `drive_proxy.go` — will proxy Google Drive files (video or audio),
    single Drive consumer regardless of client count (not yet built)
- `sync/` — will hold the WebSocket broadcast hub (not yet built)
- `config/` — will hold environment/config loading (not yet built;
  currently just inline env var reads in `main.go`)
- `db/` — will hold SQLite setup + queries (not yet built)

### Frontend (`client/`)
React + TypeScript + Vite. Not yet scaffolded (empty except `.gitkeep`).
Planned shape is documented in `context.md` under "Planned frontend
layout."

### Persistence
- **SQLite** (not yet wired up) — durable data: users, media_items
  (video AND audio, distinguished by a `media_type` column), rooms
  (metadata only), an optional room_queue for playlist/album-style
  multi-item rooms, and watch history. Schema drafted in `context.md`.
- **In-memory Go structs** — live/ephemeral room state (current playback
  position, connected clients). Lives inside the (not yet built) `sync.Hub`.

### Deployment
- `server/Dockerfile` + root `docker-compose.yml` — containerizes the Go
  server, mounts a host `./videos` directory into the container, exposes
  port 8080. Matches the homelab's existing Docker + Cloudflare Tunnel
  pattern (same approach used for Nextcloud) — this container can be
  tunneled to `hasb.dev` the same way.
- Frontend deployment not yet decided (likely: build static assets, serve
  either from the Go binary itself or a separate static host/route).

## Data flow: "watch/listen together" (target behavior, not all built yet)

The flow below is written in terms of video, but applies identically to
audio — the sync hub only ever deals with abstract playback control
messages, never with the media type itself.

1. A friend opens the TomoFlix site, connects, and a WebSocket connection
   is established to the Go server.
2. The server's `sync.Hub` tracks this connection alongside others in the
   same "room."
3. When the host presses play/pause/seeks, their browser sends a small
   JSON message over the WebSocket (e.g. `{"type":"seek","time":125.4}`).
4. `sync.Hub` broadcasts that message to every other connection in the
   room.
5. Each browser's `<video>` element (playing from `/stream/local/...` or
   `/stream/drive/...`, fetched directly from the Go server or proxied
   Drive) applies the change locally — pause, seek, etc.
6. Periodically, clients report their own `currentTime` back; if a client
   has drifted more than ~1-2 seconds from the host, the server sends a
   correction.

## What's NOT built yet (see `PLAN.md` for sequencing)
- WebSocket sync hub
- Google Drive proxy route
- SQLite persistence layer
- Frontend (any of it)
- Auth/access control (currently: anyone who can reach the server can
  watch — fine for a small trusted friend group behind Cloudflare Access,
  worth revisiting if that changes)
