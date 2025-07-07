import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as events from 'aws-cdk-lib/aws-events';
import * as targets from 'aws-cdk-lib/aws-events-targets';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as sfn from 'aws-cdk-lib/aws-stepfunctions';
import { Construct } from 'constructs';

interface ReleaseMonitorProps {
  releaseManagerAddress: string;
  rpcEndpoint: string;
  stateMachine: sfn.StateMachine;
}

export class ReleaseMonitor extends Construct {
  public readonly indexerFunction: lambda.Function;

  constructor(scope: Construct, id: string, props: ReleaseMonitorProps) {
    super(scope, id);

    // Create Lambda function to monitor ReleaseManager events
    this.indexerFunction = new lambda.Function(this, 'ReleaseIndexer', {
      runtime: lambda.Runtime.NODEJS_18_X,
      handler: 'index.handler',
      code: lambda.Code.fromInline(this.getIndexerCode()),
      environment: {
        RELEASE_MANAGER_ADDRESS: props.releaseManagerAddress,
        RPC_ENDPOINT: props.rpcEndpoint,
        STATE_MACHINE_ARN: props.stateMachine.stateMachineArn,
      },
      timeout: cdk.Duration.minutes(5),
      memorySize: 512,
      logRetention: logs.RetentionDays.ONE_WEEK,
    });

    // Grant permissions to start Step Functions execution
    props.stateMachine.grantStartExecution(this.indexerFunction);

    // Create EventBridge rule to trigger periodically
    const rule = new events.Rule(this, 'IndexerSchedule', {
      schedule: events.Schedule.rate(cdk.Duration.minutes(5)),
      description: 'Trigger release indexer every 5 minutes',
    });

    rule.addTarget(new targets.LambdaFunction(this.indexerFunction));

    // Add permissions for RPC calls
    this.indexerFunction.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ['ssm:GetParameter', 'ssm:PutParameter'],
        resources: ['arn:aws:ssm:*:*:parameter/quartz/*'],
      })
    );
  }

  private getIndexerCode(): string {
    return `
const { ethers } = require('ethers');
const AWS = require('aws-sdk');

const sfn = new AWS.StepFunctions();
const ssm = new AWS.SSM();

const RELEASE_MANAGER_ABI = [
  "event ReleasePublished((address,uint32) indexed operatorSet, uint256 indexed releaseId, ((bytes32,string)[],uint32) release)",
  "function getRelease((address,uint32) operatorSet, uint256 releaseId) view returns (((bytes32,string)[],uint32))"
];

exports.handler = async (event) => {
  console.log('Checking for new releases...');
  
  const provider = new ethers.JsonRpcProvider(process.env.RPC_ENDPOINT);
  const releaseManager = new ethers.Contract(
    process.env.RELEASE_MANAGER_ADDRESS,
    RELEASE_MANAGER_ABI,
    provider
  );

  // Get last processed block
  let lastBlock = 0;
  try {
    const param = await ssm.getParameter({
      Name: '/quartz/last-processed-block',
    }).promise();
    lastBlock = parseInt(param.Parameter.Value);
  } catch (e) {
    console.log('No last block found, starting from 0');
  }

  // Get current block
  const currentBlock = await provider.getBlockNumber();
  
  // Query events
  const filter = releaseManager.filters.ReleasePublished();
  const events = await releaseManager.queryFilter(filter, lastBlock + 1, currentBlock);
  
  console.log(\`Found \${events.length} new releases\`);

  // Process each event
  for (const event of events) {
    const { operatorSet, releaseId, release } = event.args;
    
    console.log(\`Processing release \${releaseId} for AVS \${operatorSet[0]}, OpSet \${operatorSet[1]}\`);
    
    // Start Step Functions execution
    const params = {
      stateMachineArn: process.env.STATE_MACHINE_ARN,
      name: \`release-\${operatorSet[0]}-\${operatorSet[1]}-\${releaseId}-\${Date.now()}\`,
      input: JSON.stringify({
        operatorSet: {
          avs: operatorSet[0],
          id: operatorSet[1].toString(),
        },
        releaseId: releaseId.toString(),
        release: {
          artifacts: release.artifacts.map(a => ({
            digest: a.digest,
            registryUrl: a.registryUrl,
          })),
          upgradeByTime: release.upgradeByTime.toString(),
        },
        blockNumber: event.blockNumber,
        transactionHash: event.transactionHash,
      }),
    };
    
    await sfn.startExecution(params).promise();
  }

  // Update last processed block
  await ssm.putParameter({
    Name: '/quartz/last-processed-block',
    Value: currentBlock.toString(),
    Type: 'String',
    Overwrite: true,
  }).promise();

  return {
    statusCode: 200,
    body: JSON.stringify({
      processed: events.length,
      lastBlock,
      currentBlock,
    }),
  };
};
`;
  }
}