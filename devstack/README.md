# DevStack: CDK Infrastructure for EigenLayer AVS Development

DevStack is a CDK-based infrastructure stack that provisions AWS resources to run an EigenLayer AVS development environment. It leverages the DevKit CLI's built-in `avs devnet` command to orchestrate a complete development environment.

## Overview

DevStack provides a one-click deployment solution for running EigenLayer AVS development environments in AWS. It creates an EC2 instance that automatically:

- Installs and configures DevKit CLI v0.0.8
- Creates and builds an AVS project using `devkit avs create` and `devkit avs build`
- Runs `devkit avs devnet start` which orchestrates:
  - Anvil (local Ethereum node)
  - Aggregator service
  - Executor service
  - All necessary protocol bootstrapping
- Exposes gRPC and RPC endpoints for remote access

## Prerequisites

- AWS CLI configured with appropriate credentials
- Node.js 20+ and npm
- AWS CDK CLI (`npm install -g aws-cdk`)
- AWS account with permissions to create EC2, IAM, and VPC resources

## Quick Start

1. **Install dependencies:**
   ```bash
   cd devstack
   npm install
   ```

2. **Bootstrap CDK (first time only):**
   ```bash
   npm run bootstrap
   ```

3. **Deploy the stack:**
   ```bash
   # Deploy with default Holesky RPC
   npm run deploy
   
   # Or with custom fork URL:
   npm run deploy -- --parameters ForkUrl=https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY
   ```

4. **Configure hgctl:**
   ```bash
   # Get executor endpoint from stack outputs
   EXECUTOR_ENDPOINT=$(aws cloudformation describe-stacks \
     --stack-name DevstackStack \
     --query 'Stacks[0].Outputs[?OutputKey==`ExecutorEndpoint`].OutputValue' \
     --output text)

   # Configure hgctl context
   hgctl context set aws --executor-address=$EXECUTOR_ENDPOINT
   ```

5. **Manage your AVS:**
   ```bash
   # List performers
   hgctl get performers

   # Deploy new version
   hgctl deploy artifact --digest=sha256:newversion

   # Check releases
   hgctl get releases
   ```

## Stack Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| ForkUrl | Ethereum fork URL | Holesky RPC endpoint | No |
| InstanceType | EC2 instance type | t3.large | No |

## Architecture

### Infrastructure Components

- **EC2 Instance**: Runs Amazon Linux 2023 with Docker pre-installed
- **Security Groups**: Configured for:
  - Executor gRPC (9090)
  - Aggregator gRPC (8081)
  - Ethereum RPC (8545)
  - SSH (22)
- **IAM Role**: Permissions for ECR access and Systems Manager
- **EBS Storage**: 50 GB GP3 volume for blockchain data and containers

### Software Stack

The EC2 instance automatically installs and configures:

- Docker (required by DevKit for container orchestration)
- DevKit CLI v0.0.8 (manages the entire AVS development environment)
- systemd service for persistent devnet operation

DevKit then handles:
- Creating and building the AVS project
- Running Anvil (local Ethereum node)
- Starting aggregator and executor services
- Bootstrapping the protocol with necessary configurations

## Usage Examples

### Basic Deployment

```bash
# Uses default Holesky RPC endpoint
npm run deploy
```

### Custom Configuration

```bash
# Use mainnet instead of Holesky
npm run deploy -- \
  --parameters InstanceType=t3.xlarge \
  --parameters ForkUrl=https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY
```

### Stack Outputs

After deployment, the stack provides these outputs:

- **ExecutorEndpoint**: Executor gRPC endpoint (port 9090) for hgctl connectivity
- **AggregatorEndpoint**: Aggregator gRPC endpoint (port 8081)
- **DevnetRpcUrl**: Ethereum RPC endpoint (port 8545) for blockchain interactions
- **SSHCommand**: SSH command for direct instance access
- **InstanceId**: EC2 instance ID for AWS operations

## Management Commands

### Deployment
```bash
npm run deploy          # Deploy the stack
npm run diff           # Show deployment changes
npm run synth          # Generate CloudFormation template
```

### Cleanup
```bash
npm run destroy        # Remove all resources
```

### Debugging
```bash
# SSH into the instance
ssh ec2-user@<instance-ip>

# Check devnet service status
sudo systemctl status devnet

# View logs
sudo journalctl -u devnet -f
tail -f /var/log/user-data.log

# Check Docker containers
docker ps

# Use the convenience script
devnet-status.sh
```

## Cost Optimization

- **Instance Type**: t3.large (~$0.08/hour) is sufficient for most development
- **Auto-shutdown**: Consider adding Lambda functions for scheduled shutdown
- **Spot Instances**: Can reduce costs by up to 90% for non-critical environments

## Security Considerations

- Security groups are configured for open access (0.0.0.0/0) by default
- For production use, restrict SSH and service access to specific IP ranges
- Consider using AWS Systems Manager Session Manager instead of SSH
- Use AWS Secrets Manager for sensitive configuration values

## Troubleshooting

### Common Issues

1. **Stack deployment fails**
   - Ensure AWS credentials are configured
   - Check AWS service quotas (especially for EC2)
   - Verify CDK is bootstrapped in your region

2. **Cannot connect to executor**
   - Check security group allows traffic from your IP
   - Verify the instance is running
   - Check devnet service status via SSH

3. **AVS deployment fails**
   - Verify container registry credentials
   - Check Docker daemon is running
   - Review deployment logs in `/var/log/user-data.log`

### Getting Help

- Check CloudFormation events for deployment errors
- SSH into the instance and check system logs
- Review the devnet service logs with `journalctl`

## Contributing

To contribute to DevStack:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `npm test`
5. Submit a pull request

## License

[Same as parent monorepo]
