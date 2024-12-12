FROM golang:1.23.2-alpine

WORKDIR /app
# Copy go mod and sum files
COPY go.mod  ./ 
RUN go mod download

COPY . .


RUN GOOS=linux go build -o /main  

# Build the application
# RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Run the binary
CMD ["/main"]
