FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy proto module
COPY proto/ /proto/

# Copy membership-service module
COPY membership-service/ /app/

# Update go.mod to use local proto relative to the container build path
RUN go mod edit -replace github.com/ilnur/gym-membership-app/proto=/proto

RUN go mod tidy
RUN go build -o main cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 50052
CMD ["./main"]
