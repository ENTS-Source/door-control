# ---- Stage 0 ----
# Builds media repo binaries
FROM golang:1.20-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git musl-dev dos2unix build-base

WORKDIR /opt
COPY . /opt

# Run remaining build steps
RUN dos2unix ./build.sh && chmod 744 ./build.sh
RUN ./build.sh

# ---- Stage 1 ----
# Final runtime stage.
FROM alpine:latest

RUN mkdir /plugins
RUN apk add --no-cache \
        su-exec \
        ca-certificates \
        dos2unix

COPY --from=builder \
 /opt/bin/controller \
 /usr/local/bin/

CMD /usr/local/bin/controller
EXPOSE 8000