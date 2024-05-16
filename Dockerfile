FROM golang:alpine as dev
WORKDIR /app
COPY . .
RUN go mod download
RUN apk add --update npm
RUN cd /app/frontend && npm i
CMD go run main.go
