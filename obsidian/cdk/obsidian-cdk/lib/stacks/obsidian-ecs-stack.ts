import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as ecs from 'aws-cdk-lib/aws-ecs';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2';
import * as logs from 'aws-cdk-lib/aws-logs';
import { Construct } from 'constructs';

export interface ObsidianECSStackProps extends cdk.StackProps {
  vpcId?: string;
  orchestratorConfig?: {
    desiredCount: number;
    cpu: number;
    memory: number;
  };
  registryConfig?: {
    desiredCount: number;
    cpu: number;
    memory: number;
  };
  proxyConfig?: {
    desiredCount: number;
    cpu: number;
    memory: number;
  };
}

export class ObsidianECSStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: ObsidianECSStackProps = {}) {
    super(scope, id, props);

    // VPC
    const vpc = props.vpcId 
      ? ec2.Vpc.fromLookup(this, 'VPC', { vpcId: props.vpcId })
      : new ec2.Vpc(this, 'ObsidianVPC', {
          maxAzs: 2,
          natGateways: 1,
        });

    // ECS Cluster
    const cluster = new ecs.Cluster(this, 'ObsidianCluster', {
      vpc,
      containerInsights: true,
    });

    // Application Load Balancer
    const alb = new elbv2.ApplicationLoadBalancer(this, 'ObsidianALB', {
      vpc,
      internetFacing: true,
    });

    // Gateway Task Definition (all services in one container)
    const gatewayTaskDef = new ecs.FargateTaskDefinition(this, 'GatewayTask', {
      cpu: 4096,
      memoryLimitMiB: 8192,
    });

    // IAM permissions for the task
    gatewayTaskDef.addToTaskRolePolicy(new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: [
        'ecr:GetAuthorizationToken',
        'ecr:BatchCheckLayerAvailability',
        'ecr:GetDownloadUrlForLayer',
        'ecr:BatchGetImage',
      ],
      resources: ['*'],
    }));

    gatewayTaskDef.addToTaskRolePolicy(new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: [
        'secretsmanager:GetSecretValue',
      ],
      resources: [`arn:aws:secretsmanager:${this.region}:${this.account}:secret:obsidian/*`],
    }));

    // Gateway Container
    const container = gatewayTaskDef.addContainer('gateway', {
      image: ecs.ContainerImage.fromRegistry('obsidian/gateway:latest'),
      logging: new ecs.AwsLogDriver({
        streamPrefix: 'gateway',
        logRetention: logs.RetentionDays.ONE_WEEK,
      }),
      environment: {
        ENVIRONMENT: 'staging',
        AWS_REGION: this.region,
      },
      portMappings: [
        {
          containerPort: 8080,
          protocol: ecs.Protocol.TCP,
        },
        {
          containerPort: 9090,
          protocol: ecs.Protocol.TCP,
        },
        {
          containerPort: 9091,
          protocol: ecs.Protocol.TCP,
        },
      ],
      healthCheck: {
        command: ['CMD-SHELL', 'curl -f http://localhost:8080/health || exit 1'],
        interval: cdk.Duration.seconds(30),
        timeout: cdk.Duration.seconds(5),
        retries: 3,
      },
    });

    // Gateway Service
    const service = new ecs.FargateService(this, 'GatewayService', {
      cluster,
      taskDefinition: gatewayTaskDef,
      desiredCount: 2,
      serviceName: 'obsidian-gateway',
      assignPublicIp: false,
    });

    // Configure ALB target groups and listeners
    
    // HTTP target group
    const httpTargetGroup = new elbv2.ApplicationTargetGroup(this, 'HttpTargetGroup', {
      vpc,
      port: 8080,
      protocol: elbv2.ApplicationProtocol.HTTP,
      targetType: elbv2.TargetType.IP,
      healthCheck: {
        path: '/health',
        interval: cdk.Duration.seconds(30),
        timeout: cdk.Duration.seconds(5),
        healthyThresholdCount: 2,
        unhealthyThresholdCount: 3,
      },
    });

    // gRPC target group
    const grpcTargetGroup = new elbv2.ApplicationTargetGroup(this, 'GrpcTargetGroup', {
      vpc,
      port: 9090,
      protocol: elbv2.ApplicationProtocol.HTTP,
      targetType: elbv2.TargetType.IP,
      protocolVersion: elbv2.ApplicationProtocolVersion.GRPC,
      healthCheck: {
        enabled: true,
        protocol: elbv2.Protocol.HTTP,
        path: '/grpc.health.v1.Health/Check',
        interval: cdk.Duration.seconds(30),
        timeout: cdk.Duration.seconds(5),
        healthyThresholdCount: 2,
        unhealthyThresholdCount: 3,
      },
    });

    // Metrics target group
    const metricsTargetGroup = new elbv2.ApplicationTargetGroup(this, 'MetricsTargetGroup', {
      vpc,
      port: 9091,
      protocol: elbv2.ApplicationProtocol.HTTP,
      targetType: elbv2.TargetType.IP,
      healthCheck: {
        path: '/metrics',
        interval: cdk.Duration.seconds(30),
        timeout: cdk.Duration.seconds(5),
        healthyThresholdCount: 2,
        unhealthyThresholdCount: 3,
      },
    });

    // Attach service to target groups
    service.attachToApplicationTargetGroup(httpTargetGroup);
    service.attachToApplicationTargetGroup(grpcTargetGroup);
    service.attachToApplicationTargetGroup(metricsTargetGroup);

    // ALB Listeners
    alb.addListener('HttpListener', {
      port: 80,
      defaultTargetGroups: [httpTargetGroup],
    });

    alb.addListener('GrpcListener', {
      port: 9090,
      defaultTargetGroups: [grpcTargetGroup],
    });

    alb.addListener('MetricsListener', {
      port: 9091,
      defaultTargetGroups: [metricsTargetGroup],
    });

    // Auto-scaling
    const scaling = service.autoScaleTaskCount({
      minCapacity: 2,
      maxCapacity: 10,
    });

    scaling.scaleOnCpuUtilization('CpuScaling', {
      targetUtilizationPercent: 70,
      scaleInCooldown: cdk.Duration.seconds(60),
      scaleOutCooldown: cdk.Duration.seconds(60),
    });

    scaling.scaleOnMemoryUtilization('MemoryScaling', {
      targetUtilizationPercent: 80,
      scaleInCooldown: cdk.Duration.seconds(60),
      scaleOutCooldown: cdk.Duration.seconds(60),
    });

    // Outputs
    new cdk.CfnOutput(this, 'ObsidianEndpoint', {
      value: alb.loadBalancerDnsName,
      description: 'Obsidian Gateway endpoint',
    });

    new cdk.CfnOutput(this, 'HttpURL', {
      value: `http://${alb.loadBalancerDnsName}`,
      description: 'HTTP endpoint',
    });

    new cdk.CfnOutput(this, 'GrpcEndpoint', {
      value: `${alb.loadBalancerDnsName}:9090`,
      description: 'gRPC endpoint',
    });

    new cdk.CfnOutput(this, 'MetricsEndpoint', {
      value: `http://${alb.loadBalancerDnsName}:9091/metrics`,
      description: 'Prometheus metrics endpoint',
    });

    new cdk.CfnOutput(this, 'ClusterName', {
      value: cluster.clusterName,
      description: 'ECS Cluster name',
    });

    new cdk.CfnOutput(this, 'ServiceName', {
      value: service.serviceName,
      description: 'ECS Service name',
    });
  }
}