# ============================================================
# Pote Chat Service - Multi-stage Dockerfile
# ============================================================

FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY packages/logger/ ./packages/logger/
COPY packages/config/ ./packages/config/
COPY packages/auth-utils/ ./packages/auth-utils/
COPY services/chat-service/go.mod services/chat-service/go.sum* ./services/chat-service/
WORKDIR /app/services/chat-service
RUN go mod download
COPY services/chat-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server cmd/server/main.go

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /bin/server /bin/server
COPY services/chat-service/migrations/ /migrations/
EXPOSE 8083
ENTRYPOINT ["/bin/server"]
