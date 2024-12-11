FROM golang:1.23.2-alpine

WORKDIR /app

# Copy go mod and sum files
COPY go.mod  ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Run the binary
CMD ["/app/main"]
