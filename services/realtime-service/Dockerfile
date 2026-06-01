# ============================================================
# Pote Realtime Service - Multi-stage Dockerfile
# ============================================================

FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY packages/logger/ ./packages/logger/
COPY packages/config/ ./packages/config/
COPY packages/auth-utils/ ./packages/auth-utils/
COPY services/realtime-service/go.mod services/realtime-service/go.sum* ./services/realtime-service/
WORKDIR /app/services/realtime-service
RUN go mod download
COPY services/realtime-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server cmd/server/main.go

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /bin/server /bin/server
EXPOSE 8085
ENTRYPOINT ["/bin/server"]
