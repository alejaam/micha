FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy all files from the context
COPY . .

# Robustness check: If go.mod is inside a 'backend' directory (root context),
# move its contents to /app to standardize the build environment.
# If go.mod is already at the root (backend context), do nothing.
RUN if [ -d "backend" ] && [ -f "backend/go.mod" ]; then \
      echo "Detected root context, moving backend files to /app"; \
      mv backend/* . && mv backend/.[!.]* . 2>/dev/null || true && rm -rf backend; \
    fi

# Verify we have a go.mod file
RUN if [ ! -f go.mod ]; then \
      echo "Error: go.mod not found at /app. Context might be incorrect."; \
      ls -la; \
      exit 1; \
    fi

RUN go mod tidy
RUN go build -o /app/micha-api ./cmd/api/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/micha-api .

# Ensure the binary is executable
RUN chmod +x micha-api

CMD ["./micha-api"]
