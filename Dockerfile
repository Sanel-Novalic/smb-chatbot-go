FROM golang:1.24-alpine AS builder

# Install git and C build tools (needed for some Go packages potentially, including CGO for migrate if not disabled)
RUN apk add --no-cache git build-base

WORKDIR /app

# Copy module files and download dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the main application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /smb-chatbot main.go

# --- Install migrate CLI using go install ---
# Installs the latest v4.x version. Add @v4.x.y if you need a specific version.
# Includes '-tags postgres' to ensure postgres driver support is built-in.
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Stage 2: Create the final, minimal image
FROM alpine:latest

WORKDIR /

# Copy the compiled application binary from the builder stage
COPY --from=builder /smb-chatbot /smb-chatbot

# Copy the compiled migrate CLI from the builder stage
# Default location for go install is /go/bin/
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Copy migration files into the image
COPY db/migrations /migrations

# Expose the application port
EXPOSE 8080

# Command will be overridden by docker-compose.yml
CMD ["/smb-chatbot"]