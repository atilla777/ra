# 1st stage image named "build"
FROM golang:1.15-alpine AS build

# Create appuser (default user is root and its no secure)
# Explicit UID allow to fix UID value in all created form this file images
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
RUN apk add --no-cache git gcc g++ sqlite sqlite-libs sqlite-dev nmap sudo openssl
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# Add app user to sudoers to allow him start nmap
RUN echo 'appuser ALL=(ALL) NOPASSWD:/usr/bin/nmap' >> /etc/sudoers

# These layers are only re-built when project dependences list are updated
WORKDIR /go/src/app

# Create dir for app binary
RUN mkdir /app

# Allow appuser use app folder
RUN chown -R appuser /app

# We don`t copy all app content (src, config e.t.c) before update packages.
# Its allow to use cache, if packages wasn`t changed.
COPY go.mod .
COPY go.sum .

# Download project dependencies with go.mod (go mudules)
RUN go mod download
RUN go mod verify

# Copy default config
COPY config.yaml /etc/ra/config.yaml

# Copy the entire project and build it
COPY . .

# This layer is rebuilt when a src files changes in the project directory
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/ra *.go

# Make private SSL key and cert (use them only for test and dev purpose)
RUN openssl req -x509 -newkey rsa:4096 -keyout /app/ra.key -out /app/ra.crt -days 365 -nodes -subj "/C=RU/ST=Moscow/L=Moscof/CN=www.example.com"

# 2nd stage - results in a single layer image based on zero sized image "scratch"
FROM scratch

# Import the user and group files from "build" image
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

# Import sudoers, sudo and nmap from "build" image
COPY --from=build /etc/sudoers /etc/sudoers
COPY --from=build /usr/bin/sudo /usr/bin/sudo
COPY --from=build /usr/bin/nmap /usr/bin/nmap

# Import binary and config from "build" image
COPY --from=build /etc/ra/config.yaml /etc/ra/config.yaml
COPY --from=build /app/ra /app/ra

# Import SSL key and cert from "build" image
COPY --from=build /app/ra.key /app/ra.key
COPY --from=build /app/ra.crt /app/ra.crt

# Port to be published
EXPOSE 1323

# Use an unprivileged user.
USER appuser:appuser

# Add volume
VOLUME /app/nmapxml /app/db

# Add path to find sudo and nmap (they called from app)
ENV PATH = "/usr/bin:${PATH}"

# Start app
ENTRYPOINT ["/app/ra"]

# You can build and run this image (container) by this command:
# sudo docker build -t ra:latest .
# sudo docker run --rm -p 1323:1323 -v db:/app/db -v nmapxml:/app/nmapxml ra:latest -it sh
