
.PHONY: build
build:
	go build -o ./bin/ponos-performer main.go

build-container:
	cd ../ && docker build -t hello-performer:latest -f demo/Dockerfile .
