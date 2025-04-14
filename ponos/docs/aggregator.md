
## Configuration

The config can be provided as a yaml or json file with the following structure:

```yaml
chains:                                         # Global list of chains this aggregator deployment supports
  - name: ethereum                                # friendly name of the chain to make reading/debugging easier
    network: mainnet                              # friendly name of the network to make reading/debugging easier
    chainId: 1                                    # the canonical chain id of the chain
    rpcUrl: "https://...."                        # the RPC URL to connect to the chain
  - name: base
    network: mainnet
    chainId: 8453
    rpcUrl: "https://...."
avss:                                           # list of AVSs to support
  - address: "0xavs1..."                        # The address of the AVS contract
    privateKey: ""                                # The private key of the AVS used for signing blockchain transactions
    privateSigningKey: ""                         # The private key of the AVS used for signing messages
    privateSigningKeyType: "ecdsa"                # The type of the private key, e.g. ecdsa, ed25519, bls, etc
    responseTimeout: 3000                         # How long the aggregator waits for a response for a scheduled task (in seconds)
    chainIds: [1]                                 # Chains to listen to for this AVS (should be supported in the list above)
```
