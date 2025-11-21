# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o captval ./main.go

# Runtime stage
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/captval .
ENTRYPOINT ["/app/captval"]
