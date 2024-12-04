#dockerfile for go api
FROM golang:1.23.2
WORKDIR /app
COPY go.mod go.sum ./ 
RUN go mod download 
COPY . .
RUN  GOOS=linux go build -o /source_grpc
EXPOSE 3002
CMD ["/source_grpc"]
