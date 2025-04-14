## Configuration

The config can be provided as a yaml or json file with the following structure:

```yaml
operator:
  address: "0xoperator..."            # The operator address
  operatorPrivateKey: "..."           # The private key of the operator used for signing blockchain transactions
  signingKeys:                        # available signing keys, only one key per type is needed and will depend on what is required by the AVS
    ecdsa:
      privateKey: ""
    bls:
      privateKey: ""
avss:                                 # define which AVSs to run
  - image:                              # docker container image to run
      repository: "eigenlabs/avs"         # the repository of the docker image along with the image name
      tag: "v1.0.0"                       # the tag of the image to run
    processType: "server"               # the type of process to run, e.g. server, one-off, etc
    avsAddress: "0xavs1..."             # The address of the AVS contract (mainly to identify tasks to parse and process)
```
