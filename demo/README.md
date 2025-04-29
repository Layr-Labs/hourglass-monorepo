# Hourglass demo

This demo shows how the Ponos architecture runs with an AVS performer.

## Running

1. Build the AVS Performer

```bash
make build-container
```

2. Run the Ponos stack

```bash
docker compose up
```

3. Send a task to the Executor for execution

_Make sure you have `grpcurl` installed. You can install it using `brew install grpcurl`_

```bash
grpcurl -plaintext -d '{ "avsAddress": "0xavs1...", "taskId": "0xtask1...", "payload": "eyAibnVtYmVyVG9CZVNxdWFyZWQiOiA0IH0=" }' localhost:9090 eigenlayer.hourglass.v1.ExecutorService/SubmitTask
```
