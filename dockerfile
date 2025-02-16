# Use the official Golang image as the base image
FROM golang:1.20-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

# Download and cache the Go modules
RUN go mod download

# Copy the rest of the application code to the working directory
COPY . .

# Build the Go application
RUN go build -o main .

# Expose the port that the application will run on
EXPOSE 8000

# Set environment variables (if any)
ENV PORT=8000

# Command to run the executable
CMD ["./main"]

# uilding and Running the Docker Container:
# Build the Docker image using the following command:
#* docker build -t go-docker .
# Run the Docker container using the following command:
#* docker run -p 8000:8000 go-docker
# The application will be accessible at http://localhost:8000.