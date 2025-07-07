import * as cdk from 'aws-cdk-lib';
import * as sfn from 'aws-cdk-lib/aws-stepfunctions';
import * as tasks from 'aws-cdk-lib/aws-stepfunctions-tasks';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as ecr from 'aws-cdk-lib/aws-ecr';
import { Construct } from 'constructs';

interface LambdaFactoryProps {
  ecrRepository: ecr.Repository;
}

export class LambdaFactory extends Construct {
  public readonly stateMachine: sfn.StateMachine;
  
  constructor(scope: Construct, id: string, props: LambdaFactoryProps) {
    super(scope, id);

    // Create Lambda function for processing releases
    const processReleaseFunction = new lambda.Function(this, 'ProcessRelease', {
      runtime: lambda.Runtime.NODEJS_18_X,
      handler: 'index.handler',
      code: lambda.Code.fromInline(this.getProcessReleaseCode()),
      timeout: cdk.Duration.minutes(10),
      memorySize: 1024,
      environment: {
        ECR_REPOSITORY_URI: props.ecrRepository.repositoryUri,
      },
      logRetention: logs.RetentionDays.ONE_WEEK,
    });

    // Grant permissions to manage Lambda functions and API Gateway
    processReleaseFunction.addToRolePolicy(
      new iam.PolicyStatement({
        actions: [
          'lambda:CreateFunction',
          'lambda:UpdateFunctionCode',
          'lambda:UpdateFunctionConfiguration',
          'lambda:AddPermission',
          'lambda:GetFunction',
          'lambda:TagResource',
        ],
        resources: ['arn:aws:lambda:*:*:function:avs-*-opset-*-performer'],
      })
    );

    processReleaseFunction.addToRolePolicy(
      new iam.PolicyStatement({
        actions: [
          'apigateway:POST',
          'apigateway:GET',
          'apigateway:PUT',
          'apigateway:DELETE',
        ],
        resources: ['arn:aws:apigateway:*::/restapis/*'],
      })
    );

    processReleaseFunction.addToRolePolicy(
      new iam.PolicyStatement({
        actions: [
          'iam:CreateRole',
          'iam:AttachRolePolicy',
          'iam:PassRole',
          'iam:GetRole',
        ],
        resources: ['arn:aws:iam::*:role/avs-*-performer-role'],
      })
    );

    // Grant ECR permissions
    props.ecrRepository.grantPull(processReleaseFunction);
    props.ecrRepository.grantPullPush(processReleaseFunction);

    // Create state machine
    const processReleaseTask = new tasks.LambdaInvoke(this, 'ProcessReleaseTask', {
      lambdaFunction: processReleaseFunction,
      outputPath: '$.Payload',
      retryOnServiceExceptions: true,
    });

    const success = new sfn.Succeed(this, 'Success', {
      comment: 'Quartz performer deployed successfully',
    });

    const failed = new sfn.Fail(this, 'Failed', {
      comment: 'Failed to deploy Quartz performer',
    });

    const definition = processReleaseTask
      .addCatch(failed, {
        errors: ['States.ALL'],
        resultPath: '$.error',
      })
      .next(success);

    this.stateMachine = new sfn.StateMachine(this, 'LambdaFactoryStateMachine', {
      definition,
      stateMachineName: 'quartz-factory',
      timeout: cdk.Duration.minutes(30),
      logs: {
        destination: new logs.LogGroup(this, 'StateMachineLogs', {
          retention: logs.RetentionDays.ONE_WEEK,
        }),
        level: sfn.LogLevel.ALL,
      },
    });
  }

  private getProcessReleaseCode(): string {
    return `
const AWS = require('aws-sdk');
const crypto = require('crypto');

const lambda = new AWS.Lambda();
const apigateway = new AWS.APIGateway();
const iam = new AWS.IAM();

exports.handler = async (event) => {
  console.log('Processing release:', JSON.stringify(event, null, 2));
  
  const { operatorSet, releaseId, release } = event;
  const { avs, id: opSetId } = operatorSet;
  
  // Deterministic function name based on AVS address and operator set ID
  const functionName = \`avs-\${avs.toLowerCase()}-opset-\${opSetId}-performer\`;
  
  try {
    // Create or update Lambda function
    const lambdaRoleArn = await ensureLambdaRole(functionName);
    
    // For now, use a placeholder image - in production this would pull from release artifacts
    const imageUri = \`\${process.env.ECR_REPOSITORY_URI}:latest\`;
    
    const lambdaConfig = {
      FunctionName: functionName,
      Role: lambdaRoleArn,
      Code: {
        ImageUri: imageUri,
      },
      PackageType: 'Image',
      Timeout: 300,
      MemorySize: 3008,
      Environment: {
        Variables: {
          AVS_ADDRESS: avs,
          OPERATOR_SET_ID: opSetId.toString(),
          RELEASE_ID: releaseId.toString(),
        },
      },
      Tags: {
        'avs-address': avs,
        'operator-set-id': opSetId.toString(),
        'release-id': releaseId.toString(),
        'managed-by': 'quartz-factory',
      },
    };
    
    let lambdaArn;
    try {
      // Try to update existing function
      const existingFunction = await lambda.getFunction({ FunctionName: functionName }).promise();
      lambdaArn = existingFunction.Configuration.FunctionArn;
      
      // Update function code
      await lambda.updateFunctionCode({
        FunctionName: functionName,
        ImageUri: imageUri,
      }).promise();
      
      // Update function configuration
      await lambda.updateFunctionConfiguration({
        FunctionName: functionName,
        Timeout: lambdaConfig.Timeout,
        MemorySize: lambdaConfig.MemorySize,
        Environment: lambdaConfig.Environment,
      }).promise();
      
      console.log('Updated existing Lambda function:', functionName);
    } catch (e) {
      if (e.code === 'ResourceNotFoundException') {
        // Create new function
        const createResult = await lambda.createFunction(lambdaConfig).promise();
        lambdaArn = createResult.FunctionArn;
        console.log('Created new Lambda function:', functionName);
      } else {
        throw e;
      }
    }
    
    // Create or update API Gateway
    const apiId = await ensureApiGateway(functionName, lambdaArn);
    
    return {
      success: true,
      functionName,
      functionArn: lambdaArn,
      apiId,
      message: \`Successfully deployed Quartz performer for AVS \${avs}, OpSet \${opSetId}\`,
    };
  } catch (error) {
    console.error('Error processing release:', error);
    throw error;
  }
};

async function ensureLambdaRole(functionName) {
  const roleName = \`\${functionName}-role\`;
  
  try {
    const role = await iam.getRole({ RoleName: roleName }).promise();
    return role.Role.Arn;
  } catch (e) {
    if (e.code === 'NoSuchEntity') {
      // Create role
      const assumeRolePolicyDocument = {
        Version: '2012-10-17',
        Statement: [{
          Effect: 'Allow',
          Principal: {
            Service: 'lambda.amazonaws.com',
          },
          Action: 'sts:AssumeRole',
        }],
      };
      
      const createRoleResult = await iam.createRole({
        RoleName: roleName,
        AssumeRolePolicyDocument: JSON.stringify(assumeRolePolicyDocument),
        Description: \`Execution role for \${functionName}\`,
        Tags: [{
          Key: 'managed-by',
          Value: 'quartz-factory',
        }],
      }).promise();
      
      // Attach basic execution policy
      await iam.attachRolePolicy({
        RoleName: roleName,
        PolicyArn: 'arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole',
      }).promise();
      
      // Wait for role to be available
      await new Promise(resolve => setTimeout(resolve, 10000));
      
      return createRoleResult.Role.Arn;
    }
    throw e;
  }
}

async function ensureApiGateway(functionName, lambdaArn) {
  const apiName = \`\${functionName}-api\`;
  
  try {
    // Look for existing API
    const apis = await apigateway.getRestApis({ limit: 500 }).promise();
    const existingApi = apis.items.find(api => api.name === apiName);
    
    let apiId;
    if (existingApi) {
      apiId = existingApi.id;
      console.log('Using existing API Gateway:', apiId);
    } else {
      // Create new API
      const createApiResult = await apigateway.createRestApi({
        name: apiName,
        description: \`API Gateway for \${functionName}\`,
        endpointConfiguration: {
          types: ['REGIONAL'],
        },
      }).promise();
      apiId = createApiResult.id;
      console.log('Created new API Gateway:', apiId);
      
      // Get root resource
      const resources = await apigateway.getResources({ restApiId: apiId }).promise();
      const rootId = resources.items[0].id;
      
      // Create proxy resource
      const proxyResource = await apigateway.createResource({
        restApiId: apiId,
        parentId: rootId,
        pathPart: '{proxy+}',
      }).promise();
      
      // Create ANY method
      await apigateway.putMethod({
        restApiId: apiId,
        resourceId: proxyResource.id,
        httpMethod: 'ANY',
        authorizationType: 'NONE',
        requestParameters: {
          'method.request.path.proxy': true,
        },
      }).promise();
      
      // Create integration
      await apigateway.putIntegration({
        restApiId: apiId,
        resourceId: proxyResource.id,
        httpMethod: 'ANY',
        type: 'AWS_PROXY',
        integrationHttpMethod: 'POST',
        uri: \`arn:aws:apigateway:\${process.env.AWS_REGION}:lambda:path/2015-03-31/functions/\${lambdaArn}/invocations\`,
      }).promise();
      
      // Add Lambda permission
      await lambda.addPermission({
        FunctionName: functionName,
        StatementId: \`apigateway-\${apiId}\`,
        Action: 'lambda:InvokeFunction',
        Principal: 'apigateway.amazonaws.com',
        SourceArn: \`arn:aws:execute-api:\${process.env.AWS_REGION}:*:\${apiId}/*/*\`,
      }).promise();
      
      // Deploy API
      await apigateway.createDeployment({
        restApiId: apiId,
        stageName: 'prod',
      }).promise();
    }
    
    return apiId;
  } catch (error) {
    console.error('Error managing API Gateway:', error);
    throw error;
  }
}
`;
  }
}