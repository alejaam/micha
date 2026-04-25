FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy backend source
COPY backend/ ./backend/

WORKDIR /app/backend

# Fetch dependencies and build
RUN go mod tidy
RUN go build -o /app/micha-api ./cmd/api/main.go

FROM alpine:latest

WORKDIR /app

# Install certificates for secure connections
RUN apk add --no-cache ca-certificates

# Copy binary and migrations from builder
COPY --from=builder /app/backend/migrations ./migrations
COPY --from=builder /app/micha-api .

# Default environment variables
ENV MIGRATIONS_DIR=/app/migrations
ENV PORT=8080

EXPOSE 8080

CMD ["./micha-api"]
