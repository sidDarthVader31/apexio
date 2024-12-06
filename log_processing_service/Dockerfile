#dockerfile for go api
FROM golang:1.23.2 
WORKDIR /app
COPY go.mod go.sum ./ 
RUN go mod download
COPY . .
RUN GOOS=linux go build -o /log_processing_service  
EXPOSE 3003
CMD ["/log_processing_service"]
