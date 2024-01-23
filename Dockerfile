FROM golang:1.21.5-alpine3.19 as dev
WORKDIR /app
COPY . .
RUN go mod download
ENV PROD=false
CMD ["go", "run", "main.go"]

# Prod
FROM --platform=$BUILDPLATFORM golang:1.21.5-alpine3.19 AS build
WORKDIR /src
COPY . .
RUN go mod download -x

# This is the architecture you’re building for, which is passed in by the builder.
# Placing it here allows the previous steps to be cached across architectures.
ARG TARGETARCH

# Build the application.
# Leverage a cache mount to /go/pkg/mod/ to speed up subsequent builds.
# Leverage a bind mount to the current directory to avoid having to copy the
# source code into the container.
RUN CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server .

RUN apk add --no-cache nodejs npm
RUN npm i
RUN npm run build

FROM alpine:3.19 AS final

# Install any runtime dependencies that are needed to run your application.
# Leverage a cache mount to /var/cache/apk/ to speed up subsequent builds.
RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
        ca-certificates \
        tzdata \
        && \
        update-ca-certificates

# Create a non-privileged user that the app will run under.
# See https://docs.docker.com/go/dockerfile-user-best-practices/
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser

# Copy the executable from the "build" stage.
COPY --from=build /bin/server /bin/
COPY --from=build /src/dist/ /bin/dist/

# Expose the port that the application listens on.
EXPOSE 8000

ENV PROD=true
# What the container should run when it is started.
ENTRYPOINT [ "/bin/server" ]
