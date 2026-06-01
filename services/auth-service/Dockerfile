# ============================================================
# Pote Auth Service - Multi-stage Dockerfile
# ============================================================

# ---------- Stage 1: Build ----------
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy shared packages first (for better layer caching)
COPY packages/logger/ ./packages/logger/
COPY packages/config/ ./packages/config/
COPY packages/auth-utils/ ./packages/auth-utils/

# Copy service module files
COPY services/auth-service/go.mod services/auth-service/go.sum* ./services/auth-service/

WORKDIR /app/services/auth-service
RUN go mod download

# Copy service source code
COPY services/auth-service/ ./

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server cmd/server/main.go

# ---------- Stage 2: Runtime ----------
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /bin/server /bin/server
COPY services/auth-service/migrations/ /migrations/

EXPOSE 8081

ENTRYPOINT ["/bin/server"]
