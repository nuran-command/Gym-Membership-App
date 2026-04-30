FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy the proto workspace (we assume context is the root)
COPY proto/ proto/
COPY asset-service/ asset-service/

WORKDIR /app/asset-service
RUN go mod download
RUN go build -o main ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/asset-service/main .
EXPOSE 50051
CMD ["./main"]
