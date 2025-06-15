FROM golang:1.22

WORKDIR /app

# Install CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon@latest

# Copy source code
COPY . .

# Run CompileDaemon for hot reload (will be set by compose)
