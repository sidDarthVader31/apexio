# syntax=docker/dockerfile:1
FROM golang:1.23.2-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /writer ./cmd/writer

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /writer /usr/local/bin/writer
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/writer"]
