# TomoFlix Implementation Plan

A phase-by-phase roadmap. Check items off as they're completed, and note
which tool/session completed them so we can track continuity across
Claude (chat), Claude Code, and Codex sessions.

Keep phases small enough to finish in one sitting. Update this file
whenever a phase is completed or the plan changes.

---

## Phase 1 — Local video streaming ✅ DONE
- [x] Go module + `main.go` entrypoint
- [x] `stream/local.go` — range-request-aware local file serving
- [x] Dockerfile + docker-compose.yml for containerized deployment
- Completed: 2026-07-05 (Claude chat scaffolded core; Codex added
  Docker setup)
- Note (2026-07-06): `http.ServeFile` already serves audio files
  (.mp3/.flac/etc.) with range-request support with zero changes, so
  audio support doesn't require new Phase 1 work — just the renaming
  tracked in Phase 2 below.

## Phase 2 — WebSocket sync hub (NEXT)
- [ ] Rename `LocalVideoHandler` -> `LocalMediaHandler` and
      `stream/local.go` naming to be media-neutral (supports audio as a
      first-class type, not just video)
- [ ] Define sync message schema (`sync/messages.go`): play, pause, seek,
      heartbeat/currentTime report, drift-correct
- [ ] `sync/hub.go` — central broadcast hub (tracks connected clients per
      room, in-memory)
- [ ] `sync/client.go` — per-connection read/write goroutines
- [ ] Wire a `/ws` route into `main.go`
- [ ] Manual test: two browser tabs (or `wscat`) connect and see each
      other's messages broadcast

## Phase 3 — Google Drive proxy route
- [ ] Set up a Google Cloud service account with Drive read access
- [ ] `stream/drive_proxy.go` — fetches from Drive API, forwards Range
      headers, streams response through to client
- [ ] Route: `/stream/drive/:fileId`
- [ ] Manual test: play a Drive-hosted file in a `<video>` tag, confirm
      seeking works and Drive only sees one consumer (the server)

## Phase 4 — SQLite persistence
- [ ] Add SQLite driver dependency (`modernc.org/sqlite`)
- [ ] `db/schema.sql` — create tables per `context.md` schema (users,
      media_items, rooms, room_queue, watch_history) — `media_items`
      supports both video and audio via a `media_type` column plus
      nullable `artist`/`album` fields
- [ ] `db/db.go` — connection setup, migrations-on-startup
- [ ] Basic CRUD: create user, add media item (video or audio), create
      room
- [ ] Wire room/media data into the sync hub (rooms now have real IDs,
      not just in-memory-only state)

## Phase 5 — Frontend scaffold
- [ ] Vite + React + TypeScript project in `client/`
- [ ] `VideoPlayer.tsx` — `<video>` element wired to local/Drive stream
      URLs
- [ ] `AudioPlayer.tsx` — `<audio>` element for song/recording playback,
      reuses the same `useSyncSocket.ts` hook as VideoPlayer
- [ ] `useSyncSocket.ts` — WebSocket client hook mirroring `sync/messages.go`
      (media-type-agnostic; same hook drives both players)
- [ ] `RoomControls.tsx` — play/pause/seek UI
- [ ] Basic room join flow (even if just "everyone in one room" for v1)

## Phase 6 — Polish / real usage
- [ ] Deploy behind Cloudflare Tunnel + Cloudflare Access on the T460
      homelab (same pattern as Nextcloud)
- [ ] Watch history view
- [ ] Handle reconnect/drift edge cases (what happens if someone's
      connection drops mid-watch)

## Future / stretch (not scheduled)
- [ ] Avatar customization (Three.js / React Three Fiber)
- [ ] Simple friend-group gaming features
- [ ] HLS transcoding for adaptive bitrate (if range-serving proves
      insufficient)
- [ ] Valkey, if multi-instance or faster ephemeral state sharing is
      ever needed

---

## Notes for any tool picking this up
- Read `context.md` for *why* decisions were made, `ARCHITECTURE.md` for
  the current system shape, and this file for *what's next*.
- Keep code comprehensible to a junior/mid-level developer — see
  "Coding conventions" in `context.md`.
- Do not introduce a pirate/unlicensed streaming API as a video source —
  see `context.md` for the reasoning.
- Update this file's checkboxes and `ARCHITECTURE.md`'s component list
  when a phase is completed, so the docs don't drift from the code.
