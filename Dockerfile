# 1st stage image named "build"
FROM golang:1.15-alpine AS build

# Create appuser (default user is root and its no secure).
ENV USER=appuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" "${USER}"

# Install tools required for project
# Install git and don`t make local alpine package index cache
RUN apk update && apk upgrade
RUN apk add --no-cache git gcc g++ sqlite sqlite-libs sqlite-dev
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# These layers are only re-built when project dependences list are updated
WORKDIR /go/src/app

# Copy the entire project and build it
COPY . .

# Download project dependencies with go.mod (go mudules)
RUN go mod download
RUN go mod verify

# Copy default config
COPY config.yaml /etc/ra/config.yaml

# This layer is rebuilt when a file changes in the project directory
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /bin/app *.go

# 2nd stage - results in a single layer image based on zero sized image "scratch"
FROM scratch

# Import the user and group files from the builder.
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

# Coby binary and config from "build" image
COPY --from=build /bin/app /bin/app
COPY --from=build /etc/ra/config.yaml /etc/ra/config.yaml

# Port to be published
EXPOSE 1323

# RUN chown appuser /bin/app

# Use an unprivileged user.
USER appuser:appuser
ENTRYPOINT ["/bin/app"]
