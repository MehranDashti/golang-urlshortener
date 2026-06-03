# ── Stage 1: Build ────────────────────────────────────────────────
# Use full Go image to compile
FROM golang:1.22-alpine AS builder

# Install git — needed for some go modules
RUN apk add --no-cache git

WORKDIR /app

# Copy dependency files first — Docker layer caching
# If go.mod/go.sum don't change, this layer is cached
# and go mod download is skipped on next build
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build — CGO_ENABLED=0 = fully static binary (no C dependencies)
# -ldflags "-s -w" = strip debug info → smaller binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o bin/server \
    cmd/server/main.go

# ── Stage 2: Run ──────────────────────────────────────────────────
# Use minimal Alpine — not scratch, so we can run shell commands
# and have CA certificates for HTTPS outbound calls
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy only the binary and migrations from builder
COPY --from=builder /app/bin/server .

# Non-root user — security best practice
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080

CMD ["./server"]