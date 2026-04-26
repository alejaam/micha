# Stage 1: Build React
FROM node:20-alpine AS frontend

WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go
FROM golang:1.24-alpine AS backend

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
# Copy the React build into the backend image as static assets.
COPY --from=frontend /app/frontend/dist ./static

RUN go build -o /app/server ./cmd/api

# Stage 3: Minimal runtime image
FROM alpine:latest

WORKDIR /app

COPY --from=backend /app/server ./server
COPY --from=backend /app/static ./static
COPY --from=backend /app/migrations ./migrations

EXPOSE 8080

CMD ["./server"]