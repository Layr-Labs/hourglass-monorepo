import * as cdk from 'aws-cdk-lib';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as logs from 'aws-cdk-lib/aws-logs';
import { Construct } from 'constructs';

interface PerformerRegistryProps {
  // Optional properties for future extensibility
}

export class PerformerRegistry extends Construct {
  public readonly table: dynamodb.Table;
  public readonly api: apigateway.RestApi;
  
  constructor(scope: Construct, id: string, props?: PerformerRegistryProps) {
    super(scope, id);

    // Create DynamoDB table for AVS to Lambda mapping
    this.table = new dynamodb.Table(this, 'RegistryTable', {
      tableName: 'quartz-performer-registry',
      partitionKey: {
        name: 'avsAddress',
        type: dynamodb.AttributeType.STRING,
      },
      sortKey: {
        name: 'operatorSetId',
        type: dynamodb.AttributeType.NUMBER,
      },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    // Add GSI for looking up by Lambda function name
    this.table.addGlobalSecondaryIndex({
      indexName: 'functionNameIndex',
      partitionKey: {
        name: 'functionName',
        type: dynamodb.AttributeType.STRING,
      },
    });

    // Create Lambda function for registry operations
    const registryFunction = new lambda.Function(this, 'RegistryFunction', {
      runtime: lambda.Runtime.NODEJS_18_X,
      handler: 'index.handler',
      code: lambda.Code.fromInline(this.getRegistryCode()),
      environment: {
        TABLE_NAME: this.table.tableName,
      },
      timeout: cdk.Duration.seconds(30),
      memorySize: 256,
      logRetention: logs.RetentionDays.ONE_WEEK,
    });

    // Grant permissions
    this.table.grantReadWriteData(registryFunction);

    // Add Lambda permissions
    registryFunction.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ['lambda:GetFunction'],
        resources: ['arn:aws:lambda:*:*:function:avs-*-opset-*-performer'],
      })
    );

    registryFunction.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ['apigateway:GET'],
        resources: ['arn:aws:apigateway:*::/restapis/*'],
      })
    );

    // Create API Gateway
    this.api = new apigateway.RestApi(this, 'RegistryApi', {
      restApiName: 'quartz-performer-registry',
      description: 'API for Quartz performer registry operations',
      deployOptions: {
        stageName: 'prod',
        loggingLevel: apigateway.MethodLoggingLevel.INFO,
        dataTraceEnabled: true,
      },
    });

    // Add Lambda integration
    const lambdaIntegration = new apigateway.LambdaIntegration(registryFunction);

    // Create resources and methods
    const performersResource = this.api.root.addResource('performers');
    
    // GET /performers - List all performers
    performersResource.addMethod('GET', lambdaIntegration);
    
    // GET /performers/{avs}/{opsetId} - Get specific performer
    const avsResource = performersResource.addResource('{avs}');
    const opsetResource = avsResource.addResource('{opsetId}');
    opsetResource.addMethod('GET', lambdaIntegration);
    
    // PUT /performers/{avs}/{opsetId} - Register/update performer
    opsetResource.addMethod('PUT', lambdaIntegration);
    
    // DELETE /performers/{avs}/{opsetId} - Deregister performer
    opsetResource.addMethod('DELETE', lambdaIntegration);
  }

  private getRegistryCode(): string {
    return `
const AWS = require('aws-sdk');
const dynamodb = new AWS.DynamoDB.DocumentClient();
const lambda = new AWS.Lambda();
const apigateway = new AWS.APIGateway();

exports.handler = async (event) => {
  console.log('Registry request:', JSON.stringify(event, null, 2));
  
  const tableName = process.env.TABLE_NAME;
  const { httpMethod, pathParameters, body } = event;
  
  try {
    let response;
    
    switch (httpMethod) {
      case 'GET':
        if (pathParameters && pathParameters.avs && pathParameters.opsetId) {
          response = await getPerformer(tableName, pathParameters.avs, parseInt(pathParameters.opsetId));
        } else {
          response = await listPerformers(tableName);
        }
        break;
        
      case 'PUT':
        response = await registerPerformer(tableName, pathParameters.avs, parseInt(pathParameters.opsetId), JSON.parse(body || '{}'));
        break;
        
      case 'DELETE':
        response = await deregisterPerformer(tableName, pathParameters.avs, parseInt(pathParameters.opsetId));
        break;
        
      default:
        throw new Error(\`Unsupported method: \${httpMethod}\`);
    }
    
    return {
      statusCode: 200,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(response),
    };
  } catch (error) {
    console.error('Error:', error);
    return {
      statusCode: error.statusCode || 500,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        error: error.message,
      }),
    };
  }
};

async function getPerformer(tableName, avsAddress, operatorSetId) {
  const params = {
    TableName: tableName,
    Key: {
      avsAddress: avsAddress.toLowerCase(),
      operatorSetId: operatorSetId,
    },
  };
  
  const result = await dynamodb.get(params).promise();
  
  if (!result.Item) {
    const error = new Error('Performer not found');
    error.statusCode = 404;
    throw error;
  }
  
  // Enrich with current Lambda status
  try {
    const functionStatus = await lambda.getFunction({
      FunctionName: result.Item.functionName,
    }).promise();
    
    result.Item.functionStatus = {
      state: functionStatus.Configuration.State,
      lastModified: functionStatus.Configuration.LastModified,
      codeSize: functionStatus.Configuration.CodeSize,
    };
  } catch (e) {
    result.Item.functionStatus = { error: 'Function not found' };
  }
  
  return result.Item;
}

async function listPerformers(tableName) {
  const params = {
    TableName: tableName,
  };
  
  const result = await dynamodb.scan(params).promise();
  return {
    performers: result.Items,
    count: result.Count,
  };
}

async function registerPerformer(tableName, avsAddress, operatorSetId, data) {
  const functionName = \`avs-\${avsAddress.toLowerCase()}-opset-\${operatorSetId}-performer\`;
  
  // Verify Lambda function exists
  try {
    await lambda.getFunction({ FunctionName: functionName }).promise();
  } catch (e) {
    const error = new Error(\`Lambda function \${functionName} not found\`);
    error.statusCode = 400;
    throw error;
  }
  
  const item = {
    avsAddress: avsAddress.toLowerCase(),
    operatorSetId: operatorSetId,
    functionName: functionName,
    registeredAt: new Date().toISOString(),
    ...data,
  };
  
  const params = {
    TableName: tableName,
    Item: item,
  };
  
  await dynamodb.put(params).promise();
  
  return {
    message: 'Performer registered successfully',
    performer: item,
  };
}

async function deregisterPerformer(tableName, avsAddress, operatorSetId) {
  const params = {
    TableName: tableName,
    Key: {
      avsAddress: avsAddress.toLowerCase(),
      operatorSetId: operatorSetId,
    },
  };
  
  // Check if exists
  const existing = await dynamodb.get(params).promise();
  if (!existing.Item) {
    const error = new Error('Performer not found');
    error.statusCode = 404;
    throw error;
  }
  
  await dynamodb.delete(params).promise();
  
  return {
    message: 'Performer deregistered successfully',
    deletedPerformer: existing.Item,
  };
}
`;
  }
}