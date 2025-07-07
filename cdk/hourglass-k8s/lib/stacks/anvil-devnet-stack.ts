import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as ecs from 'aws-cdk-lib/aws-ecs';
import * as efs from 'aws-cdk-lib/aws-efs';
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2';
import * as ssm from 'aws-cdk-lib/aws-ssm';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as path from 'path';
import { Construct } from 'constructs';
import { ContractDeployer } from '../constructs/contract-deployer';

interface AnvilDevnetStackProps extends cdk.StackProps {
  vpc: ec2.Vpc;
}

export class AnvilDevnetStack extends cdk.Stack {
  public readonly l1Endpoint: string;
  public readonly l2Endpoint: string;
  public readonly contractAddresses: { [key: string]: string };

  constructor(scope: Construct, id: string, props: AnvilDevnetStackProps) {
    super(scope, id, props);

    const { vpc } = props;

    // Read configuration
    const devnetConfig = require('../../config/devnet.json');
    const accounts = require('../../config/accounts.json');

    // Create ECS cluster
    const cluster = new ecs.Cluster(this, 'AnvilCluster', {
      vpc,
      clusterName: 'hourglass-anvil-devnet',
    });

    // Create EFS for persistent state
    const fileSystem = new efs.FileSystem(this, 'AnvilStateFS', {
      vpc,
      lifecyclePolicy: efs.LifecyclePolicy.AFTER_14_DAYS,
      performanceMode: efs.PerformanceMode.GENERAL_PURPOSE,
      throughputMode: efs.ThroughputMode.BURSTING,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    // Create task definition for Anvil L1
    const anvilL1Task = new ecs.FargateTaskDefinition(this, 'AnvilL1Task', {
      memoryLimitMiB: 4096,
      cpu: 2048,
    });

    // Add EFS volume
    anvilL1Task.addVolume({
      name: 'anvil-state',
      efsVolumeConfiguration: {
        fileSystemId: fileSystem.fileSystemId,
        rootDirectory: '/anvil-l1',
      },
    });

    // Create Anvil L1 container
    const anvilL1Container = anvilL1Task.addContainer('anvil-l1', {
      image: ecs.ContainerImage.fromRegistry('ghcr.io/foundry-rs/foundry:latest'),
      command: [
        'anvil',
        '--fork-url', devnetConfig.l1.forkUrl,
        '--fork-block-number', devnetConfig.l1.forkBlock.toString(),
        '--chain-id', devnetConfig.l1.chainId.toString(),
        '--host', '0.0.0.0',
        '--port', '8545',
        '--block-time', devnetConfig.l1.blockTime.toString(),
        '--state', '/anvil-state/state.json',
        '--state-interval', '60',
      ],
      environment: {
        ANVIL_IP_ADDR: '0.0.0.0',
      },
      logging: ecs.LogDrivers.awsLogs({
        streamPrefix: 'anvil-l1',
      }),
      portMappings: [
        {
          containerPort: 8545,
          protocol: ecs.Protocol.TCP,
        },
      ],
    });

    // Mount EFS
    anvilL1Container.addMountPoints({
      sourceVolume: 'anvil-state',
      containerPath: '/anvil-state',
      readOnly: false,
    });

    // Create Anvil L1 service
    const anvilL1Service = new ecs.FargateService(this, 'AnvilL1Service', {
      cluster,
      taskDefinition: anvilL1Task,
      desiredCount: 1,
      assignPublicIp: false,
      serviceName: 'anvil-l1',
    });

    // Allow EFS access
    fileSystem.connections.allowDefaultPortFrom(anvilL1Service);

    // Create NLB for Anvil L1
    const anvilL1Nlb = new elbv2.NetworkLoadBalancer(this, 'AnvilL1NLB', {
      vpc,
      internetFacing: false,
      scheme: elbv2.LoadBalancerScheme.INTERNAL,
    });

    const anvilL1Listener = anvilL1Nlb.addListener('AnvilL1Listener', {
      port: 8545,
      protocol: elbv2.Protocol.TCP,
    });

    anvilL1Listener.addTargets('AnvilL1Targets', {
      port: 8545,
      protocol: elbv2.Protocol.TCP,
      targets: [anvilL1Service],
      healthCheck: {
        enabled: true,
        protocol: elbv2.Protocol.TCP,
      },
    });

    // Create task definition for Anvil L2
    const anvilL2Task = new ecs.FargateTaskDefinition(this, 'AnvilL2Task', {
      memoryLimitMiB: 4096,
      cpu: 2048,
    });

    // Add EFS volume for L2
    anvilL2Task.addVolume({
      name: 'anvil-state',
      efsVolumeConfiguration: {
        fileSystemId: fileSystem.fileSystemId,
        rootDirectory: '/anvil-l2',
      },
    });

    // Create Anvil L2 container
    const anvilL2Container = anvilL2Task.addContainer('anvil-l2', {
      image: ecs.ContainerImage.fromRegistry('ghcr.io/foundry-rs/foundry:latest'),
      command: [
        'anvil',
        '--fork-url', devnetConfig.l2.forkUrl,
        '--fork-block-number', devnetConfig.l2.forkBlock.toString(),
        '--chain-id', devnetConfig.l2.chainId.toString(),
        '--host', '0.0.0.0',
        '--port', '8545',
        '--block-time', devnetConfig.l2.blockTime.toString(),
        '--state', '/anvil-state/state.json',
        '--state-interval', '60',
      ],
      environment: {
        ANVIL_IP_ADDR: '0.0.0.0',
      },
      logging: ecs.LogDrivers.awsLogs({
        streamPrefix: 'anvil-l2',
      }),
      portMappings: [
        {
          containerPort: 8545,
          protocol: ecs.Protocol.TCP,
        },
      ],
    });

    // Mount EFS
    anvilL2Container.addMountPoints({
      sourceVolume: 'anvil-state',
      containerPath: '/anvil-state',
      readOnly: false,
    });

    // Create Anvil L2 service
    const anvilL2Service = new ecs.FargateService(this, 'AnvilL2Service', {
      cluster,
      taskDefinition: anvilL2Task,
      desiredCount: 1,
      assignPublicIp: false,
      serviceName: 'anvil-l2',
    });

    // Allow EFS access
    fileSystem.connections.allowDefaultPortFrom(anvilL2Service);

    // Create NLB for Anvil L2
    const anvilL2Nlb = new elbv2.NetworkLoadBalancer(this, 'AnvilL2NLB', {
      vpc,
      internetFacing: false,
      scheme: elbv2.LoadBalancerScheme.INTERNAL,
    });

    const anvilL2Listener = anvilL2Nlb.addListener('AnvilL2Listener', {
      port: 8545,
      protocol: elbv2.Protocol.TCP,
    });

    anvilL2Listener.addTargets('AnvilL2Targets', {
      port: 8545,
      protocol: elbv2.Protocol.TCP,
      targets: [anvilL2Service],
      healthCheck: {
        enabled: true,
        protocol: elbv2.Protocol.TCP,
      },
    });

    // Store endpoints
    this.l1Endpoint = `http://${anvilL1Nlb.loadBalancerDnsName}:8545`;
    this.l2Endpoint = `http://${anvilL2Nlb.loadBalancerDnsName}:8545`;

    // Deploy contracts using Lambda
    const contractDeployer = new ContractDeployer(this, 'ContractDeployer', {
      vpc,
      l1Endpoint: this.l1Endpoint,
      l2Endpoint: this.l2Endpoint,
      accounts,
      devnetConfig,
    });

    // Store contract addresses
    this.contractAddresses = {
      mailboxL1: contractDeployer.mailboxL1Address,
      mailboxL2: contractDeployer.mailboxL2Address,
      avsTaskRegistrar: contractDeployer.avsTaskRegistrarAddress,
      taskHookL1: contractDeployer.taskHookL1Address,
      taskHookL2: contractDeployer.taskHookL2Address,
    };

    // Store endpoints in Parameter Store
    new ssm.StringParameter(this, 'L1EndpointParam', {
      parameterName: '/hourglass/devnet/l1-endpoint',
      stringValue: this.l1Endpoint,
    });

    new ssm.StringParameter(this, 'L2EndpointParam', {
      parameterName: '/hourglass/devnet/l2-endpoint',
      stringValue: this.l2Endpoint,
    });

    // Store contract addresses in Parameter Store
    Object.entries(this.contractAddresses).forEach(([key, value]) => {
      new ssm.StringParameter(this, `${key}Param`, {
        parameterName: `/hourglass/contracts/${key}`,
        stringValue: value,
      });
    });

    // Outputs
    new cdk.CfnOutput(this, 'AnvilL1Endpoint', {
      value: this.l1Endpoint,
      description: 'Anvil L1 RPC endpoint',
    });

    new cdk.CfnOutput(this, 'AnvilL2Endpoint', {
      value: this.l2Endpoint,
      description: 'Anvil L2 RPC endpoint',
    });

    new cdk.CfnOutput(this, 'ContractAddresses', {
      value: JSON.stringify(this.contractAddresses, null, 2),
      description: 'Deployed contract addresses',
    });
  }
}