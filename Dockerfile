FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY backend/ ./

WORKDIR /app/cmd/api

RUN go mod tidy
RUN go build -o /tmp/micha-api .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /tmp/micha-api ./micha-api

CMD ["./micha-api"]
