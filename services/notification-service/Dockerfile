# ============================================================
# Pote Notification Service - Multi-stage Dockerfile
# ============================================================

FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY packages/logger/ ./packages/logger/
COPY packages/config/ ./packages/config/
COPY packages/auth-utils/ ./packages/auth-utils/
COPY services/notification-service/go.mod services/notification-service/go.sum* ./services/notification-service/
WORKDIR /app/services/notification-service
RUN go mod download
COPY services/notification-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server cmd/server/main.go

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /bin/server /bin/server
COPY services/notification-service/migrations/ /migrations/
EXPOSE 8087
ENTRYPOINT ["/bin/server"]
