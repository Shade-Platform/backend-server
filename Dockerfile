# Start with the official Golang base image
FROM golang:1.23.2

# Install reflex for file watching and reloading
RUN go install github.com/cespare/reflex@latest

# Set the current working directory inside the container
WORKDIR /app

# Copy the Go module files (go.mod and go.sum) and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code into the container
COPY . .

# Expose the application port
EXPOSE 8080

# Command to run the application with reflex for live-reload
CMD ["reflex", "-s", "-r", "\\.go$", "--", "go", "run", "main.go"]
# CMD ["reflex", "-s", "-r", "\\.go$", "--", "go", "run", "main.go", "migrate", "up"]

# Run main.go
# CMD ["go", "run", "main.go"]
