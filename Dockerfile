FROM golang:1.21.5-alpine

ENV CGO_ENABLE=1

RUN apk add --no-cache \
	nodejs \
	npm

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest

COPY . .

RUN go mod tidy

CMD ["sh", "-c", "npx prisma migrate dev --name init && air"]
