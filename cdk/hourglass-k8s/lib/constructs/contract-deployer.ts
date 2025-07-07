import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as cr from 'aws-cdk-lib/custom-resources';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as logs from 'aws-cdk-lib/aws-logs';
import { Construct } from 'constructs';

interface ContractDeployerProps {
  vpc: ec2.Vpc;
  l1Endpoint: string;
  l2Endpoint: string;
  accounts: any[];
  devnetConfig: any;
}

export class ContractDeployer extends Construct {
  public readonly mailboxL1Address: string;
  public readonly mailboxL2Address: string;
  public readonly avsTaskRegistrarAddress: string;
  public readonly taskHookL1Address: string;
  public readonly taskHookL2Address: string;

  constructor(scope: Construct, id: string, props: ContractDeployerProps) {
    super(scope, id);

    // Create Lambda function for contract deployment
    const deployerFunction = new lambda.Function(this, 'DeployerFunction', {
      runtime: lambda.Runtime.NODEJS_18_X,
      handler: 'index.handler',
      code: lambda.Code.fromInline(this.getDeployerCode()),
      vpc: props.vpc,
      vpcSubnets: {
        subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
      },
      timeout: cdk.Duration.minutes(15),
      memorySize: 2048,
      environment: {
        L1_ENDPOINT: props.l1Endpoint,
        L2_ENDPOINT: props.l2Endpoint,
        ACCOUNTS: JSON.stringify(props.accounts),
        DEVNET_CONFIG: JSON.stringify(props.devnetConfig),
      },
    });

    // Grant permissions
    deployerFunction.addToRolePolicy(
      new iam.PolicyStatement({
        actions: [
          'ssm:PutParameter',
          'ssm:GetParameter',
        ],
        resources: ['arn:aws:ssm:*:*:parameter/hourglass/*'],
      })
    );

    // Create custom resource
    const provider = new cr.Provider(this, 'DeployerProvider', {
      onEventHandler: deployerFunction,
      logRetention: logs.RetentionDays.ONE_WEEK,
    });

    const deployment = new cdk.CustomResource(this, 'ContractDeployment', {
      serviceToken: provider.serviceToken,
      properties: {
        timestamp: Date.now(), // Force update on each deployment
      },
    });

    // Extract addresses from deployment
    this.mailboxL1Address = deployment.getAttString('mailboxL1');
    this.mailboxL2Address = deployment.getAttString('mailboxL2');
    this.avsTaskRegistrarAddress = deployment.getAttString('avsTaskRegistrar');
    this.taskHookL1Address = deployment.getAttString('taskHookL1');
    this.taskHookL2Address = deployment.getAttString('taskHookL2');
  }

  private getDeployerCode(): string {
    return `
const { ethers } = require('ethers');
const AWS = require('aws-sdk');

const ssm = new AWS.SSM();

exports.handler = async (event) => {
  console.log('Event:', JSON.stringify(event, null, 2));
  
  if (event.RequestType === 'Delete') {
    return { PhysicalResourceId: event.PhysicalResourceId };
  }

  try {
    const accounts = JSON.parse(process.env.ACCOUNTS);
    const config = JSON.parse(process.env.DEVNET_CONFIG);
    
    // Connect to L1 and L2
    const l1Provider = new ethers.JsonRpcProvider(process.env.L1_ENDPOINT);
    const l2Provider = new ethers.JsonRpcProvider(process.env.L2_ENDPOINT);
    
    // Get deployer account
    const deployerAccount = accounts[0];
    const l1Signer = new ethers.Wallet(deployerAccount.private_key, l1Provider);
    const l2Signer = new ethers.Wallet(deployerAccount.private_key, l2Provider);
    
    // Fund accounts
    console.log('Funding accounts...');
    for (const account of accounts) {
      await fundAccount(l1Provider, l2Provider, account.address);
    }
    
    // Deploy contracts (simplified for this example)
    console.log('Deploying contracts...');
    
    // In a real implementation, you would:
    // 1. Deploy mailbox contracts to L1 and L2
    // 2. Deploy AVS contracts
    // 3. Setup operators
    // 4. Stake
    
    // For now, return mock addresses
    const addresses = {
      mailboxL1: '0x' + '1'.repeat(40),
      mailboxL2: '0x' + '2'.repeat(40),
      avsTaskRegistrar: '0x' + '3'.repeat(40),
      taskHookL1: '0x' + '4'.repeat(40),
      taskHookL2: '0x' + '5'.repeat(40),
    };
    
    // Store in Parameter Store
    for (const [key, value] of Object.entries(addresses)) {
      await ssm.putParameter({
        Name: \`/hourglass/contracts/\${key}\`,
        Value: value,
        Type: 'String',
        Overwrite: true,
      }).promise();
    }
    
    return {
      PhysicalResourceId: 'hourglass-contracts-' + Date.now(),
      Data: addresses,
    };
  } catch (error) {
    console.error('Error:', error);
    throw error;
  }
};

async function fundAccount(l1Provider, l2Provider, address) {
  const fundAmount = ethers.parseEther('10000');
  
  // Use Anvil RPC methods to set balance
  await l1Provider.send('anvil_setBalance', [address, fundAmount.toString(16)]);
  await l2Provider.send('anvil_setBalance', [address, fundAmount.toString(16)]);
  
  console.log(\`Funded \${address} with 10,000 ETH on both chains\`);
}
`;
  }
}