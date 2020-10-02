FROM golang:1.15.2-buster AS build

# Create appuser (default user is root and it`s not secure)
# Explicit UID allow to fix UID value in all created form this Dockerfile images
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
RUN  apt-get update \
  && apt-get install -y \
  git \
  gcc \
  g++ \
  sqlite \
  libsqlite3-dev \
  nmap \
  sudo \
  openssl \
  && rm -rf /var/lib/apt/lists/*

# Add app user to sudoers to allow him start nmap
RUN echo 'appuser ALL=(ALL) NOPASSWD:/usr/bin/nmap' >> /etc/sudoers

# Make dirs for binary app and config and for volumes mountpoints
RUN mkdir /app /db /nmapxml
RUN chmod 07777 /app && chmod 07777 /nmapxml && chmod 07777 /db

# Copy default config
COPY docker.config.yaml /app/docker.config.yaml

# These layers are only re-built when project dependences list are updated
WORKDIR /go/src/app

# We don`t copy all app content (src, config e.t.c) before update packages.
# Its allow to use cache, if packages wasn`t changed.
COPY go.mod .
COPY go.sum .

# Download project dependencies with go.mod (go mudules)
RUN go mod download
RUN go mod verify

# Copy the entire project and build it
COPY . .

# Allow appuser to access app foldres
RUN chown -R 10001:10001 /app

# This layer is rebuilt when a src files changes in the project directory
 RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/ra *.go

# Make private SSL key and cert (use them only for test and dev purpose)
RUN openssl req -x509 -newkey rsa:4096 -keyout /app/ra.key -out /app/ra.crt -days 365 -nodes -subj "/C=RU/ST=Moscow/L=Moscof/CN=www.example.com"

# 2nd stage

# Results in a single layer image based on zero sized image "scratch"
FROM scratch

# Import the user and group files from "build" image
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

# Import sudoers, sudo and nmap from "build" image
COPY --from=build /etc/sudoers /etc/sudoers
COPY --from=build /usr/bin/sudo /usr/bin/sudo
COPY --from=build /usr/bin/nmap /usr/bin/nmap

# Import binary and config from "build" image
COPY --chown=10001:10001 --from=build /app /app

# Import volume mountpoints
COPY --chown=10001:10001 --from=build /nmapxml /nmapxml
COPY --chown=10001:10001 --from=build /db /db

# Resolve CGO (sqlite) problem
COPY --from=build /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2
COPY --from=build /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc.so.6
COPY --from=build /lib/x86_64-linux-gnu/libm.so.6 /lib/x86_64-linux-gnu/libm.so.6
COPY --from=build /lib/x86_64-linux-gnu/libdl.so.2 /lib/x86_64-linux-gnu/libdl.so.2
COPY --from=build /lib/x86_64-linux-gnu/libpthread.so.0 /lib/x86_64-linux-gnu/libpthread.so.0

# Resolve nmap dependences
COPY --from=build /lib/x86_64-linux-gnu/libpcre.so.3 /lib/x86_64-linux-gnu/libpcre.so.3
COPY --from=build /usr/lib/x86_64-linux-gnu/libpcap.so.0.8 /usr/lib/x86_64-linux-gnu/libpcap.so.0.8
COPY --from=build /usr/lib/x86_64-linux-gnu/libssh2.so.1 /usr/lib/x86_64-linux-gnu/libssh2.so.1
COPY --from=build /usr/lib/x86_64-linux-gnu/libssl.so.1.1 /usr/lib/x86_64-linux-gnu/libssl.so.1.1
COPY --from=build /usr/lib/x86_64-linux-gnu/libcrypto.so.1.1 /usr/lib/x86_64-linux-gnu/libcrypto.so.1.1
COPY --from=build /lib/x86_64-linux-gnu/libz.so.1 /lib/x86_64-linux-gnu/libz.so.1
COPY --from=build /lib/x86_64-linux-gnu/libz.so.1 /lib/x86_64-linux-gnu/libz.so.1
COPY --from=build /usr/lib/x86_64-linux-gnu/liblua5.3.so.0 /usr/lib/x86_64-linux-gnu/liblua5.3.so.0
COPY --from=build /usr/lib/x86_64-linux-gnu/liblinear.so.3 /usr/lib/x86_64-linux-gnu/liblinear.so.3
COPY --from=build /usr/lib/x86_64-linux-gnu/libstdc++.so.6 /usr/lib/x86_64-linux-gnu/libstdc++.so.6
COPY --from=build /lib/x86_64-linux-gnu/libgcc_s.so.1 /lib/x86_64-linux-gnu/libgcc_s.so.1
COPY --from=build /lib/x86_64-linux-gnu/libgcrypt.so.20 /lib/x86_64-linux-gnu/libgcrypt.so.20
COPY --from=build /usr/lib/x86_64-linux-gnu/libblas.so.3 /usr/lib/x86_64-linux-gnu/libblas.so.3
COPY --from=build /lib/x86_64-linux-gnu/libgpg-error.so.0 /lib/x86_64-linux-gnu/libgpg-error.so.0
COPY --from=build /usr/lib/x86_64-linux-gnu/libgfortran.so.5 /usr/lib/x86_64-linux-gnu/libgfortran.so.5
COPY --from=build /usr/lib/x86_64-linux-gnu/libquadmath.so.0 /usr/lib/x86_64-linux-gnu/libquadmath.so.0

# Resolve sudo dependences
COPY --from=build /lib/x86_64-linux-gnu/libaudit.so.1 /lib/x86_64-linux-gnu/libaudit.so.1
COPY --from=build /lib/x86_64-linux-gnu/libselinux.so.1 /lib/x86_64-linux-gnu/libselinux.so.1
COPY --from=build /lib/x86_64-linux-gnu/libutil.so.1 /lib/x86_64-linux-gnu/libutil.so.1
COPY --from=build /usr/lib/sudo/libsudo_util.so.0 /usr/lib/sudo/libsudo_util.so.0
COPY --from=build /lib/x86_64-linux-gnu/libcap-ng.so.0 /lib/x86_64-linux-gnu/libcap-ng.so.0
COPY --from=build /lib/x86_64-linux-gnu/libpcre.so.3 /lib/x86_64-linux-gnu/libpcre.so.3

# Port to be published
EXPOSE 1323

# Use an unprivileged user.
USER 10001:10001

# Add volume
VOLUME /nmapxml /db

# Add path to find sudo and nmap (they called from app)
ENV PATH = "/usr/bin:${PATH}"

WORKDIR /app

# Start app
ENTRYPOINT ["/app/ra", "-conf-name", "docker.config.yaml"]

# You can build and run this image (container) by this command:
# sudo docker build -t ra:latest .
# sudo docker run --rm -p 1323:1323 -v db:/app/db -v nmapxml:/app/nmapxml ra:latest
