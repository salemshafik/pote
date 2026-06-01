# ============================================================
# Pote Audit Service - Multi-stage Dockerfile
# ============================================================

FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY packages/logger/ ./packages/logger/
COPY packages/config/ ./packages/config/
COPY packages/auth-utils/ ./packages/auth-utils/
COPY services/audit-service/go.mod services/audit-service/go.sum* ./services/audit-service/
WORKDIR /app/services/audit-service
RUN go mod download
COPY services/audit-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server cmd/server/main.go

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /bin/server /bin/server
COPY services/audit-service/migrations/ /migrations/
EXPOSE 8089
ENTRYPOINT ["/bin/server"]
