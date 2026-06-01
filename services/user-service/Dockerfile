# ============================================================
# Pote User Service - Multi-stage Dockerfile
# ============================================================

FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY packages/logger/ ./packages/logger/
COPY packages/config/ ./packages/config/
COPY packages/auth-utils/ ./packages/auth-utils/
COPY services/user-service/go.mod services/user-service/go.sum* ./services/user-service/
WORKDIR /app/services/user-service
RUN go mod download
COPY services/user-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server cmd/server/main.go

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /bin/server /bin/server
COPY services/user-service/migrations/ /migrations/
EXPOSE 8082
ENTRYPOINT ["/bin/server"]
