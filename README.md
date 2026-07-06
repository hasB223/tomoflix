# TomoFlix

A private, self-hosted watch-party site for me and a few friends.

Solves the problem of watching together via Discord screen share, where
Discord's video pipeline heavily compresses the stream. TomoFlix serves video
directly (local files or a proxied Google Drive source) and keeps everyone in
sync via a WebSocket control channel, so quality isn't bottlenecked by a
call/chat app.

See `context.md` for full project notes, architecture, and decision log.

## Running it

The server is containerized. From the repo root:

```sh
docker compose up --build
```

This builds `server/` (Go, multi-stage Dockerfile) and serves it on
`http://localhost:8080`. Drop media files (video or audio — `.mp4`, `.mp3`,
etc.) into a `./videos` directory at the repo root — it's bind-mounted into
the container and served at `/stream/local/<filename>`. Check
`http://localhost:8080/health` to confirm it's up.

There's no frontend yet (`client/` is still a placeholder), so there's
nothing to watch through a browser UI just yet.

## Status
Backend scaffolding + Docker deploy in place (local file streaming route,
containerized server). Sync server, Drive proxy, persistence, and the
frontend are still to come — see `context.md` "Open questions / next steps".
