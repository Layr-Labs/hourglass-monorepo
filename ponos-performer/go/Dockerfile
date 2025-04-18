FROM golang:1.23-bullseye

ADD . /build

WORKDIR /build/ponos-performer/go

RUN ls -al && go build -o ponos-performer ./example/*.go

EXPOSE 8080

CMD ["./ponos-performer"]
