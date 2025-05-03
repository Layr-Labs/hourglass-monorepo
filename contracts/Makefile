.PHONY: deploy-task-mailbox deploy-avs-l1-contracts setup-avs-l1 deploy-avs-l2-contracts setup-avs-task-mailbox-config create-task

# Deploy Task Mailbox
deploy-task-mailbox:
	forge script script/DeployTaskMailbox.s.sol --rpc-url $(RPC_URL) --broadcast --verify

# Deploy AVS L1 Contracts
deploy-avs-l1-contracts:
	forge script script/DeployAVSL1Contracts.s.sol --rpc-url $(RPC_URL) --broadcast --verify --sig "run(address)" $(AVS_ADDRESS)

# Setup AVS L1
setup-avs-l1:
	forge script script/SetupAVSL1.s.sol --rpc-url $(RPC_URL) --broadcast --verify --sig "run(address)" $(TASK_AVS_REGISTRAR_ADDRESS)

# Deploy AVS L2 Contracts
deploy-avs-l2-contracts:
	forge script script/DeployAVSL2Contracts.s.sol --rpc-url $(RPC_URL) --broadcast --verify

# Setup AVS Task Mailbox Config
setup-avs-task-mailbox-config:
	forge script script/SetupAVSTaskMailboxConfig.s.sol --rpc-url $(RPC_URL) --broadcast --verify --sig "run(address, address, address)" $(TASK_MAILBOX_ADDRESS) $(CERTIFICATE_VERIFIER_ADDRESS) $(TASK_HOOK_ADDRESS)

# Create Task
create-task:
	forge script script/CreateTask.s.sol --rpc-url $(RPC_URL) --broadcast --verify --sig "run(address, address)" $(TASK_MAILBOX_ADDRESS) $(AVS_ADDRESS)

# Helper message
help:
	@echo "Available commands:"
	@echo "  make deploy-task-mailbox    - Deploy Task Mailbox"
	@echo "  make deploy-avs-l1-contracts AVS_ADDRESS=0x... - Deploy AVS L1 Contracts"
	@echo "  make setup-avs-l1 TASK_AVS_REGISTRAR_ADDRESS=0x... - Setup AVS on L1"
	@echo "  make deploy-avs-l2-contracts - Deploy AVS L2 Contracts"
	@echo "  make setup-avs-task-mailbox-config TASK_MAILBOX_ADDRESS=0x... CERTIFICATE_VERIFIER_ADDRESS=0x... TASK_HOOK_ADDRESS=0x... - Setup AVS Task Mailbox Config"
	@echo "  make create-task TASK_MAILBOX_ADDRESS=0x... AVS_ADDRESS=0x... - Create Task"
	@echo ""
	@echo "Note: Make sure to set RPC_URL and PRIVATE_KEY in your environment or .env file" 