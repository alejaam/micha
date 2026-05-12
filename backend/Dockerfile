# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Instalamos certificados para llamadas HTTPS si fueran necesarias
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Aprovechamos el cache de Docker para las dependencias
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build estático para Alpine
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/micha-api ./cmd/api/main.go

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Copiamos solo lo necesario desde el builder
COPY --from=builder /app/micha-api .
COPY --from=builder /app/migrations ./migrations

# Exponemos el puerto estándar de la API
EXPOSE 8080

# Ejecutamos el binario directamente (sin go run para producción)
CMD ["./micha-api"]
