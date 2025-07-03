import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as ecs from 'aws-cdk-lib/aws-ecs';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as autoscaling from 'aws-cdk-lib/aws-autoscaling';
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2';
import * as cloudwatch from 'aws-cdk-lib/aws-cloudwatch';
import * as logs from 'aws-cdk-lib/aws-logs';
import { Construct } from 'constructs';

export interface ObsidianHybridStackProps extends cdk.StackProps {
  minComputeNodes?: number;
  maxComputeNodes?: number;
  computeInstanceType?: ec2.InstanceType;
  vpcId?: string;
  controlPlane?: {
    orchestrator: {
      desiredCount: number;
      cpu: number;
      memory: number;
    };
    registry: {
      desiredCount: number;
      cpu: number;
      memory: number;
    };
    proxy: {
      desiredCount: number;
      cpu: number;
      memory: number;
    };
  };
}

export class ObsidianHybridStack extends cdk.Stack {
  public readonly vpc: ec2.IVpc;
  public readonly cluster: ecs.Cluster;
  public readonly alb: elbv2.ApplicationLoadBalancer;
  
  constructor(scope: Construct, id: string, props: ObsidianHybridStackProps = {}) {
    super(scope, id, props);

    // VPC
    this.vpc = props.vpcId 
      ? ec2.Vpc.fromLookup(this, 'VPC', { vpcId: props.vpcId })
      : new ec2.Vpc(this, 'ObsidianVPC', {
          maxAzs: 2,
          natGateways: 2,
        });

    // ECS Cluster for control plane
    this.cluster = new ecs.Cluster(this, 'ObsidianControlPlane', {
      vpc: this.vpc,
      containerInsights: true,
    });

    // Application Load Balancer
    this.alb = new elbv2.ApplicationLoadBalancer(this, 'ObsidianALB', {
      vpc: this.vpc,
      internetFacing: true,
    });

    // Deploy control plane components
    const orchestratorService = this.createOrchestratorService(props);
    const registryService = this.createRegistryService(props);
    const proxyService = this.createProxyService(props);

    // Deploy compute nodes
    const computeNodes = this.createComputeNodes(props);

    // Create CloudWatch Dashboard
    this.createDashboard(orchestratorService, registryService, proxyService);

    // Outputs
    new cdk.CfnOutput(this, 'ObsidianEndpoint', {
      value: this.alb.loadBalancerDnsName,
      description: 'Obsidian Gateway endpoint',
    });

    new cdk.CfnOutput(this, 'ClusterName', {
      value: this.cluster.clusterName,
      description: 'ECS Cluster name',
    });
  }

  private createOrchestratorService(props: ObsidianHybridStackProps): ecs.FargateService {
    const taskDefinition = new ecs.FargateTaskDefinition(this, 'OrchestratorTask', {
      cpu: props.controlPlane?.orchestrator?.cpu || 2048,
      memoryLimitMiB: props.controlPlane?.orchestrator?.memory || 4096,
    });

    taskDefinition.addContainer('orchestrator', {
      image: ecs.ContainerImage.fromRegistry('obsidian/orchestrator:latest'),
      logging: new ecs.AwsLogDriver({
        streamPrefix: 'orchestrator',
        logRetention: logs.RetentionDays.ONE_WEEK,
      }),
      environment: {
        SERVICE_NAME: 'orchestrator',
      },
      portMappings: [{
        containerPort: 9090,
        protocol: ecs.Protocol.TCP,
      }],
    });

    const service = new ecs.FargateService(this, 'OrchestratorService', {
      cluster: this.cluster,
      taskDefinition,
      desiredCount: props.controlPlane?.orchestrator?.desiredCount || 2,
      serviceName: 'obsidian-orchestrator',
    });

    // Add ALB target
    const targetGroup = new elbv2.ApplicationTargetGroup(this, 'OrchestratorTargetGroup', {
      vpc: this.vpc,
      port: 9090,
      protocol: elbv2.ApplicationProtocol.HTTP,
      targetType: elbv2.TargetType.IP,
      healthCheck: {
        path: '/health',
        interval: cdk.Duration.seconds(30),
      },
    });

    service.attachToApplicationTargetGroup(targetGroup);

    this.alb.addListener('OrchestratorListener', {
      port: 9090,
      defaultTargetGroups: [targetGroup],
    });

    return service;
  }

  private createRegistryService(props: ObsidianHybridStackProps): ecs.FargateService {
    const taskDefinition = new ecs.FargateTaskDefinition(this, 'RegistryTask', {
      cpu: props.controlPlane?.registry?.cpu || 1024,
      memoryLimitMiB: props.controlPlane?.registry?.memory || 2048,
    });

    // Add EFS volume for image cache
    const fileSystem = new cdk.aws_efs.FileSystem(this, 'RegistryCache', {
      vpc: this.vpc,
      encrypted: true,
      performanceMode: cdk.aws_efs.PerformanceMode.GENERAL_PURPOSE,
    });

    taskDefinition.addVolume({
      name: 'cache',
      efsVolumeConfiguration: {
        fileSystemId: fileSystem.fileSystemId,
      },
    });

    const container = taskDefinition.addContainer('registry', {
      image: ecs.ContainerImage.fromRegistry('obsidian/registry:latest'),
      logging: new ecs.AwsLogDriver({
        streamPrefix: 'registry',
        logRetention: logs.RetentionDays.ONE_WEEK,
      }),
      environment: {
        SERVICE_NAME: 'registry',
        CACHE_PATH: '/cache',
      },
      portMappings: [{
        containerPort: 9091,
        protocol: ecs.Protocol.TCP,
      }],
    });

    container.addMountPoints({
      sourceVolume: 'cache',
      containerPath: '/cache',
      readOnly: false,
    });

    const service = new ecs.FargateService(this, 'RegistryService', {
      cluster: this.cluster,
      taskDefinition,
      desiredCount: props.controlPlane?.registry?.desiredCount || 2,
      serviceName: 'obsidian-registry',
    });

    return service;
  }

  private createProxyService(props: ObsidianHybridStackProps): ecs.FargateService {
    const taskDefinition = new ecs.FargateTaskDefinition(this, 'ProxyTask', {
      cpu: props.controlPlane?.proxy?.cpu || 512,
      memoryLimitMiB: props.controlPlane?.proxy?.memory || 1024,
    });

    taskDefinition.addContainer('proxy', {
      image: ecs.ContainerImage.fromRegistry('obsidian/proxy:latest'),
      logging: new ecs.AwsLogDriver({
        streamPrefix: 'proxy',
        logRetention: logs.RetentionDays.ONE_WEEK,
      }),
      environment: {
        SERVICE_NAME: 'proxy',
      },
      portMappings: [{
        containerPort: 8090,
        protocol: ecs.Protocol.TCP,
      }],
    });

    const service = new ecs.FargateService(this, 'ProxyService', {
      cluster: this.cluster,
      taskDefinition,
      desiredCount: props.controlPlane?.proxy?.desiredCount || 3,
      serviceName: 'obsidian-proxy',
    });

    return service;
  }

  private createComputeNodes(props: ObsidianHybridStackProps): autoscaling.AutoScalingGroup {
    // Security Group for compute nodes
    const computeSecurityGroup = new ec2.SecurityGroup(this, 'ComputeSecurityGroup', {
      vpc: this.vpc,
      description: 'Security group for Obsidian compute nodes',
      allowAllOutbound: true,
    });

    // Allow communication from control plane
    computeSecurityGroup.addIngressRule(
      ec2.Peer.ipv4(this.vpc.vpcCidrBlock),
      ec2.Port.allTraffic(),
      'Allow all traffic from VPC'
    );

    // IAM Role for compute nodes
    const computeRole = new iam.Role(this, 'ComputeNodeRole', {
      assumedBy: new iam.ServicePrincipal('ec2.amazonaws.com'),
      managedPolicies: [
        iam.ManagedPolicy.fromAwsManagedPolicyName('AmazonSSMManagedInstanceCore'),
      ],
    });

    // Add Docker permissions
    computeRole.addToPolicy(new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: [
        'ecr:GetAuthorizationToken',
        'ecr:BatchCheckLayerAvailability',
        'ecr:GetDownloadUrlForLayer',
        'ecr:BatchGetImage',
      ],
      resources: ['*'],
    }));

    // User Data for compute nodes
    const userData = ec2.UserData.forLinux();
    userData.addCommands(
      'yum update -y',
      'yum install -y docker',
      'service docker start',
      'usermod -a -G docker ec2-user',
      
      // Configure Docker for container execution
      'cat > /etc/docker/daemon.json << EOF',
      '{',
      '  "log-driver": "json-file",',
      '  "log-opts": {',
      '    "max-size": "10m",',
      '    "max-file": "3"',
      '  },',
      '  "default-ulimits": {',
      '    "nofile": {',
      '      "Name": "nofile",',
      '      "Hard": 65536,',
      '      "Soft": 65536',
      '    }',
      '  }',
      '}',
      'EOF',
      
      'systemctl restart docker'
    );

    // Launch Template for compute nodes
    const launchTemplate = new ec2.LaunchTemplate(this, 'ComputeLaunchTemplate', {
      instanceType: props.computeInstanceType || ec2.InstanceType.of(ec2.InstanceClass.M5, ec2.InstanceSize.XLARGE2),
      machineImage: ec2.MachineImage.latestAmazonLinux2(),
      userData,
      role: computeRole,
      securityGroup: computeSecurityGroup,
      blockDevices: [{
        deviceName: '/dev/xvda',
        volume: ec2.BlockDeviceVolume.ebs(200, {
          volumeType: ec2.EbsDeviceVolumeType.GP3,
          encrypted: true,
        }),
      }],
    });

    // Auto Scaling Group for compute nodes
    const asg = new autoscaling.AutoScalingGroup(this, 'ComputeASG', {
      vpc: this.vpc,
      launchTemplate,
      minCapacity: props.minComputeNodes || 3,
      maxCapacity: props.maxComputeNodes || 20,
      healthCheck: autoscaling.HealthCheck.ec2({
        grace: cdk.Duration.minutes(5),
      }),
    });

    // Auto-scaling based on CPU
    asg.scaleOnCpuUtilization('ScaleOnCPU', {
      targetUtilizationPercent: 70,
      cooldown: cdk.Duration.minutes(5),
    });

    return asg;
  }

  private createDashboard(
    orchestrator: ecs.FargateService,
    registry: ecs.FargateService,
    proxy: ecs.FargateService
  ): cloudwatch.Dashboard {
    const dashboard = new cloudwatch.Dashboard(this, 'ObsidianDashboard', {
      dashboardName: 'obsidian-production',
    });

    // Orchestrator metrics
    dashboard.addWidgets(
      new cloudwatch.GraphWidget({
        title: 'Orchestrator CPU Utilization',
        left: [orchestrator.metricCpuUtilization()],
      }),
      new cloudwatch.GraphWidget({
        title: 'Orchestrator Memory Utilization',
        left: [orchestrator.metricMemoryUtilization()],
      })
    );

    // Registry metrics
    dashboard.addWidgets(
      new cloudwatch.GraphWidget({
        title: 'Registry CPU Utilization',
        left: [registry.metricCpuUtilization()],
      }),
      new cloudwatch.GraphWidget({
        title: 'Registry Memory Utilization',
        left: [registry.metricMemoryUtilization()],
      })
    );

    // Proxy metrics
    dashboard.addWidgets(
      new cloudwatch.GraphWidget({
        title: 'Proxy CPU Utilization',
        left: [proxy.metricCpuUtilization()],
      }),
      new cloudwatch.GraphWidget({
        title: 'Proxy Memory Utilization',
        left: [proxy.metricMemoryUtilization()],
      })
    );

    return dashboard;
  }
}