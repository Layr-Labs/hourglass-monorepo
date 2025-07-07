# Example Lambda Function for AVS Performer

This directory contains an example Lambda function that demonstrates how an AVS can implement custom logic to be executed by the Quartz infrastructure.

## Structure

- `index.js` - The Lambda handler function with example processing logic
- `Dockerfile` - Container definition for the Lambda function

## Example Operations

The example Lambda supports three types of operations:

### 1. Compute Operation
```json
{
  "operation": "compute",
  "params": {
    "a": 10,
    "b": 20
  }
}
```

### 2. Transform Operation
```json
{
  "operation": "transform",
  "params": {
    "input": "hello world"
  }
}
```

### 3. Async Operation
```json
{
  "operation": "async",
  "params": {}
}
```

## Building and Deploying

1. Build the Docker image:
```bash
docker build -t example-lambda .
```

2. Tag for ECR:
```bash
docker tag example-lambda:latest <account-id>.dkr.ecr.<region>.amazonaws.com/quartz-performer:example
```

3. Push to ECR:
```bash
aws ecr get-login-password --region <region> | docker login --username AWS --password-stdin <account-id>.dkr.ecr.<region>.amazonaws.com
docker push <account-id>.dkr.ecr.<region>.amazonaws.com/quartz-performer:example
```

## Testing Locally

You can test the Lambda function locally using the AWS SAM CLI:

```bash
# Install SAM CLI if not already installed
pip install aws-sam-cli

# Test with a sample event
echo '{"avsAddress":"0x123","taskId":"task-1","payload":{"operation":"compute","params":{"a":5,"b":3}}}' > event.json
sam local invoke -e event.json
```

## Customizing for Your AVS

To create your own AVS performer:

1. Replace the `processTask` function with your AVS-specific logic
2. Update the handler to process your specific task format
3. Add any required dependencies to the Dockerfile
4. Build and deploy your custom image