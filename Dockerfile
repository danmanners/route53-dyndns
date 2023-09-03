# Build the Go Binary then copy it into a scratch container
FROM docker.io/library/golang:1.21-alpine AS builder

# Create the user & map the home directory
RUN mkdir -p /app \
    && adduser -u 10001 -DHh /app appuser \
    && chown -R appuser /app

# Install and run UPX
RUN apk add --no-cache upx ca-certificates


# Switch to the appuser
USER appuser

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum r53-dyndns.go ./

# Download all dependencies, Build the App, and Run UPX
RUN go mod download \
    && CGO_ENABLED=0 GOOS=linux go build -o r53-dyndns . \
    && upx -9 /app/r53-dyndns

# Start a new stage from scratch
FROM scratch

# Copy the /etc/passwd file from the previous stage
COPY --from=builder /etc/passwd /etc/passwd

# Copy the /etc/group file from the previous stage
COPY --from=builder /etc/group /etc/group

# Copy the SSL Certificates from the previous stage
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/

# Set the User ID for the container
USER 10001

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/r53-dyndns /app/r53-dyndns

# Command to run the executable
ENTRYPOINT ["/app/r53-dyndns"]
