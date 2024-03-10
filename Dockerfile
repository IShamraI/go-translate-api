# Use the official Golang image to create a build artifact.
FROM golang:1.16 AS builder

# Create a working directory inside the container.
WORKDIR /app

# Copy the Go modules files.
COPY go.mod go.sum ./

# Download the Go dependencies.
RUN go mod download

# Copy the source code into the container.
COPY . .

# Build the Go application.
RUN go build -o app .

# Use a minimal base image to reduce the final container size.
FROM alpine:latest

# Copy the binary from the builder stage to the new image.
COPY --from=builder /app/app /app/app

# Set the working directory inside the container.
WORKDIR /app

# Expose the port on which the Go application will run.
EXPOSE 8080

# Command to run the Go application.
CMD ["./app"]
