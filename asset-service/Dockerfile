FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy proto module
COPY proto/ /proto/

# Copy asset-service module
COPY asset-service/ /app/

# Update go.mod to use local proto
RUN go mod edit -replace gym-membership/proto=/proto

RUN go mod tidy
RUN go build -o main cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 50051
CMD ["./main"]
