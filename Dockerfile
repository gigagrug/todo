FROM golang:1.21.5-alpine

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest

COPY . .

RUN go mod tidy
