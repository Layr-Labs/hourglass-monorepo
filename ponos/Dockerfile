FROM golang:1.23.6-bookworm

RUN apt-get update

# Assumes the repo root as the context
ADD . /build

WORKDIR /build/ponos

RUN make all


FROM alpine:latest

COPY --from=0 /build/ponos/bin/executor /usr/local/bin/
COPY --from=0 /build/ponos/bin/aggregator /usr/local/bin/


