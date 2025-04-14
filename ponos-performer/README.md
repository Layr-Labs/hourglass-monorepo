## HTTP/JSON interface

Task payload request
```json
{
    "taskId": "0x...",
    "avs": "0xavs1...",
    "operatorSetId": 1234,
    "metadata": "...", // base64 encoded bytes
    "payload": "...", // base64 encoded bytes"
}
```

Task payload response
```json
{
	"taskId": "0x...",
	"avs": "0xavs1...",
	"operatorSetId": 1234,
    "payload": "..." // base64 encoded bytes
}
```

