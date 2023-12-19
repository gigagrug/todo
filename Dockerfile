FROM golang:1.21.5-alpine

ENV CGO_ENABLE=1

RUN apk add --no-cache \
	gcc \
	musl-dev \
	nodejs \
	npm

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest

COPY . .

RUN go mod tidy

RUN npx prisma migrate dev --name init 

CMD ["air"]
