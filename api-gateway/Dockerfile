FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy proto module
COPY proto /proto

# Copy gateway module
COPY api-gateway /app/api-gateway
WORKDIR /app/api-gateway

# Add replacement for proto
RUN go mod edit -replace gym-membership/proto=/proto
RUN go mod download
RUN go build -o gateway main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api-gateway/gateway .
EXPOSE 8080
CMD ["./gateway"]
