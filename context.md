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
├── server/             <- sync websocket server + streaming/proxy routes
│   ├── (sync server: play/pause/seek broadcast, drift correction)
│   └── (streaming routes: /stream/local/:file, /stream/drive/:fileId)
└── client/             <- frontend
    └── (HTML5 <video> or video.js, wired to WebSocket sync client)
```

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

## Open questions / next steps
- [ ] Decide: plain WebSocket (ws) vs Socket.IO for the sync server.
- [ ] Decide: simple range-serving vs ffmpeg HLS transcoding for local files.
- [ ] Build `/stream/local/:file` route with Range header support.
- [ ] Build `/stream/drive/:fileId` proxy route using a Drive service account.
- [ ] Build minimal frontend: video element + sync client + simple room/host
      model (how many concurrent "rooms" do we need? Probably just 1 for a
      small friend group).
- [ ] Decide hosting: run on a home machine (Tailscale/ngrok?) vs small VPS.
