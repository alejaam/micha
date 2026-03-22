FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/cmd/api

RUN go mod tidy
RUN go build -o micha-api .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/cmd/api/micha-api .

CMD ["./micha-api"]
