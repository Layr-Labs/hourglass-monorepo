FROM golang:1.23.6-bookworm

RUN apt-get update

# Assumes the repo root as the context
ADD . /build

WORKDIR /build/ponos

RUN make all

RUN cp bin/* /usr/local/bin

FROM debian:stable-slim

COPY --from=0 /usr/local/bin/executor /usr/local/bin/
COPY --from=0 /usr/local/bin/aggregator /usr/local/bin/
COPY --from=0 /usr/local/bin/bls-helper /usr/local/bin/
