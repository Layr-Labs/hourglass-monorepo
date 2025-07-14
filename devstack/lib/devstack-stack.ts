import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as ssm from 'aws-cdk-lib/aws-ssm';
import { Construct } from 'constructs';
import { generateUserDataScript } from './utils/user-data-generator';
import * as fs from 'fs';
import * as path from 'path';

export interface DevstackStackProps extends cdk.StackProps {
  // None of these are required as we'll use parameters
}

export class DevstackStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: DevstackStackProps) {
    super(scope, id, props);

    // Parameters - simplified for devkit approach
    const forkUrl = new cdk.CfnParameter(this, 'ForkUrl', {
      type: 'String',
      default: 'https://practical-serene-mound.ethereum-sepolia.quiknode.pro/3aaa48bd95f3d6aed60e89a1a466ed1e2a440b61/',
      description: 'Ethereum fork URL (defaults to Sepolia)',
    });

    const instanceType = new cdk.CfnParameter(this, 'InstanceType', {
      type: 'String',
      default: 't3.large',
      description: 'EC2 instance type',
      allowedValues: ['t3.large', 't3.xlarge', 't3.2xlarge', 'm5.large', 'm5.xlarge'],
    });

    // Create SSM Parameter for systemd service file
    const systemdServiceContent = fs.readFileSync(
      path.join(__dirname, 'config', 'devnet.service'),
      'utf-8'
    );

    const systemdParameter = new ssm.StringParameter(this, 'DevnetSystemdService', {
      parameterName: '/devstack/systemd/devnet.service',
      stringValue: systemdServiceContent,
      description: 'Systemd service definition for Hourglass DevNet',
      tier: ssm.ParameterTier.STANDARD,
    });

    // Create SSM Parameter for CloudWatch Agent configuration
    const cloudwatchConfig = {
      agent: {
        metrics_collection_interval: 60,
        run_as_user: "cwagent"
      },
      logs: {
        logs_collected: {
          files: {
            collect_list: [
              {
                file_path: "/var/log/user-data.log",
                log_group_name: "/aws/ec2/devstack/user-data",
                log_stream_name: "{instance_id}",
                timezone: "UTC"
              },
              {
                file_path: "/var/log/devnet.log",
                log_group_name: "/aws/ec2/devstack/devnet",
                log_stream_name: "{instance_id}",
                timezone: "UTC"
              }
            ]
          }
        }
      }
    };

    const cloudwatchParameter = new ssm.StringParameter(this, 'CloudWatchAgentConfig', {
      parameterName: '/devstack/cloudwatch/agent-config.json',
      stringValue: JSON.stringify(cloudwatchConfig, null, 2),
      description: 'CloudWatch Agent configuration for DevStack',
      tier: ssm.ParameterTier.STANDARD,
    });

    // Create SSM Document for service management
    const serviceManagementDocument = new ssm.CfnDocument(this, 'DevnetServiceManagement', {
      documentType: 'Command',
      name: 'DevStack-ServiceManagement',
      documentFormat: 'YAML',
      content: {
        schemaVersion: '2.2',
        description: 'Manage DevNet service on DevStack instances',
        parameters: {
          action: {
            type: 'String',
            description: 'Action to perform on the service',
            allowedValues: ['start', 'stop', 'restart', 'status', 'logs'],
            default: 'status',
          },
          lines: {
            type: 'String',
            description: 'Number of log lines to show (only for logs action)',
            default: '50',
          },
        },
        mainSteps: [
          {
            action: 'aws:runShellScript',
            name: 'ManageService',
            inputs: {
              runCommand: [
                '#!/bin/bash',
                'ACTION="{{ action }}"',
                'LINES="{{ lines }}"',
                '',
                'case "$ACTION" in',
                '  start)',
                '    echo "Starting devnet service..."',
                '    sudo systemctl start devnet.service',
                '    sudo systemctl status devnet.service --no-pager',
                '    ;;',
                '  stop)',
                '    echo "Stopping devnet service..."',
                '    sudo systemctl stop devnet.service',
                '    sudo systemctl status devnet.service --no-pager',
                '    ;;',
                '  restart)',
                '    echo "Restarting devnet service..."',
                '    sudo systemctl restart devnet.service',
                '    sleep 5',
                '    sudo systemctl status devnet.service --no-pager',
                '    ;;',
                '  status)',
                '    echo "Devnet service status:"',
                '    sudo systemctl status devnet.service --no-pager',
                '    echo ""',
                '    echo "Port status:"',
                '    ss -tlnp | grep -E "(9090|8081|8545)" || echo "No services listening on expected ports"',
                '    ;;',
                '  logs)',
                '    echo "Recent devnet service logs (last $LINES lines):"',
                '    sudo journalctl -u devnet.service -n "$LINES" --no-pager',
                '    ;;',
                '  *)',
                '    echo "Invalid action: $ACTION"',
                '    exit 1',
                '    ;;',
                'esac',
              ],
            },
          },
        ],
      },
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

    // Add permissions for SSM Parameter Store access
    role.addToPolicy(new iam.PolicyStatement({
      actions: [
        'ssm:GetParameter',
        'ssm:GetParameters',
        'ssm:GetParameterHistory',
        'ssm:GetParametersByPath',
      ],
      resources: [
        `arn:aws:ssm:${this.region}:${this.account}:parameter/devstack/*`,
      ],
    }));

    // Add permission to check SSM agent status
    role.addToPolicy(new iam.PolicyStatement({
      actions: ['ssm:DescribeInstanceInformation'],
      resources: ['*'],
    }));

    // Add permissions for CloudWatch Logs
    role.addToPolicy(new iam.PolicyStatement({
      actions: [
        'logs:CreateLogGroup',
        'logs:CreateLogStream',
        'logs:PutLogEvents',
        'logs:DescribeLogStreams',
      ],
      resources: [
        `arn:aws:logs:${this.region}:${this.account}:log-group:/aws/ec2/devstack/*`,
      ],
    }));

    // EC2 Instance
    const instance = new ec2.Instance(this, 'DevstackInstance', {
      instanceType: new ec2.InstanceType(instanceType.valueAsString),
      machineImage: ec2.MachineImage.lookup({
        name: 'ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*',
        owners: ['099720109477'], // Canonical
      }),
      vpc,
      vpcSubnets: {
        subnetType: ec2.SubnetType.PUBLIC,
      },
      securityGroup,
      role,
      blockDevices: [{
        deviceName: '/dev/sda1',
        volume: ec2.BlockDeviceVolume.ebs(1000, {
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
      'sudo apt-get install -y ec2-instance-connect'
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

    new cdk.CfnOutput(this, 'InstanceId', {
      value: instance.instanceId,
      description: 'EC2 Instance ID',
    });
  }
}