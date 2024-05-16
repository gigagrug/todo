FROM golang:alpine as dev
WORKDIR /app
COPY . .
RUN go mod download
<<<<<<< HEAD
RUN apk add --update npm
RUN npm i
CMD go run main.go
=======
RUN apk add --update gcc musl-dev
RUN CGO_ENABLED=1
CMD ["go", "run", "main.go"]

# Prod
# FROM --platform=$BUILDPLATFORM golang:alpine AS build
# WORKDIR /src
# COPY . .
# RUN go mod download -x
# ARG TARGETARCH
# RUN CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server .
#
# FROM alpine:3.19 AS final
# RUN --mount=type=cache,target=/var/cache/apk \
#     apk --update add \
#         ca-certificates \
#         tzdata \
#         && \
#         update-ca-certificates
# ARG UID=10001
# RUN adduser \
#     --disabled-password \
#     --gecos "" \
#     --home "/nonexistent" \
#     --shell "/sbin/nologin" \
#     --no-create-home \
#     --uid "${UID}" \
#     appuser
# USER appuser
# COPY --from=build /bin/server /bin/
# EXPOSE 8000
# ENTRYPOINT [ "/bin/server" ]
>>>>>>> parent of 51f5b36 (postgres)
