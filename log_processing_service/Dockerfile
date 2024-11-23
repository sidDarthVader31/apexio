#dockerfile for go api
FROM golang:1.21 
WORKDIR /app
COPY go.mod go.sum ./ 
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping
EXPOSE 3000
CMD ["/docker-gs-ping"]
