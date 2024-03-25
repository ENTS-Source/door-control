# ---- Stage 0 ----
# Builds media repo binaries
FROM golang:1.20-bookworm AS builder

# Install build dependencies
RUN apt-get update
RUN apt-get install -y git dos2unix build-essential ca-certificates

WORKDIR /opt
COPY . /opt

# Run remaining build steps
RUN dos2unix ./build.sh && chmod 744 ./build.sh
RUN ./build.sh

# ---- Stage 1 ----
# Final runtime stage.
FROM debian:bookworm

RUN apt-get update
RUN apt-get install -y ca-certificates curl

COPY --from=builder \
 /opt/bin/controller \
 /usr/local/bin/

CMD /usr/local/bin/controller
EXPOSE 8080