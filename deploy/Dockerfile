FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

# Fix: Use the correct path to the backend directory
WORKDIR /app/backend

RUN go mod tidy
RUN go build -o /app/micha-api ./cmd/api/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/micha-api .

CMD ["./micha-api"]
