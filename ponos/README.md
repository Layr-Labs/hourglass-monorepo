# Ponos - Aggregation and Execution

## Development

```bash
# install deps
make deps

# run the test suite
make test

# build the protos
make proto

# build all binaries
make all

# lint the project
make lint
```

## Key generation

Ponos comes with a `keygen` cli utility to make generating keys easy for testing

```bash
go run ./cmd/keygen/*.go generate --curve-type bn254 --output-dir ../testKeys --use-keystore
```
