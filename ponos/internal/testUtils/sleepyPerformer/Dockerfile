FROM golang:1.23-bullseye

ADD . /ponos
WORKDIR /ponos/internal/testUtils/sleepyPerformer

RUN make build
RUN cp bin/* /usr/local/bin

EXPOSE 8080

CMD ["/usr/local/bin/ponos-performer"]