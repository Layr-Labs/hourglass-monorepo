.PHONY: build clean test

GO = $(shell which go)
BIN = ./bin

PROTO_OPTS=--proto_path=protos --go_out=paths=source_relative:protos

deps: deps/go
	./scripts/installDeps.sh

.PHONY: deps/dev
deps/dev:
	# ${GO} get github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts
	cd protos && buf dep update

.PHONY: deps
deps/go:
	${GO} mod tidy

.PHONY: proto
proto:
	buf generate protos


.PHONY: cmd/worker
cmd/worker:
	go build -o ${BIN}/worker ./cmd/worker
