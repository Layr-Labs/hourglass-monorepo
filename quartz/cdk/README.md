# Quartz Infrastructure

This CDK stack provides infrastructure for running AVS performers using Quartz, enabling serverless execution of AVS workloads with automatic scaling and pay-per-use pricing.

## Architecture Overview

The Quartz infrastructure consists of:

1. **Release Monitor** - Monitors the ReleaseManager contract for new releases
2. **Lambda Factory** - Step Functions state machine that creates/updates Lambda functions
3. **Performer Registry** - DynamoDB table and API for tracking AVS to Lambda mappings
4. **Lambda Adapter** - Go container that routes Executor requests to Lambda functions

### Flow Diagram

```
ReleaseManager Contract
        |
        v
Release Monitor (Lambda)
        |
        v
Step Functions State Machine
        |
        v
Lambda Factory (Creates/Updates)
        |
        v
AVS Lambda Functions
        ^
        |
Lambda Adapter Container <-- Executor
```

## Prerequisites

- AWS Account with appropriate permissions
- AWS CDK CLI installed (`npm install -g aws-cdk`)
- Node.js 18+ and npm
- Docker for building containers
- Access to an Ethereum RPC endpoint

## Deployment

### 1. Install Dependencies

```bash
cd quartz/cdk
npm install
```

### 2. Configure Environment

Set the following environment variables or CDK context:

```bash
export CDK_DEFAULT_ACCOUNT=123456789012
export CDK_DEFAULT_REGION=us-east-1
```

### 3. Deploy the Infrastructure

```bash
# Bootstrap CDK (first time only)
cdk bootstrap

# Deploy with configuration
cdk deploy \
  -c releaseManagerAddress=0xYOUR_RELEASE_MANAGER_ADDRESS \
  -c rpcEndpoint=https://your-rpc-endpoint.com \
  -c eksNodeRoleArn=arn:aws:iam::123456789012:role/eks-node-role  # Optional
```

### 4. Build and Push Lambda Adapter

```bash
cd ../adapter
make docker-push
```

### 5. Deploy Example Lambda (Optional)

```bash
cd ../example
docker build -t example-lambda .
docker tag example-lambda:latest $ECR_REPOSITORY_URI:example
docker push $ECR_REPOSITORY_URI:example
```

## Configuration

### Environment Variables for Lambda Adapter

- `AVS_ADDRESS` - The AVS contract address
- `OPERATOR_SET_ID` - The operator set ID
- `AWS_REGION` - AWS region (default: us-east-1)
- `LAMBDA_API_ENDPOINT` - Optional API Gateway endpoint
- `PORT` - HTTP server port (default: 8080)

### Lambda Function Naming Convention

Lambda functions are named deterministically:
```
avs-{address}-opset-{id}-performer
```

For example:
```
avs-0x1234567890123456789012345678901234567890-opset-1-performer
```

## API Endpoints

### Performer Registry API

The registry API provides endpoints for managing performer registrations:

- `GET /performers` - List all registered performers
- `GET /performers/{avs}/{opsetId}` - Get specific performer details
- `PUT /performers/{avs}/{opsetId}` - Register/update performer
- `DELETE /performers/{avs}/{opsetId}` - Deregister performer

Example:
```bash
# List all performers
curl https://your-api-id.execute-api.region.amazonaws.com/prod/performers

# Get specific performer
curl https://your-api-id.execute-api.region.amazonaws.com/prod/performers/0x123.../1
```

## Monitoring

### CloudWatch Logs

All components write logs to CloudWatch:

- `/aws/lambda/QuartzStack-ReleaseIndexer*` - Release monitor logs
- `/aws/lambda/QuartzStack-ProcessRelease*` - Lambda factory logs
- `/aws/lambda/avs-*-performer` - Individual AVS Lambda logs
- `/aws/states/quartz-factory` - Step Functions execution logs

### Metrics

Key metrics to monitor:

- Release monitor invocation count and errors
- Step Functions execution success/failure rate
- Lambda function invocations and duration
- API Gateway request count and latency

## Cost Considerations

The Lambda performer infrastructure is designed to be cost-effective:

1. **Lambda Functions** - Pay only for actual execution time
2. **Step Functions** - Charged per state transition
3. **DynamoDB** - Pay-per-request pricing
4. **API Gateway** - Charged per request

Estimated monthly costs for a typical AVS:
- Light usage (1000 tasks/month): ~$5-10
- Medium usage (10,000 tasks/month): ~$20-50
- Heavy usage (100,000 tasks/month): ~$100-300

## Security

### IAM Permissions

The infrastructure uses least-privilege IAM policies:

- Release Monitor: Read RPC, start Step Functions
- Lambda Factory: Create/update specific Lambda functions
- Lambda Functions: Basic execution only
- Lambda Adapter: Invoke Lambda functions

### Network Security

- Lambda functions run in isolated environments
- API Gateway provides request validation
- All inter-service communication uses IAM authentication

## Troubleshooting

### Common Issues

1. **Lambda function not found**
   - Check that the Release Monitor has processed the release
   - Verify the Step Functions execution succeeded
   - Check CloudWatch logs for errors

2. **Permission denied errors**
   - Ensure the EKS node role has ECR pull permissions
   - Verify Lambda execution roles are created

3. **Timeout errors**
   - Increase Lambda function timeout (max 15 minutes)
   - Check if the task requires more memory

### Debug Commands

```bash
# Check Release Monitor logs
aws logs tail /aws/lambda/QuartzStack-ReleaseIndexer --follow

# List Step Functions executions
aws stepfunctions list-executions --state-machine-arn arn:aws:states:...

# Get Lambda function details
aws lambda get-function --function-name avs-0x123...-opset-1-performer
```

## Development

### Local Testing

1. Run the Lambda adapter locally:
```bash
cd lambda-adapter
make run
```

2. Test with curl:
```bash
curl -X POST http://localhost:8080/perform \
  -H "Content-Type: application/json" \
  -d '{
    "avsAddress": "0x123...",
    "taskId": "task-1",
    "payload": {"operation": "compute", "params": {"a": 5, "b": 3}}
  }'
```

### Adding Custom Logic

To implement your own AVS logic:

1. Create a new Lambda function based on the example
2. Build and push to ECR
3. Update the Release Manager to point to your image
4. The infrastructure will automatically deploy it

## License

This infrastructure is part of the Hourglass framework and follows the same license terms.