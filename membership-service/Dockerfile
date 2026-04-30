FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy the proto workspace
COPY proto/ proto/
COPY membership-service/ membership-service/

WORKDIR /app/membership-service
RUN go mod download
RUN go build -o main ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/membership-service/main .
EXPOSE 50052
CMD ["./main"]
