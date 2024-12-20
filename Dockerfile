# Dockerfile

# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o wordwizardry ./main.go

# Final stage
FROM alpine:3.18

WORKDIR /app

# Add necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy static files
COPY --from=builder /app/public ./public

# Copy binary from builder
COPY --from=builder /app/wordwizardry .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./wordwizardry"]