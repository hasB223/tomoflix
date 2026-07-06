# TomoFlix — Project Context

## Purpose
A private, self-hosted "watch party" site for personal use with a few friends.

### Pain point being solved
Currently watching together via Discord screen share, but Discord heavily
compresses/re-encodes the video stream, resulting in low quality on the
viewer's end. Goal: watch the *same* video, in sync, at full quality, without
routing video bytes through Discord's call pipeline.

## Core concept
- Each viewer's browser plays video directly from our own server (or a proxy
  we control), so quality is limited only by our own hosting/bandwidth — not
  Discord's voice/video codec.
- A lightweight real-time control channel (WebSocket) keeps everyone's
  playback position (play/pause/seek/timestamp) in sync.
- One "host" (or the server itself) is authoritative for play state; clients
  periodically report their `currentTime`, and if a client drifts too far
  (e.g. >1–2s) the server nudges them back in sync.

## Video sources (planned)
1. **Local .mp4 files**
   - Serve via HTTP range requests (native `<video>` tag support for seeking).
   - Options: simple Node/Express static server with range support, or
     transcode to HLS (ffmpeg -> .m3u8 + .ts segments) for adaptive bitrate /
     better handling of uneven friend connections. Range-serving is likely
     sufficient for a small friend group; HLS is a stretch goal.
2. **Google Drive files**
   - Feasible via Drive API `files.get?alt=media` (supports Range headers).
   - Caveats: Drive isn't a CDN — repeated/concurrent access can trigger
     throttling or "too many viewers" errors, especially on shared files.
   - Chosen approach: treat Drive as **origin storage**, not direct client
     source. Our server proxies the Drive file (single consumer from
     Google's perspective) and re-serves it to all clients with our own
     range-request handling. Optionally cache/transcode once to local
     disk/HLS for repeat watches or reliability.
3. **External streaming APIs (e.g. of the kind used by pirate streaming
   sites)**
   - Technically discussed as a possible source type, but Claude will not
     help source, discover, or integrate infringing/pirate content APIs —
     this is a hard no regardless of "personal use" framing, since it
     constitutes distributing copyrighted content without a license.
   - User has stated they're capable of researching this on their own later
     if desired. Anything built collaboratively in this project should
     assume **legally-sourced** video (files you own/ripped from owned
     media, Drive-hosted personal files, or official streaming service
     integrations where ToS-compliant, e.g. official watch-party features or
     browser-tab sharing).

## Architecture (planned)
```
tomoflix/
├── context.md          <- this file
├── README.md
├── docker-compose.yml  <- builds/runs the server container, mounts video dir
├── server/             <- sync websocket server + streaming/proxy routes
│   ├── Dockerfile       <- multi-stage build -> distroless runtime image
│   ├── (sync server: play/pause/seek broadcast, drift correction)
│   └── (streaming routes: /stream/local/:file, /stream/drive/:fileId)
└── client/             <- frontend
    └── (HTML5 <video> or video.js, wired to WebSocket sync client)
```

## Stack decision (2026-07-05)
- **Backend: Go** — chosen for the learning experience and because it's a
  strong natural fit for this workload: `http.ServeContent` gives range-
  request serving almost for free, goroutines/channels fit the broadcast/
  sync-hub model well, and single static binary deploys are convenient for a
  home machine or small VPS.
- **Frontend: React + TypeScript + Vite.** Chosen over plain HTML/JS because
  of a stated stretch goal: possibly expanding TomoFlix into user avatar
  customization (Three.js) and maybe a lightweight custom-friend-gaming site
  down the line. React + **React Three Fiber** (R3F) is the standard,
  well-supported way to combine Three.js with a component-based UI
  (avatar customization panels, lobby/room UI, etc.), so this keeps that
  future feature additive rather than a rewrite. The watch-party MVP itself
  is just plain React components; R3F only gets added when the
  avatar/gaming feature is actually built.

### Planned backend layout (Go)
```
server/
├── go.mod
├── main.go              # entrypoint, wires up HTTP + WS servers
├── Dockerfile           # multi-stage build: golang:1.22-alpine -> distroless
├── .dockerignore
├── stream/
│   ├── local.go          # http.ServeContent range serving for local mp4
│   └── drive_proxy.go    # proxies Google Drive API file -> client
├── sync/
│   ├── hub.go             # broadcast hub: connected clients, room state
│   ├── client.go          # per-connection goroutine (read/write pumps)
│   └── messages.go        # message types: play/pause/seek/heartbeat/drift
└── config/
    └── config.go          # ports, Drive service account path, etc.
```

### Planned frontend layout (React/TS/Vite)
```
client/
├── package.json / tsconfig.json / vite.config.ts
├── src/
│   ├── main.tsx
│   ├── App.tsx
│   ├── components/
│   │   ├── VideoPlayer.tsx     # <video> element + sync client hookup
│   │   └── RoomControls.tsx    # play/pause/seek UI
│   ├── sync/
│   │   └── useSyncSocket.ts    # WebSocket client hook (mirrors server messages.go)
│   └── (future) avatar/        # Three.js / R3F avatar customization, later
```

## Persistence decision (2026-07-05)
- **Primary store: SQLite.** Fits the scale (a handful of friends, low write
  volume), zero ops overhead, single file (`tomoflix.db`) pairs naturally
  with the single-Go-binary deploy goal. Use a pure-Go driver
  (`modernc.org/sqlite`) to avoid cgo cross-compile friction, or
  `mattn/go-sqlite3` if cgo isn't a problem in the deploy environment.
  Holds: users, rooms (metadata, not live state), videos (local path or
  Drive file ID + title), watch history, and later avatar/customization
  configs.
- **Live/ephemeral room state (current playback position, connected
  clients): kept in-memory** in Go structs behind `sync/hub.go`, not
  persisted to SQLite. Doesn't need to survive a server restart in any
  meaningful way for this use case.
- **Redis/Valkey: not adopted yet.** Considered for cross-instance pub/sub
  or fast ephemeral session storage, but unnecessary at current scale
  (single server instance, a few concurrent users). If ever needed, decided
  to use **Valkey** over Redis: BSD-3 licensed (vs Redis's
  RSALv2/SSPL/AGPL), wire/protocol-compatible drop-in replacement (same
  clients, no code changes), and it's the default package on Debian 12
  (matches the homelab OS already in use). Revisit only if TomoFlix grows
  into multiple server instances or the gaming-site ambition demands faster
  ephemeral state sharing.
- **Migration path if it grows**: `database/sql` usage kept idiomatic (or
  via `sqlc`) so swapping SQLite -> Postgres later is mostly a driver/DSN
  change, not a rewrite, if the gaming-site ambition ever demands real
  concurrent-writer scale.

### Planned initial schema (SQLite)
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    display_name TEXT,
    avatar_config TEXT,        -- JSON blob, for future Three.js avatar customization
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Renamed from "videos" to "media_items" (2026-07-06) to support audio
-- (downloaded songs/recordings) as a first-class type alongside video,
-- not just an afterthought bolted onto a video-shaped table.
CREATE TABLE media_items (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    media_type TEXT NOT NULL CHECK (media_type IN ('video', 'audio')),
    source_type TEXT NOT NULL CHECK (source_type IN ('local', 'drive')),
    source_ref TEXT NOT NULL, -- local file path OR google drive file ID
    duration_seconds INTEGER,
    -- Audio-specific fields, nullable since they don't apply to video.
    -- Kept as plain columns rather than a JSON blob since these are the
    -- only two known type-specific fields right now; revisit if more
    -- type-specific metadata shows up later.
    artist TEXT,
    album TEXT,
    added_by INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE rooms (
    id INTEGER PRIMARY KEY,
    name TEXT,
    current_media_id INTEGER REFERENCES media_items(id),
    host_user_id INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    -- live playback position/connected clients intentionally NOT here;
    -- lives in in-memory hub state, not the DB
);

-- Optional lightweight queue, for "listen to an album together" or
-- "watch a playlist" use cases. Not required for v1 (single item per
-- room is fine to start), but the shape is here so it's not a surprise
-- schema change later.
CREATE TABLE room_queue (
    id INTEGER PRIMARY KEY,
    room_id INTEGER REFERENCES rooms(id),
    media_id INTEGER REFERENCES media_items(id),
    position INTEGER NOT NULL, -- order within the queue
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE watch_history (
    id INTEGER PRIMARY KEY,
    room_id INTEGER REFERENCES rooms(id),
    media_id INTEGER REFERENCES media_items(id),
    watched_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Coding conventions
- Code should stay comprehensible to a junior/mid-level developer, even where
  Go idioms (goroutines, channels) might tempt cleverness. Favor explicit,
  readable control flow over dense/idiomatic-but-opaque patterns. Comment
  the *why* especially around concurrency (goroutines, channels, mutexes),
  since that's the least familiar part of Go coming from JS/TS.

## Deployment (2026-07-05)
- **Server is containerized.** `server/Dockerfile` is a multi-stage build:
  `golang:1.22-alpine` compiles a static binary (`CGO_ENABLED=0`), then it's
  copied into a minimal `gcr.io/distroless/static-debian12` runtime image
  (no shell, no package manager, non-root by default) — small attack
  surface and small image size, in keeping with the single-binary deploy
  goal.
- **`docker-compose.yml`** at the repo root builds the `server/` image,
  publishes port 8080, and bind-mounts a host `./videos` directory into
  `/app/videos` (matching `TOMOFLIX_VIDEO_DIR`'s default). Run with
  `docker compose up --build`.
- Frontend (`client/`) has no Docker setup yet since it's still just a
  placeholder — will get its own build stage (or an nginx-serves-static-
  build stage) once the React app exists.

## Key decisions log
- **2026-07-05**: Project named "TomoFlix" (友 = "friend" in Japanese +
  "-flix"). Private/personal use only; no plans to make it a public-facing
  product, so no trademark concerns in practice.
- **2026-07-05**: Sync approach = WebSocket control channel + direct video
  serving, instead of Discord screen share.
- **2026-07-05**: Local files served via HTTP range requests initially;
  HLS transcoding via ffmpeg considered as a later enhancement.
- **2026-07-05**: Google Drive sources will be proxied through our own
  server rather than linked directly to clients, to avoid Drive's per-file
  viewer throttling and to keep control over range-request handling.
- **2026-07-05**: Explicitly decided NOT to integrate pirate/unlicensed
  streaming APIs as part of this collaborative build.
- **2026-07-06**: Decided to support audio (downloaded songs/recordings)
  as a first-class media type alongside video. Backend streaming layer
  (`http.ServeFile`-based range serving) already works for any file type
  without changes. Renamed the planned `videos` table to `media_items`
  with a `media_type` column, plus nullable `artist`/`album` columns for
  audio metadata. Added an optional `room_queue` table for playlist/
  album-style multi-item rooms. Sync hub (WebSocket play/pause/seek
  logic) is unaffected either way — it only deals in abstract playback
  control messages, not video vs audio specifically. Code naming
  (`LocalVideoHandler`, `stream/local.go`) will be renamed to
  media-neutral terms (`LocalMediaHandler`) during Phase 2/3 work, before
  Phase 4 (SQLite) locks in the schema.
- **2026-07-05**: Server containerized via a multi-stage Dockerfile
  (distroless runtime image) plus a root-level `docker-compose.yml`, ahead
  of building further backend features, so the deploy story is settled
  early.

## Open questions / next steps
- [x] Build `/stream/local/:file` route with Range header support.
- [x] Containerize the server (Dockerfile + docker-compose.yml).
- [ ] Decide: plain WebSocket (ws) vs Socket.IO for the sync server.
- [ ] Decide: simple range-serving vs ffmpeg HLS transcoding for local files.
- [ ] Build `/stream/drive/:fileId` proxy route using a Drive service account.
- [ ] Build minimal frontend: video element + sync client + simple room/host
      model (how many concurrent "rooms" do we need? Probably just 1 for a
      small friend group).
- [ ] Decide hosting: run on a home machine (Tailscale/ngrok?) vs small VPS
      (containerization done — this is now just about *where* the
      container runs).
