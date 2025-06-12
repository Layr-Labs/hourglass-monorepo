# Build
.PHONY: build
build:
	forge clean && forge build && make bindings

# Test
.PHONY: test
test:
	forge test

# Generate Contract Bindings
.PHONY: bindings
bindings:
	./bin/compile-bindings.sh

# Deploy Task Mailbox
.PHONY: deploy-task-mailbox
deploy-task-mailbox:
	forge script script/local/DeployTaskMailbox.s.sol --rpc-url $(RPC_URL) --broadcast -vvvv

# Deploy AVS L1 Contracts
.PHONY: deploy-avs-l1-contracts
deploy-avs-l1-contracts:
	forge script script/local/DeployAVSL1Contracts.s.sol --rpc-url $(RPC_URL) --broadcast --sig "run(address)" $(AVS_ADDRESS) -vvvv

# Setup AVS L1
.PHONY: setup-avs-l1
setup-avs-l1:
	forge script script/local/SetupAVSL1.s.sol --rpc-url $(RPC_URL) --broadcast --sig "run(address)" $(TASK_AVS_REGISTRAR_ADDRESS) -vvvv

# Deploy AVS L2 Contracts
.PHONY: deploy-avs-l2-contracts
deploy-avs-l2-contracts:
	forge script script/local/DeployAVSL2Contracts.s.sol --rpc-url $(RPC_URL) --broadcast -vvvv

# Setup AVS Task Mailbox Config
.PHONY: setup-avs-task-mailbox-config
setup-avs-task-mailbox-config:
	forge script script/local/SetupAVSTaskMailboxConfig.s.sol --rpc-url $(RPC_URL) --broadcast --sig "run(address, address)" $(TASK_MAILBOX_ADDRESS) $(TASK_HOOK_ADDRESS) -vvvv

# Create Task
.PHONY: create-task
create-task:
	forge script script/local/CreateTask.s.sol --rpc-url $(RPC_URL) --broadcast --sig "run(address, address)" $(TASK_MAILBOX_ADDRESS) $(AVS_ADDRESS) -vvvv

# Helper message
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make deploy-task-mailbox    - Deploy Task Mailbox"
	@echo "  make deploy-avs-l1-contracts AVS_ADDRESS=0x... - Deploy AVS L1 Contracts"
	@echo "  make setup-avs-l1 TASK_AVS_REGISTRAR_ADDRESS=0x... - Setup AVS on L1"
	@echo "  make deploy-avs-l2-contracts - Deploy AVS L2 Contracts"
	@echo "  make setup-avs-task-mailbox-config TASK_MAILBOX_ADDRESS=0x... CERTIFICATE_VERIFIER_ADDRESS=0x... TASK_HOOK_ADDRESS=0x... - Setup AVS Task Mailbox Config"
	@echo "  make generate-bls-params OPERATOR_ADDRESS=0x... CHAIN_ID=1 TASK_AVS_REGISTRAR_ADDRESS=0x... RPC_URL=http://... - Generate BLS Parameters for Operator Registration"
	@echo "  make register-operator OPERATOR_PRIVATE_KEY=0x... ALLOCATION_DELAY=0 METADATA_URI=0x... AVS_ADDRESS=0x... OPERATOR_SET_ID=0 SOCKET=0x... PUBKEY_REGISTRATION_PARAMS=0x... - Register Operator"
	@echo "  make create-task TASK_MAILBOX_ADDRESS=0x... AVS_ADDRESS=0x... - Create Task"
	@echo ""
	@echo "Note: Make sure to set RPC_URL and PRIVATE_KEY in your environment or .env file"
