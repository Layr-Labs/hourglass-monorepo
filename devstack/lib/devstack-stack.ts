import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';
import { generateUserDataScript } from './utils/user-data-generator';

export interface DevstackStackProps extends cdk.StackProps {
  // None of these are required as we'll use parameters
}

export class DevstackStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: DevstackStackProps) {
    super(scope, id, props);

    // Parameters - simplified for devkit approach
    const forkUrl = new cdk.CfnParameter(this, 'ForkUrl', {
      type: 'String',
      default: 'https://fragrant-intensive-resonance.quiknode.pro/d13303176efa5210c6b9f264823045bf2400fd16/',
      description: 'Ethereum fork URL (defaults to Holesky)',
    });

    const instanceType = new cdk.CfnParameter(this, 'InstanceType', {
      type: 'String',
      default: 't3.large',
      description: 'EC2 instance type',
      allowedValues: ['t3.large', 't3.xlarge', 't3.2xlarge', 'm5.large', 'm5.xlarge'],
    });

    // VPC - create a new one with minimal subnets to save costs
    const vpc = new ec2.Vpc(this, 'DevstackVPC', {
      maxAzs: 2,
      natGateways: 0, // Save costs - no NAT gateway needed for public instances
      subnetConfiguration: [
        {
          name: 'Public',
          subnetType: ec2.SubnetType.PUBLIC,
          cidrMask: 24,
        },
      ],
    });

    // Security Group
    const securityGroup = new ec2.SecurityGroup(this, 'DevstackSecurityGroup', {
      vpc,
      description: 'Security group for DevStack EC2 instance',
      allowAllOutbound: true,
    });

    // Allow inbound traffic for metrics
    securityGroup.addIngressRule(
        ec2.Peer.anyIpv4(),
        ec2.Port.tcp(9000),
        'Allow Aggregator metrics/monitoring access'
    );

    // Allow inbound traffic for devnet services
    // Executor gRPC port (default 9090)
    securityGroup.addIngressRule(
      ec2.Peer.anyIpv4(),
      ec2.Port.tcp(9090),
      'Allow executor gRPC access'
    );

    // Ethereum RPC port (default 8545)
    securityGroup.addIngressRule(
      ec2.Peer.anyIpv4(),
      ec2.Port.tcp(8545),
      'Allow Ethereum RPC access'
    );
    
    // Aggregator gRPC port (default 8081)
    securityGroup.addIngressRule(
      ec2.Peer.anyIpv4(),
      ec2.Port.tcp(8081),
      'Allow aggregator gRPC access'
    );

    // IAM Role for EC2 instance
    const role = new iam.Role(this, 'DevstackInstanceRole', {
      assumedBy: new iam.ServicePrincipal('ec2.amazonaws.com'),
      managedPolicies: [
        iam.ManagedPolicy.fromAwsManagedPolicyName('AmazonSSMManagedInstanceCore'),
      ],
    });

    // Add permissions for ECR if using private registries
    role.addToPolicy(new iam.PolicyStatement({
      actions: [
        'ecr:GetAuthorizationToken',
        'ecr:BatchCheckLayerAvailability',
        'ecr:GetDownloadUrlForLayer',
        'ecr:BatchGetImage',
      ],
      resources: ['*'],
    }));

    // EC2 Instance
    const instance = new ec2.Instance(this, 'DevstackInstance', {
      instanceType: new ec2.InstanceType(instanceType.valueAsString),
      machineImage: ec2.MachineImage.latestAmazonLinux2023(),
      vpc,
      vpcSubnets: {
        subnetType: ec2.SubnetType.PUBLIC,
      },
      securityGroup,
      role,
      blockDevices: [{
        deviceName: '/dev/xvda',
        volume: ec2.BlockDeviceVolume.ebs(50, {
          volumeType: ec2.EbsDeviceVolumeType.GP3,
          deleteOnTermination: true,
        }),
      }],
    });

    // User data script - simplified for devkit
    const userDataScript = generateUserDataScript({
      forkUrl: forkUrl.valueAsString,
    });

    instance.addUserData(userDataScript);
    
    // Enable EC2 Instance Connect for temporary SSH access
    instance.addUserData(
      '# Install EC2 Instance Connect',
      'sudo yum install -y ec2-instance-connect'
    );
    
    // Add a tag with user data hash to force instance replacement when script changes
    // This ensures the new user-data script runs on a fresh instance
    const crypto = require('crypto');
    const userDataHash = crypto.createHash('sha256').update(userDataScript).digest('hex').substring(0, 8);
    cdk.Tags.of(instance).add('UserDataVersion', userDataHash);

    // Outputs
    new cdk.CfnOutput(this, 'ExecutorEndpoint', {
      value: `${instance.instancePublicIp}:9090`,
      description: 'Executor gRPC endpoint for hgctl',
    });

    new cdk.CfnOutput(this, 'AggregatorEndpoint', {
      value: `${instance.instancePublicIp}:8081`,
      description: 'Aggregator gRPC endpoint',
    });

    new cdk.CfnOutput(this, 'DevnetRpcUrl', {
      value: `http://${instance.instancePublicIp}:8545`,
      description: 'Ethereum RPC endpoint',
    });

    new cdk.CfnOutput(this, 'SessionManagerCommand', {
      value: `aws ssm start-session --target ${instance.instanceId}`,
      description: 'Connect to instance via Session Manager',
    });

    new cdk.CfnOutput(this, 'InstanceId', {
      value: instance.instanceId,
      description: 'EC2 Instance ID',
    });
  }
}