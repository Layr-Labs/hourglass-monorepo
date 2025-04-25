FROM golang:1.23.6-bookworm

RUN apt-get update

# Assumes the repo root as the context
ADD . /build

WORKDIR /build/ponos

RUN make all

RUN cp bin/* /usr/local/bin


