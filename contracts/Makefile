.PHONY: deploy-avs deploy-task-mailbox all

# Default target
all: deploy-avs deploy-task-mailbox

# Deploy and register AVS
deploy-avs:
	forge script script/DeployAndRegisterAVS.s.sol --rpc-url $(RPC_URL) --broadcast --verify

# Deploy Task Mailbox
deploy-task-mailbox:
	forge script script/DeployTaskMailbox.s.sol --rpc-url $(RPC_URL) --broadcast --verify

# Helper message
help:
	@echo "Available commands:"
	@echo "  make deploy-avs             - Deploy and register AVS"
	@echo "  make deploy-task-mailbox    - Deploy Task Mailbox"
	@echo "  make all                    - Run all deploy scripts"
	@echo ""
	@echo "Note: Make sure to set RPC_URL and PRIVATE_KEY in your environment or .env file" 