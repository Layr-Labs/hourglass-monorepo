# AVS Contracts Script Setup

This README provides step-by-step instructions for setting up and deploying the AVS contracts using the scripts in this directory. These commands are intended for local development and testing using a forked Ethereum mainnet environment.

## Prerequisites

- [Foundry](https://book.getfoundry.sh/) installed
- [Anvil](https://book.getfoundry.sh/anvil/) (comes with Foundry)
- `make` installed
- Access to a mainnet RPC URL (e.g., from [QuikNode](https://quiknode.io/), [Alchemy](https://www.alchemy.com/), or [Infura](https://infura.io/))

## Step-by-Step Setup

### 1. Start a Forked Mainnet Node

Start a local Anvil node forked from Ethereum mainnet at a specific block (This includes the Eigenlayer Protocol already deployed):

```sh
anvil --fork-url <HOLESKY_RPC_URL> --fork-block-number 3987755
```

This will run a local node at `127.0.0.1:8545`.
Update the `.env` with the correct private keys from the anvil bootup message.

### 2. Deploy TaskMailbox Contract

In a new terminal, deploy the TaskMailbox contract to your local fork:

```sh
make deploy-task-mailbox RPC_URL="127.0.0.1:8545"
```

### 3. Deploy AVS L1 Contracts

Deploy the AVS L1 contracts, specifying the AVS address:

```sh
make deploy-avs-l1-contracts AVS_ADDRESS='0x70997970C51812dc3A010C7d01b50e0d17dc79C8' RPC_URL="127.0.0.1:8545"
```

### 4. Setup AVS L1

Register the AVS L1 contracts with the EigenLayer core protocol:

```sh
make setup-avs-l1 TASK_AVS_REGISTRAR_ADDRESS='0x050F5F4De07C1298162a675AdA47A5B65bCB32F4' RPC_URL="127.0.0.1:8545"
```

### 5. Deploy AVS L2 Contracts

Deploy the AVS L2 contracts:

```sh
make deploy-avs-l2-contracts RPC_URL="127.0.0.1:8545"
```

### 6. Setup AVS Task Mailbox Config

Configure the TaskMailbox with the required addresses:

```sh
make setup-avs-task-mailbox-config \
  TASK_MAILBOX_ADDRESS='0xdef31014c9886bd715f4055a140151a7296C0e18' \
  TASK_HOOK_ADDRESS='0xac81E70942DF7b08238Cb09572b74F400AdE24C1' \
  RPC_URL="127.0.0.1:8545"
```

### 7. Create Task

Create a Task on L2 Mailbox:

```sh
make create-task \
  TASK_MAILBOX_ADDRESS='0xdef31014c9886bd715f4055a140151a7296C0e18' \
  AVS_ADDRESS='0x70997970C51812dc3A010C7d01b50e0d17dc79C8' \
  RPC_URL="127.0.0.1:8545"
```

## Environment Variables

Some scripts require environment variables, such as `PRIVATE_KEY_AVS`, to be set. Refer to the `.env.example` file in the parent directory for required variables and copy it as `.env`. Get the private keys from the Anvil bootup message.

## Script Reference

- `DeployTaskMailbox.s.sol`: Deploys the TaskMailbox contract
- `DeployAVSL1Contracts.s.sol`: Deploys AVS L1 contracts
- `SetupAVSL1.s.sol`: Registers AVS L1 contracts
- `DeployAVSL2Contracts.s.sol`: Deploys AVS L2 contracts
- `SetupAVSTaskMailboxConfig.s.sol`: Configures the TaskMailbox with AVS and verifier addresses
- `RegisterOperator.s.sol`: Registers an operator to the EigenLayer Core Protocol and operator set
- `CreateTask.s.sol`: Creates a Task on L2 Mailbox
- `generate_bls_params.go`: Generates BLS parameters for operator registration

## Additional Notes

- Ensure your local node is running before executing deployment or setup commands.
- The provided addresses are for local development and testing only.
- For more details on each script, review the source files in this directory. 