# Hourglass AVS Deployment Starter Guide

## Introduction

This guide will walk you through deploying a task-based AVS on Sepolia that uses the Multichain Service, enabling the AVS to function on Base Sepolia. We will cover an end-to-end deployment that leverages Devkit to deploy the Middleware Contracts which define your AVS, hook them up to the EigenLayer Core Contracts, and use hgctl to create and manage Operator resources.

By the end of this guide you will have completed:
- Deployment of a task-based AVS on Sepolia that is verifiable on Base Sepolia
- Creation of Operator Sets for the AVS
- Configuring the Operator Sets for the Multichain Service
- Creation and registration of Operators to EigenLayer and the task-based AVS
- Using the Multichain Service to verify the work of an Operator on Base Sepolia

## Architecture Overview

This architecture leverages the Hourglass framework, which enables the creation, orchestration, execution and validation of tasks - the unit of work for task-based AVSs. For a deeper understanding of the Hourglass framework and its architecture, refer to the [documentation](https://docs.eigenlayer.xyz).

### Components

#### Onchain Components:
- **EigenLayer**: The EigenLayer contracts make up the EigenLayer protocol on Sepolia. Operators will register for EigenLayer and AVSs through these contracts.
- **ReleaseCoordinator**: This contract manages software releases and coordinates upgrades for Operators within an AVS. An AVS will publish software artifacts to this contract to be run by Operators. Similar to an onchain container registry.
- **TaskAVSRegistrar**: This contract plugs into the EigenLayer core contracts and is responsible for registering Operators to the AVS and configuring the Operator Sets handling the tasks. You will deploy this contract.
- **AVSTaskHook**: This contract allows the AVS to define their task-handling logic through hooks like pre and post task validation. You will deploy this contract.
- **TaskMailbox**: This contract manages the lifecycle of task requests for the AVS, enabling the creation and verification of tasks.

#### Offchain Components:
- **Operator Set 0 (Aggregator)**: The Aggregator is responsible for listening for task submissions on the TaskMailbox, distributing that work to Executors (Operator Set 1) and submitting the results back to the TaskMailbox.
- **Operator Set 1 (Executor)**: The Executor is responsible for executing the work of the task and returning the result back to the Aggregator (Operator Set 0).
- **Multichain Service**: A service run by Eigen Labs that transports EigenLayer state on Sepolia to Base Sepolia.

### Task Flow

To understand this architecture better, here's the flow of a task:

1. The user submits a transaction containing a task request to the TaskMailbox, emitting an event
2. The Aggregator consumes this task and sends the work to the Executor
3. The Executor receives the work, performs the computation, signs the result and returns it to the Aggregator
4. The Aggregator submits the final result to the TaskMailbox

## Requirements

Before we begin, you will need two funded accounts, RPCs and installed software:

### Required Resources:
- **Sepolia and Base Sepolia RPC endpoints**
- **Sepolia Testnet ETH** (Get from faucets):
  - https://www.alchemy.com/faucets/ethereum-sepolia
  - https://cloud.google.com/application/web3/faucet/ethereum/sepolia
  - https://www.infura.io/zh/faucet/sepolia
- **Sepolia Testnet WETH**: 
  - Wrap your Sepolia ETH by depositing it in the [WETH contract](https://sepolia.etherscan.io/address/0x7b79995e5f793A07Bc00c21412e50Ecae098E7f9)

### Required Software:
- [Devkit CLI](https://github.com/Layr-Labs/devkit-cli) for AVS contract deployment
- [hgctl](https://github.com/Layr-Labs/hourglass-monorepo/tree/master/hgctl-go) for operator management
- [Docker](https://docs.docker.com/engine/install/) (latest version)
- A valid Github access token
- Logged into ghcr (Github Container Registry) - [Instructions](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)

## Step 1: Deploying the Contracts with Devkit

The first step is deploying and setting up the AVS contracts with Devkit. By the end of this step you will have:
- Deployed the TaskAVSRegistrar and AVSTaskHook
- Registered your AVS with EigenLayer
- Created and configured the Aggregator and Executor Operator Sets

### Installation and Setup

1. **Install devkit** following the steps outlined [here](https://github.com/Layr-Labs/devkit-cli#installation).

2. **Create your AVS project**:
   ```bash
   devkit avs create sepolia-deployment
   cd sepolia-deployment
   ```

3. **Create a context** to manage the deployment:
   ```bash
   devkit avs context create --context testnet
   ```

4. **Configure testnet.yaml** - Edit the following fields:
   ```yaml
   deployer_private_key: # Your funded Sepolia account private key
   l1:
     chain_id: 11155111  # Sepolia
     fork_url: # Your Sepolia RPC URL
   l2:
     chain_id: 84532    # Base Sepolia
     fork_url: # Your Base Sepolia RPC URL
   avs_private_key: # Private key for AVS identity
   address: # Public key associated with avs_private_key
   stakers: []
   operators: []
   ```

5. **Set the testnet context as current**:
   ```bash
   devkit avs config --set project.context="testnet"
   ```

### Deploy Contracts

6. **Deploy L1 contracts and set up the AVS**:
   ```bash
   devkit avs deploy contracts l1
   ```
   
   This command performs multiple operations:
   - Deploys the TaskAVSRegistrar on Sepolia
   - Sets the metadata URI for the AVS
   - Sets the AVS Registrar for your AVS
   - Creates Operator Sets
   - Configures the curve type (ECDSA or BN254) for the Operator Sets
   - Registers the Operator Set for the Multichain Service

7. **Wait for Multichain Service** to pick up your new AVS (approximately 5-10 minutes)

8. **Deploy L2 contracts**:
   ```bash
   devkit avs deploy contracts l2
   ```

> **Note**: The context file is the main configuration for your AVS. It allows control over what Operators and Stakers to register for your AVS as well as the strategies (stake tokens) that your Operator Set supports.

## Step 2: Publishing Container Release with Devkit

Once the contracts are deployed, we need to make the software to be run by Operators available. To do this, we will publish software artifacts to the ReleaseCoordinator contract using Devkit.

By the end of this step you will have:
- Uploaded software artifacts for Operators to ReleaseManager contract
- Coordinated an upgrade for Operators on the ReleaseManager contract

### Steps:

1. **Authenticate with Github Container Registry**:
   ```bash
   echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin
   ```

2. **Set metadata URI for your Operator Sets**:
   ```bash
   devkit avs release uri --metadata-uri "https://example.com/metadata.json" --operator-set-id 0
   devkit avs release uri --metadata-uri "https://example.com/metadata.json" --operator-set-id 1
   ```

3. **Build the AVS**:
   ```bash
   devkit avs build
   ```

4. **Publish the release**:
   ```bash
   devkit avs release publish --upgrade-by-time 1850000000 --registry YOUR_REGISTRY_PACKAGE
   ```

5. **Make the registry public** in your Github repository settings

## Step 3: Create and Register Operator with hgctl

Now that the AVS is deployed onchain with the Operator Sets created and configured, we need to create an Operator and register it with the AVS.

By the end of this step you will have:
- Created Operator keys
- Registered the Operator into EigenLayer
- Allocated stake to an Operator
- Registered the Operator's keys
- Allocated and Registered the Operator for the AVS's Operator Set

### Operator Creation Steps:

1. **Install hgctl** following the steps [here](https://github.com/Layr-Labs/hourglass-monorepo/tree/master/hgctl-go#installation):
   ```bash
   git clone https://github.com/Layr-Labs/hourglass-monorepo
   cd hourglass-monorepo/hgctl-go
   make install
   ```

2. **Create a context** to manage deployment environment:
   ```bash
   hgctl context create sepolia
   ```
   Follow the prompted steps, using the same RPC URLs from Devkit.

3. **Create ECDSA and BLS keypairs** for the Operator:
   ```bash
   hgctl keystore create --name tutorial-operator --type ecdsa
   hgctl keystore create --name tutorial-operator-bls --type bn254
   ```

4. **Configure the signer** to manage the private keys:
   ```bash
   hgctl signer operator keystore --keystore-name tutorial-operator
   hgctl signer system keystore --keystore-name tutorial-operator-bls --type bn254
   ```

5. **Set Operator and AVS addresses** for the current context:
   ```bash
   hgctl context set --avs-address <The same AVS address from Devkit>
   hgctl context set --operator-address <The ECDSA public key of the Operator>
   ```

6. **Fund the Operator address** with Sepolia ETH and WETH:
   - Send Sepolia ETH to the operator address
   - Use the [deposit function](https://sepolia.etherscan.io/address/0x7b79995e5f793A07Bc00c21412e50Ecae098E7f9#writeContract) to get WETH

7. **Export the operator private key** for use:
   ```bash
   hgctl keystore show --name tutorial-operator
   export OPERATOR_PRIVATE_KEY=<ECDSA PRIVATE KEY>
   ```

### Operator Registration Steps:

1. **Register with EigenLayer**:
   ```bash
   hgctl el register-operator --metadata-uri https://example.com/operator/metadata.json --allocation-delay 0
   ```

2. **Self-delegate stake** (required after registration):
   ```bash
   hgctl el delegate
   ```

3. **Deposit WETH into the WETH strategy**:
   ```bash
   hgctl el deposit \
     --strategy 0x424246eF71b01ee33aA33aC590fd9a0855F5eFbc \
     --token-address 0x7b79995e5f793A07Bc00c21412e50Ecae098E7f9 \
     --amount '0.00001 ether'
   ```

4. **Register the Operator's key for AVS Operator Set 0**:
   ```bash
   hgctl el register-key \
     --operator-set-id 0 \
     --key-type bn254 \
     --keystore-path ~/.hgctl/sepolia/keystores/tutorial-operator-bls/key.json \
     --password test
   ```

5. **Register the Operator for AVS Operator Set 0**:
   ```bash
   hgctl el register-avs \
     --operator-set-ids 0 \
     --socket https://operator.example.com:8080
   ```

6. **Wait 15-20 minutes** for the configuration delay on the AllocationManager

7. **Allocate Stake to AVS Operator Set 0**:
   ```bash
   hgctl el allocate \
     --operator-set-id 0 \
     --strategy 0x424246eF71b01ee33aA33aC590fd9a0855F5eFbc \
     --magnitude 1e13
   ```

## Step 4: Deploy Operator Infrastructure with hgctl

The next step is to deploy the software to be run by the Operators. This will pull the container images stored in the ReleaseManager contract.

### Deploy Aggregator (Operator Set 0)

1. **Ensure Docker is running**

2. **Deploy the aggregator**:
   ```bash
   hgctl deploy aggregator --operator-set-id 0
   ```

### Deploy Executor (Operator Set 1)

For the executor, you'll need to register a second operator or configure the same operator for Operator Set 1:

1. **Create additional keys if needed**:
   ```bash
   hgctl keystore create --name tutorial-operator-bls-02 --type bn254
   ```

2. **Register key for Operator Set 1**:
   ```bash
   hgctl el register-key \
     --operator-set-id 1 \
     --key-type bn254 \
     --keystore-path ~/.hgctl/sepolia/keystores/tutorial-operator-bls-02/key.json \
     --password test
   ```

3. **Set the operator set context**:
   ```bash
   hgctl context set --operator-set-id 1
   ```

4. **Deploy the executor**:
   ```bash
   hgctl deploy executor --operator-set-id 1
   ```

## Step 5: Using the Multichain Service

The Multichain Service enables cross-chain AVS operations between Sepolia and Base Sepolia. Once your infrastructure is deployed:

1. **Submit tasks** to the TaskMailbox on either L1 (Sepolia) or L2 (Base Sepolia)
2. **Monitor task execution** through the aggregator logs:
   ```bash
   docker logs hgctl-aggregator-<avs-address>
   ```
3. **Verify task completion** on the destination chain

### Monitoring Your Deployment

Check the status of your deployed components:

```bash
# List running containers
docker ps

# View aggregator logs
docker logs -f hgctl-aggregator-<avs-address>

# View executor logs
docker logs -f hgctl-executor-<avs-address>

# Check performer status
hgctl get performer
```

## Troubleshooting

### Common Issues and Solutions

**"Required addresses not configured"**
- Ensure all required addresses are set in your context:
  ```bash
  hgctl context show
  ```

**"Executor not available"**
- Check that the executor container is running:
  ```bash
  docker ps | grep executor
  ```

**"Transaction failed"**
- Enable verbose mode for detailed error information:
  ```bash
  hgctl --verbose el register-operator
  ```
- Check operator balance:
  ```bash
  cast balance $OPERATOR_ADDRESS
  ```

**"Allocation delay not met"**
- Wait for the configured allocation delay period (15-20 minutes on testnet)
- Check the current allocation delay:
  ```bash
  hgctl el set-allocation-delay
  ```

### Getting Help

For additional support:
- Open an issue in the [GitHub repository](https://github.com/Layr-Labs/hourglass-monorepo/issues)
- Join the [Discord community](https://discord.gg/eigenlayer)
- Check the [documentation](https://docs.eigenlayer.xyz)

## Next Steps

Now that you have a working AVS deployment:

1. **Customize your AVS logic** by modifying the AVSTaskHook contract
2. **Add more operators** to increase decentralization
3. **Implement custom task types** for your specific use case
4. **Monitor performance** and optimize your infrastructure
5. **Prepare for mainnet deployment** by testing thoroughly on testnet

## Additional Resources

- [Hourglass Framework Documentation](https://docs.eigenlayer.xyz)
- [EigenLayer Developer Guide](https://docs.eigenlayer.xyz/developers)
- [Devkit CLI Documentation](https://github.com/Layr-Labs/devkit-cli)
- [hgctl Reference](https://github.com/Layr-Labs/hourglass-monorepo/tree/master/hgctl-go)