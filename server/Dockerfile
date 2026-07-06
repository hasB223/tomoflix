# Build stage: compiles a static binary. Kept separate from the runtime
# stage so the final image doesn't carry the Go toolchain around.
FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 gives a statically-linked binary that runs on the
# minimal base image below (no libc dependency).
RUN CGO_ENABLED=0 GOOS=linux go build -o /tomoflix-server .

# Runtime stage: just the binary on a minimal, non-root base.
FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /tomoflix-server ./tomoflix-server

# Matches the TOMOFLIX_MEDIA_DIR default in main.go - media files are
# expected to be mounted here as a volume.
VOLUME ["/app/videos"]

EXPOSE 8080

ENTRYPOINT ["./tomoflix-server"]
