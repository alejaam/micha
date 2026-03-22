FROM golang:1.24

# Install PostgreSQL client for migrations
RUN apt-get update && \
    apt-get install -y postgresql-client && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy application code
COPY . .

# Make start script executable
RUN chmod +x /app/start.sh

CMD ["/app/start.sh"]
