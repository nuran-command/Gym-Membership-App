FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy the proto workspace
COPY proto/ proto/
COPY telemetry-service/ telemetry-service/

WORKDIR /app/telemetry-service
RUN go mod download
RUN go build -o main ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/telemetry-service/main .
EXPOSE 50053
CMD ["./main"]
