"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ObsidianHybridStack = void 0;
const cdk = require("aws-cdk-lib");
const ec2 = require("aws-cdk-lib/aws-ec2");
const ecs = require("aws-cdk-lib/aws-ecs");
const iam = require("aws-cdk-lib/aws-iam");
const autoscaling = require("aws-cdk-lib/aws-autoscaling");
const elbv2 = require("aws-cdk-lib/aws-elasticloadbalancingv2");
const cloudwatch = require("aws-cdk-lib/aws-cloudwatch");
const logs = require("aws-cdk-lib/aws-logs");
class ObsidianHybridStack extends cdk.Stack {
    constructor(scope, id, props = {}) {
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
    createOrchestratorService(props) {
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
    createRegistryService(props) {
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
    createProxyService(props) {
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
    createComputeNodes(props) {
        // Security Group for compute nodes
        const computeSecurityGroup = new ec2.SecurityGroup(this, 'ComputeSecurityGroup', {
            vpc: this.vpc,
            description: 'Security group for Obsidian compute nodes',
            allowAllOutbound: true,
        });
        // Allow communication from control plane
        computeSecurityGroup.addIngressRule(ec2.Peer.ipv4(this.vpc.vpcCidrBlock), ec2.Port.allTraffic(), 'Allow all traffic from VPC');
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
        userData.addCommands('yum update -y', 'yum install -y docker', 'service docker start', 'usermod -a -G docker ec2-user', 
        // Configure Docker for container execution
        'cat > /etc/docker/daemon.json << EOF', '{', '  "log-driver": "json-file",', '  "log-opts": {', '    "max-size": "10m",', '    "max-file": "3"', '  },', '  "default-ulimits": {', '    "nofile": {', '      "Name": "nofile",', '      "Hard": 65536,', '      "Soft": 65536', '    }', '  }', '}', 'EOF', 'systemctl restart docker');
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
    createDashboard(orchestrator, registry, proxy) {
        const dashboard = new cloudwatch.Dashboard(this, 'ObsidianDashboard', {
            dashboardName: 'obsidian-production',
        });
        // Orchestrator metrics
        dashboard.addWidgets(new cloudwatch.GraphWidget({
            title: 'Orchestrator CPU Utilization',
            left: [orchestrator.metricCpuUtilization()],
        }), new cloudwatch.GraphWidget({
            title: 'Orchestrator Memory Utilization',
            left: [orchestrator.metricMemoryUtilization()],
        }));
        // Registry metrics
        dashboard.addWidgets(new cloudwatch.GraphWidget({
            title: 'Registry CPU Utilization',
            left: [registry.metricCpuUtilization()],
        }), new cloudwatch.GraphWidget({
            title: 'Registry Memory Utilization',
            left: [registry.metricMemoryUtilization()],
        }));
        // Proxy metrics
        dashboard.addWidgets(new cloudwatch.GraphWidget({
            title: 'Proxy CPU Utilization',
            left: [proxy.metricCpuUtilization()],
        }), new cloudwatch.GraphWidget({
            title: 'Proxy Memory Utilization',
            left: [proxy.metricMemoryUtilization()],
        }));
        return dashboard;
    }
}
exports.ObsidianHybridStack = ObsidianHybridStack;
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoib2JzaWRpYW4taHlicmlkLXN0YWNrLmpzIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsib2JzaWRpYW4taHlicmlkLXN0YWNrLnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiI7OztBQUFBLG1DQUFtQztBQUNuQywyQ0FBMkM7QUFDM0MsMkNBQTJDO0FBQzNDLDJDQUEyQztBQUMzQywyREFBMkQ7QUFDM0QsZ0VBQWdFO0FBQ2hFLHlEQUF5RDtBQUN6RCw2Q0FBNkM7QUEyQjdDLE1BQWEsbUJBQW9CLFNBQVEsR0FBRyxDQUFDLEtBQUs7SUFLaEQsWUFBWSxLQUFnQixFQUFFLEVBQVUsRUFBRSxRQUFrQyxFQUFFO1FBQzVFLEtBQUssQ0FBQyxLQUFLLEVBQUUsRUFBRSxFQUFFLEtBQUssQ0FBQyxDQUFDO1FBRXhCLE1BQU07UUFDTixJQUFJLENBQUMsR0FBRyxHQUFHLEtBQUssQ0FBQyxLQUFLO1lBQ3BCLENBQUMsQ0FBQyxHQUFHLENBQUMsR0FBRyxDQUFDLFVBQVUsQ0FBQyxJQUFJLEVBQUUsS0FBSyxFQUFFLEVBQUUsS0FBSyxFQUFFLEtBQUssQ0FBQyxLQUFLLEVBQUUsQ0FBQztZQUN6RCxDQUFDLENBQUMsSUFBSSxHQUFHLENBQUMsR0FBRyxDQUFDLElBQUksRUFBRSxhQUFhLEVBQUU7Z0JBQy9CLE1BQU0sRUFBRSxDQUFDO2dCQUNULFdBQVcsRUFBRSxDQUFDO2FBQ2YsQ0FBQyxDQUFDO1FBRVAsZ0NBQWdDO1FBQ2hDLElBQUksQ0FBQyxPQUFPLEdBQUcsSUFBSSxHQUFHLENBQUMsT0FBTyxDQUFDLElBQUksRUFBRSxzQkFBc0IsRUFBRTtZQUMzRCxHQUFHLEVBQUUsSUFBSSxDQUFDLEdBQUc7WUFDYixpQkFBaUIsRUFBRSxJQUFJO1NBQ3hCLENBQUMsQ0FBQztRQUVILDRCQUE0QjtRQUM1QixJQUFJLENBQUMsR0FBRyxHQUFHLElBQUksS0FBSyxDQUFDLHVCQUF1QixDQUFDLElBQUksRUFBRSxhQUFhLEVBQUU7WUFDaEUsR0FBRyxFQUFFLElBQUksQ0FBQyxHQUFHO1lBQ2IsY0FBYyxFQUFFLElBQUk7U0FDckIsQ0FBQyxDQUFDO1FBRUgsa0NBQWtDO1FBQ2xDLE1BQU0sbUJBQW1CLEdBQUcsSUFBSSxDQUFDLHlCQUF5QixDQUFDLEtBQUssQ0FBQyxDQUFDO1FBQ2xFLE1BQU0sZUFBZSxHQUFHLElBQUksQ0FBQyxxQkFBcUIsQ0FBQyxLQUFLLENBQUMsQ0FBQztRQUMxRCxNQUFNLFlBQVksR0FBRyxJQUFJLENBQUMsa0JBQWtCLENBQUMsS0FBSyxDQUFDLENBQUM7UUFFcEQsdUJBQXVCO1FBQ3ZCLE1BQU0sWUFBWSxHQUFHLElBQUksQ0FBQyxrQkFBa0IsQ0FBQyxLQUFLLENBQUMsQ0FBQztRQUVwRCw4QkFBOEI7UUFDOUIsSUFBSSxDQUFDLGVBQWUsQ0FBQyxtQkFBbUIsRUFBRSxlQUFlLEVBQUUsWUFBWSxDQUFDLENBQUM7UUFFekUsVUFBVTtRQUNWLElBQUksR0FBRyxDQUFDLFNBQVMsQ0FBQyxJQUFJLEVBQUUsa0JBQWtCLEVBQUU7WUFDMUMsS0FBSyxFQUFFLElBQUksQ0FBQyxHQUFHLENBQUMsbUJBQW1CO1lBQ25DLFdBQVcsRUFBRSwyQkFBMkI7U0FDekMsQ0FBQyxDQUFDO1FBRUgsSUFBSSxHQUFHLENBQUMsU0FBUyxDQUFDLElBQUksRUFBRSxhQUFhLEVBQUU7WUFDckMsS0FBSyxFQUFFLElBQUksQ0FBQyxPQUFPLENBQUMsV0FBVztZQUMvQixXQUFXLEVBQUUsa0JBQWtCO1NBQ2hDLENBQUMsQ0FBQztJQUNMLENBQUM7SUFFTyx5QkFBeUIsQ0FBQyxLQUErQjtRQUMvRCxNQUFNLGNBQWMsR0FBRyxJQUFJLEdBQUcsQ0FBQyxxQkFBcUIsQ0FBQyxJQUFJLEVBQUUsa0JBQWtCLEVBQUU7WUFDN0UsR0FBRyxFQUFFLEtBQUssQ0FBQyxZQUFZLEVBQUUsWUFBWSxFQUFFLEdBQUcsSUFBSSxJQUFJO1lBQ2xELGNBQWMsRUFBRSxLQUFLLENBQUMsWUFBWSxFQUFFLFlBQVksRUFBRSxNQUFNLElBQUksSUFBSTtTQUNqRSxDQUFDLENBQUM7UUFFSCxjQUFjLENBQUMsWUFBWSxDQUFDLGNBQWMsRUFBRTtZQUMxQyxLQUFLLEVBQUUsR0FBRyxDQUFDLGNBQWMsQ0FBQyxZQUFZLENBQUMsOEJBQThCLENBQUM7WUFDdEUsT0FBTyxFQUFFLElBQUksR0FBRyxDQUFDLFlBQVksQ0FBQztnQkFDNUIsWUFBWSxFQUFFLGNBQWM7Z0JBQzVCLFlBQVksRUFBRSxJQUFJLENBQUMsYUFBYSxDQUFDLFFBQVE7YUFDMUMsQ0FBQztZQUNGLFdBQVcsRUFBRTtnQkFDWCxZQUFZLEVBQUUsY0FBYzthQUM3QjtZQUNELFlBQVksRUFBRSxDQUFDO29CQUNiLGFBQWEsRUFBRSxJQUFJO29CQUNuQixRQUFRLEVBQUUsR0FBRyxDQUFDLFFBQVEsQ0FBQyxHQUFHO2lCQUMzQixDQUFDO1NBQ0gsQ0FBQyxDQUFDO1FBRUgsTUFBTSxPQUFPLEdBQUcsSUFBSSxHQUFHLENBQUMsY0FBYyxDQUFDLElBQUksRUFBRSxxQkFBcUIsRUFBRTtZQUNsRSxPQUFPLEVBQUUsSUFBSSxDQUFDLE9BQU87WUFDckIsY0FBYztZQUNkLFlBQVksRUFBRSxLQUFLLENBQUMsWUFBWSxFQUFFLFlBQVksRUFBRSxZQUFZLElBQUksQ0FBQztZQUNqRSxXQUFXLEVBQUUsdUJBQXVCO1NBQ3JDLENBQUMsQ0FBQztRQUVILGlCQUFpQjtRQUNqQixNQUFNLFdBQVcsR0FBRyxJQUFJLEtBQUssQ0FBQyxzQkFBc0IsQ0FBQyxJQUFJLEVBQUUseUJBQXlCLEVBQUU7WUFDcEYsR0FBRyxFQUFFLElBQUksQ0FBQyxHQUFHO1lBQ2IsSUFBSSxFQUFFLElBQUk7WUFDVixRQUFRLEVBQUUsS0FBSyxDQUFDLG1CQUFtQixDQUFDLElBQUk7WUFDeEMsVUFBVSxFQUFFLEtBQUssQ0FBQyxVQUFVLENBQUMsRUFBRTtZQUMvQixXQUFXLEVBQUU7Z0JBQ1gsSUFBSSxFQUFFLFNBQVM7Z0JBQ2YsUUFBUSxFQUFFLEdBQUcsQ0FBQyxRQUFRLENBQUMsT0FBTyxDQUFDLEVBQUUsQ0FBQzthQUNuQztTQUNGLENBQUMsQ0FBQztRQUVILE9BQU8sQ0FBQyw4QkFBOEIsQ0FBQyxXQUFXLENBQUMsQ0FBQztRQUVwRCxJQUFJLENBQUMsR0FBRyxDQUFDLFdBQVcsQ0FBQyxzQkFBc0IsRUFBRTtZQUMzQyxJQUFJLEVBQUUsSUFBSTtZQUNWLG1CQUFtQixFQUFFLENBQUMsV0FBVyxDQUFDO1NBQ25DLENBQUMsQ0FBQztRQUVILE9BQU8sT0FBTyxDQUFDO0lBQ2pCLENBQUM7SUFFTyxxQkFBcUIsQ0FBQyxLQUErQjtRQUMzRCxNQUFNLGNBQWMsR0FBRyxJQUFJLEdBQUcsQ0FBQyxxQkFBcUIsQ0FBQyxJQUFJLEVBQUUsY0FBYyxFQUFFO1lBQ3pFLEdBQUcsRUFBRSxLQUFLLENBQUMsWUFBWSxFQUFFLFFBQVEsRUFBRSxHQUFHLElBQUksSUFBSTtZQUM5QyxjQUFjLEVBQUUsS0FBSyxDQUFDLFlBQVksRUFBRSxRQUFRLEVBQUUsTUFBTSxJQUFJLElBQUk7U0FDN0QsQ0FBQyxDQUFDO1FBRUgsaUNBQWlDO1FBQ2pDLE1BQU0sVUFBVSxHQUFHLElBQUksR0FBRyxDQUFDLE9BQU8sQ0FBQyxVQUFVLENBQUMsSUFBSSxFQUFFLGVBQWUsRUFBRTtZQUNuRSxHQUFHLEVBQUUsSUFBSSxDQUFDLEdBQUc7WUFDYixTQUFTLEVBQUUsSUFBSTtZQUNmLGVBQWUsRUFBRSxHQUFHLENBQUMsT0FBTyxDQUFDLGVBQWUsQ0FBQyxlQUFlO1NBQzdELENBQUMsQ0FBQztRQUVILGNBQWMsQ0FBQyxTQUFTLENBQUM7WUFDdkIsSUFBSSxFQUFFLE9BQU87WUFDYixzQkFBc0IsRUFBRTtnQkFDdEIsWUFBWSxFQUFFLFVBQVUsQ0FBQyxZQUFZO2FBQ3RDO1NBQ0YsQ0FBQyxDQUFDO1FBRUgsTUFBTSxTQUFTLEdBQUcsY0FBYyxDQUFDLFlBQVksQ0FBQyxVQUFVLEVBQUU7WUFDeEQsS0FBSyxFQUFFLEdBQUcsQ0FBQyxjQUFjLENBQUMsWUFBWSxDQUFDLDBCQUEwQixDQUFDO1lBQ2xFLE9BQU8sRUFBRSxJQUFJLEdBQUcsQ0FBQyxZQUFZLENBQUM7Z0JBQzVCLFlBQVksRUFBRSxVQUFVO2dCQUN4QixZQUFZLEVBQUUsSUFBSSxDQUFDLGFBQWEsQ0FBQyxRQUFRO2FBQzFDLENBQUM7WUFDRixXQUFXLEVBQUU7Z0JBQ1gsWUFBWSxFQUFFLFVBQVU7Z0JBQ3hCLFVBQVUsRUFBRSxRQUFRO2FBQ3JCO1lBQ0QsWUFBWSxFQUFFLENBQUM7b0JBQ2IsYUFBYSxFQUFFLElBQUk7b0JBQ25CLFFBQVEsRUFBRSxHQUFHLENBQUMsUUFBUSxDQUFDLEdBQUc7aUJBQzNCLENBQUM7U0FDSCxDQUFDLENBQUM7UUFFSCxTQUFTLENBQUMsY0FBYyxDQUFDO1lBQ3ZCLFlBQVksRUFBRSxPQUFPO1lBQ3JCLGFBQWEsRUFBRSxRQUFRO1lBQ3ZCLFFBQVEsRUFBRSxLQUFLO1NBQ2hCLENBQUMsQ0FBQztRQUVILE1BQU0sT0FBTyxHQUFHLElBQUksR0FBRyxDQUFDLGNBQWMsQ0FBQyxJQUFJLEVBQUUsaUJBQWlCLEVBQUU7WUFDOUQsT0FBTyxFQUFFLElBQUksQ0FBQyxPQUFPO1lBQ3JCLGNBQWM7WUFDZCxZQUFZLEVBQUUsS0FBSyxDQUFDLFlBQVksRUFBRSxRQUFRLEVBQUUsWUFBWSxJQUFJLENBQUM7WUFDN0QsV0FBVyxFQUFFLG1CQUFtQjtTQUNqQyxDQUFDLENBQUM7UUFFSCxPQUFPLE9BQU8sQ0FBQztJQUNqQixDQUFDO0lBRU8sa0JBQWtCLENBQUMsS0FBK0I7UUFDeEQsTUFBTSxjQUFjLEdBQUcsSUFBSSxHQUFHLENBQUMscUJBQXFCLENBQUMsSUFBSSxFQUFFLFdBQVcsRUFBRTtZQUN0RSxHQUFHLEVBQUUsS0FBSyxDQUFDLFlBQVksRUFBRSxLQUFLLEVBQUUsR0FBRyxJQUFJLEdBQUc7WUFDMUMsY0FBYyxFQUFFLEtBQUssQ0FBQyxZQUFZLEVBQUUsS0FBSyxFQUFFLE1BQU0sSUFBSSxJQUFJO1NBQzFELENBQUMsQ0FBQztRQUVILGNBQWMsQ0FBQyxZQUFZLENBQUMsT0FBTyxFQUFFO1lBQ25DLEtBQUssRUFBRSxHQUFHLENBQUMsY0FBYyxDQUFDLFlBQVksQ0FBQyx1QkFBdUIsQ0FBQztZQUMvRCxPQUFPLEVBQUUsSUFBSSxHQUFHLENBQUMsWUFBWSxDQUFDO2dCQUM1QixZQUFZLEVBQUUsT0FBTztnQkFDckIsWUFBWSxFQUFFLElBQUksQ0FBQyxhQUFhLENBQUMsUUFBUTthQUMxQyxDQUFDO1lBQ0YsV0FBVyxFQUFFO2dCQUNYLFlBQVksRUFBRSxPQUFPO2FBQ3RCO1lBQ0QsWUFBWSxFQUFFLENBQUM7b0JBQ2IsYUFBYSxFQUFFLElBQUk7b0JBQ25CLFFBQVEsRUFBRSxHQUFHLENBQUMsUUFBUSxDQUFDLEdBQUc7aUJBQzNCLENBQUM7U0FDSCxDQUFDLENBQUM7UUFFSCxNQUFNLE9BQU8sR0FBRyxJQUFJLEdBQUcsQ0FBQyxjQUFjLENBQUMsSUFBSSxFQUFFLGNBQWMsRUFBRTtZQUMzRCxPQUFPLEVBQUUsSUFBSSxDQUFDLE9BQU87WUFDckIsY0FBYztZQUNkLFlBQVksRUFBRSxLQUFLLENBQUMsWUFBWSxFQUFFLEtBQUssRUFBRSxZQUFZLElBQUksQ0FBQztZQUMxRCxXQUFXLEVBQUUsZ0JBQWdCO1NBQzlCLENBQUMsQ0FBQztRQUVILE9BQU8sT0FBTyxDQUFDO0lBQ2pCLENBQUM7SUFFTyxrQkFBa0IsQ0FBQyxLQUErQjtRQUN4RCxtQ0FBbUM7UUFDbkMsTUFBTSxvQkFBb0IsR0FBRyxJQUFJLEdBQUcsQ0FBQyxhQUFhLENBQUMsSUFBSSxFQUFFLHNCQUFzQixFQUFFO1lBQy9FLEdBQUcsRUFBRSxJQUFJLENBQUMsR0FBRztZQUNiLFdBQVcsRUFBRSwyQ0FBMkM7WUFDeEQsZ0JBQWdCLEVBQUUsSUFBSTtTQUN2QixDQUFDLENBQUM7UUFFSCx5Q0FBeUM7UUFDekMsb0JBQW9CLENBQUMsY0FBYyxDQUNqQyxHQUFHLENBQUMsSUFBSSxDQUFDLElBQUksQ0FBQyxJQUFJLENBQUMsR0FBRyxDQUFDLFlBQVksQ0FBQyxFQUNwQyxHQUFHLENBQUMsSUFBSSxDQUFDLFVBQVUsRUFBRSxFQUNyQiw0QkFBNEIsQ0FDN0IsQ0FBQztRQUVGLDZCQUE2QjtRQUM3QixNQUFNLFdBQVcsR0FBRyxJQUFJLEdBQUcsQ0FBQyxJQUFJLENBQUMsSUFBSSxFQUFFLGlCQUFpQixFQUFFO1lBQ3hELFNBQVMsRUFBRSxJQUFJLEdBQUcsQ0FBQyxnQkFBZ0IsQ0FBQyxtQkFBbUIsQ0FBQztZQUN4RCxlQUFlLEVBQUU7Z0JBQ2YsR0FBRyxDQUFDLGFBQWEsQ0FBQyx3QkFBd0IsQ0FBQyw4QkFBOEIsQ0FBQzthQUMzRTtTQUNGLENBQUMsQ0FBQztRQUVILHlCQUF5QjtRQUN6QixXQUFXLENBQUMsV0FBVyxDQUFDLElBQUksR0FBRyxDQUFDLGVBQWUsQ0FBQztZQUM5QyxNQUFNLEVBQUUsR0FBRyxDQUFDLE1BQU0sQ0FBQyxLQUFLO1lBQ3hCLE9BQU8sRUFBRTtnQkFDUCwyQkFBMkI7Z0JBQzNCLGlDQUFpQztnQkFDakMsNEJBQTRCO2dCQUM1QixtQkFBbUI7YUFDcEI7WUFDRCxTQUFTLEVBQUUsQ0FBQyxHQUFHLENBQUM7U0FDakIsQ0FBQyxDQUFDLENBQUM7UUFFSiw4QkFBOEI7UUFDOUIsTUFBTSxRQUFRLEdBQUcsR0FBRyxDQUFDLFFBQVEsQ0FBQyxRQUFRLEVBQUUsQ0FBQztRQUN6QyxRQUFRLENBQUMsV0FBVyxDQUNsQixlQUFlLEVBQ2YsdUJBQXVCLEVBQ3ZCLHNCQUFzQixFQUN0QiwrQkFBK0I7UUFFL0IsMkNBQTJDO1FBQzNDLHNDQUFzQyxFQUN0QyxHQUFHLEVBQ0gsOEJBQThCLEVBQzlCLGlCQUFpQixFQUNqQix3QkFBd0IsRUFDeEIscUJBQXFCLEVBQ3JCLE1BQU0sRUFDTix3QkFBd0IsRUFDeEIsaUJBQWlCLEVBQ2pCLHlCQUF5QixFQUN6QixzQkFBc0IsRUFDdEIscUJBQXFCLEVBQ3JCLE9BQU8sRUFDUCxLQUFLLEVBQ0wsR0FBRyxFQUNILEtBQUssRUFFTCwwQkFBMEIsQ0FDM0IsQ0FBQztRQUVGLG9DQUFvQztRQUNwQyxNQUFNLGNBQWMsR0FBRyxJQUFJLEdBQUcsQ0FBQyxjQUFjLENBQUMsSUFBSSxFQUFFLHVCQUF1QixFQUFFO1lBQzNFLFlBQVksRUFBRSxLQUFLLENBQUMsbUJBQW1CLElBQUksR0FBRyxDQUFDLFlBQVksQ0FBQyxFQUFFLENBQUMsR0FBRyxDQUFDLGFBQWEsQ0FBQyxFQUFFLEVBQUUsR0FBRyxDQUFDLFlBQVksQ0FBQyxPQUFPLENBQUM7WUFDOUcsWUFBWSxFQUFFLEdBQUcsQ0FBQyxZQUFZLENBQUMsa0JBQWtCLEVBQUU7WUFDbkQsUUFBUTtZQUNSLElBQUksRUFBRSxXQUFXO1lBQ2pCLGFBQWEsRUFBRSxvQkFBb0I7WUFDbkMsWUFBWSxFQUFFLENBQUM7b0JBQ2IsVUFBVSxFQUFFLFdBQVc7b0JBQ3ZCLE1BQU0sRUFBRSxHQUFHLENBQUMsaUJBQWlCLENBQUMsR0FBRyxDQUFDLEdBQUcsRUFBRTt3QkFDckMsVUFBVSxFQUFFLEdBQUcsQ0FBQyxtQkFBbUIsQ0FBQyxHQUFHO3dCQUN2QyxTQUFTLEVBQUUsSUFBSTtxQkFDaEIsQ0FBQztpQkFDSCxDQUFDO1NBQ0gsQ0FBQyxDQUFDO1FBRUgsdUNBQXVDO1FBQ3ZDLE1BQU0sR0FBRyxHQUFHLElBQUksV0FBVyxDQUFDLGdCQUFnQixDQUFDLElBQUksRUFBRSxZQUFZLEVBQUU7WUFDL0QsR0FBRyxFQUFFLElBQUksQ0FBQyxHQUFHO1lBQ2IsY0FBYztZQUNkLFdBQVcsRUFBRSxLQUFLLENBQUMsZUFBZSxJQUFJLENBQUM7WUFDdkMsV0FBVyxFQUFFLEtBQUssQ0FBQyxlQUFlLElBQUksRUFBRTtZQUN4QyxXQUFXLEVBQUUsV0FBVyxDQUFDLFdBQVcsQ0FBQyxHQUFHLENBQUM7Z0JBQ3ZDLEtBQUssRUFBRSxHQUFHLENBQUMsUUFBUSxDQUFDLE9BQU8sQ0FBQyxDQUFDLENBQUM7YUFDL0IsQ0FBQztTQUNILENBQUMsQ0FBQztRQUVILDRCQUE0QjtRQUM1QixHQUFHLENBQUMscUJBQXFCLENBQUMsWUFBWSxFQUFFO1lBQ3RDLHdCQUF3QixFQUFFLEVBQUU7WUFDNUIsUUFBUSxFQUFFLEdBQUcsQ0FBQyxRQUFRLENBQUMsT0FBTyxDQUFDLENBQUMsQ0FBQztTQUNsQyxDQUFDLENBQUM7UUFFSCxPQUFPLEdBQUcsQ0FBQztJQUNiLENBQUM7SUFFTyxlQUFlLENBQ3JCLFlBQWdDLEVBQ2hDLFFBQTRCLEVBQzVCLEtBQXlCO1FBRXpCLE1BQU0sU0FBUyxHQUFHLElBQUksVUFBVSxDQUFDLFNBQVMsQ0FBQyxJQUFJLEVBQUUsbUJBQW1CLEVBQUU7WUFDcEUsYUFBYSxFQUFFLHFCQUFxQjtTQUNyQyxDQUFDLENBQUM7UUFFSCx1QkFBdUI7UUFDdkIsU0FBUyxDQUFDLFVBQVUsQ0FDbEIsSUFBSSxVQUFVLENBQUMsV0FBVyxDQUFDO1lBQ3pCLEtBQUssRUFBRSw4QkFBOEI7WUFDckMsSUFBSSxFQUFFLENBQUMsWUFBWSxDQUFDLG9CQUFvQixFQUFFLENBQUM7U0FDNUMsQ0FBQyxFQUNGLElBQUksVUFBVSxDQUFDLFdBQVcsQ0FBQztZQUN6QixLQUFLLEVBQUUsaUNBQWlDO1lBQ3hDLElBQUksRUFBRSxDQUFDLFlBQVksQ0FBQyx1QkFBdUIsRUFBRSxDQUFDO1NBQy9DLENBQUMsQ0FDSCxDQUFDO1FBRUYsbUJBQW1CO1FBQ25CLFNBQVMsQ0FBQyxVQUFVLENBQ2xCLElBQUksVUFBVSxDQUFDLFdBQVcsQ0FBQztZQUN6QixLQUFLLEVBQUUsMEJBQTBCO1lBQ2pDLElBQUksRUFBRSxDQUFDLFFBQVEsQ0FBQyxvQkFBb0IsRUFBRSxDQUFDO1NBQ3hDLENBQUMsRUFDRixJQUFJLFVBQVUsQ0FBQyxXQUFXLENBQUM7WUFDekIsS0FBSyxFQUFFLDZCQUE2QjtZQUNwQyxJQUFJLEVBQUUsQ0FBQyxRQUFRLENBQUMsdUJBQXVCLEVBQUUsQ0FBQztTQUMzQyxDQUFDLENBQ0gsQ0FBQztRQUVGLGdCQUFnQjtRQUNoQixTQUFTLENBQUMsVUFBVSxDQUNsQixJQUFJLFVBQVUsQ0FBQyxXQUFXLENBQUM7WUFDekIsS0FBSyxFQUFFLHVCQUF1QjtZQUM5QixJQUFJLEVBQUUsQ0FBQyxLQUFLLENBQUMsb0JBQW9CLEVBQUUsQ0FBQztTQUNyQyxDQUFDLEVBQ0YsSUFBSSxVQUFVLENBQUMsV0FBVyxDQUFDO1lBQ3pCLEtBQUssRUFBRSwwQkFBMEI7WUFDakMsSUFBSSxFQUFFLENBQUMsS0FBSyxDQUFDLHVCQUF1QixFQUFFLENBQUM7U0FDeEMsQ0FBQyxDQUNILENBQUM7UUFFRixPQUFPLFNBQVMsQ0FBQztJQUNuQixDQUFDO0NBQ0Y7QUEzVUQsa0RBMlVDIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0ICogYXMgY2RrIGZyb20gJ2F3cy1jZGstbGliJztcbmltcG9ydCAqIGFzIGVjMiBmcm9tICdhd3MtY2RrLWxpYi9hd3MtZWMyJztcbmltcG9ydCAqIGFzIGVjcyBmcm9tICdhd3MtY2RrLWxpYi9hd3MtZWNzJztcbmltcG9ydCAqIGFzIGlhbSBmcm9tICdhd3MtY2RrLWxpYi9hd3MtaWFtJztcbmltcG9ydCAqIGFzIGF1dG9zY2FsaW5nIGZyb20gJ2F3cy1jZGstbGliL2F3cy1hdXRvc2NhbGluZyc7XG5pbXBvcnQgKiBhcyBlbGJ2MiBmcm9tICdhd3MtY2RrLWxpYi9hd3MtZWxhc3RpY2xvYWRiYWxhbmNpbmd2Mic7XG5pbXBvcnQgKiBhcyBjbG91ZHdhdGNoIGZyb20gJ2F3cy1jZGstbGliL2F3cy1jbG91ZHdhdGNoJztcbmltcG9ydCAqIGFzIGxvZ3MgZnJvbSAnYXdzLWNkay1saWIvYXdzLWxvZ3MnO1xuaW1wb3J0IHsgQ29uc3RydWN0IH0gZnJvbSAnY29uc3RydWN0cyc7XG5cbmV4cG9ydCBpbnRlcmZhY2UgT2JzaWRpYW5IeWJyaWRTdGFja1Byb3BzIGV4dGVuZHMgY2RrLlN0YWNrUHJvcHMge1xuICBtaW5Db21wdXRlTm9kZXM/OiBudW1iZXI7XG4gIG1heENvbXB1dGVOb2Rlcz86IG51bWJlcjtcbiAgY29tcHV0ZUluc3RhbmNlVHlwZT86IGVjMi5JbnN0YW5jZVR5cGU7XG4gIHZwY0lkPzogc3RyaW5nO1xuICBjb250cm9sUGxhbmU/OiB7XG4gICAgb3JjaGVzdHJhdG9yOiB7XG4gICAgICBkZXNpcmVkQ291bnQ6IG51bWJlcjtcbiAgICAgIGNwdTogbnVtYmVyO1xuICAgICAgbWVtb3J5OiBudW1iZXI7XG4gICAgfTtcbiAgICByZWdpc3RyeToge1xuICAgICAgZGVzaXJlZENvdW50OiBudW1iZXI7XG4gICAgICBjcHU6IG51bWJlcjtcbiAgICAgIG1lbW9yeTogbnVtYmVyO1xuICAgIH07XG4gICAgcHJveHk6IHtcbiAgICAgIGRlc2lyZWRDb3VudDogbnVtYmVyO1xuICAgICAgY3B1OiBudW1iZXI7XG4gICAgICBtZW1vcnk6IG51bWJlcjtcbiAgICB9O1xuICB9O1xufVxuXG5leHBvcnQgY2xhc3MgT2JzaWRpYW5IeWJyaWRTdGFjayBleHRlbmRzIGNkay5TdGFjayB7XG4gIHB1YmxpYyByZWFkb25seSB2cGM6IGVjMi5JVnBjO1xuICBwdWJsaWMgcmVhZG9ubHkgY2x1c3RlcjogZWNzLkNsdXN0ZXI7XG4gIHB1YmxpYyByZWFkb25seSBhbGI6IGVsYnYyLkFwcGxpY2F0aW9uTG9hZEJhbGFuY2VyO1xuICBcbiAgY29uc3RydWN0b3Ioc2NvcGU6IENvbnN0cnVjdCwgaWQ6IHN0cmluZywgcHJvcHM6IE9ic2lkaWFuSHlicmlkU3RhY2tQcm9wcyA9IHt9KSB7XG4gICAgc3VwZXIoc2NvcGUsIGlkLCBwcm9wcyk7XG5cbiAgICAvLyBWUENcbiAgICB0aGlzLnZwYyA9IHByb3BzLnZwY0lkIFxuICAgICAgPyBlYzIuVnBjLmZyb21Mb29rdXAodGhpcywgJ1ZQQycsIHsgdnBjSWQ6IHByb3BzLnZwY0lkIH0pXG4gICAgICA6IG5ldyBlYzIuVnBjKHRoaXMsICdPYnNpZGlhblZQQycsIHtcbiAgICAgICAgICBtYXhBenM6IDIsXG4gICAgICAgICAgbmF0R2F0ZXdheXM6IDIsXG4gICAgICAgIH0pO1xuXG4gICAgLy8gRUNTIENsdXN0ZXIgZm9yIGNvbnRyb2wgcGxhbmVcbiAgICB0aGlzLmNsdXN0ZXIgPSBuZXcgZWNzLkNsdXN0ZXIodGhpcywgJ09ic2lkaWFuQ29udHJvbFBsYW5lJywge1xuICAgICAgdnBjOiB0aGlzLnZwYyxcbiAgICAgIGNvbnRhaW5lckluc2lnaHRzOiB0cnVlLFxuICAgIH0pO1xuXG4gICAgLy8gQXBwbGljYXRpb24gTG9hZCBCYWxhbmNlclxuICAgIHRoaXMuYWxiID0gbmV3IGVsYnYyLkFwcGxpY2F0aW9uTG9hZEJhbGFuY2VyKHRoaXMsICdPYnNpZGlhbkFMQicsIHtcbiAgICAgIHZwYzogdGhpcy52cGMsXG4gICAgICBpbnRlcm5ldEZhY2luZzogdHJ1ZSxcbiAgICB9KTtcblxuICAgIC8vIERlcGxveSBjb250cm9sIHBsYW5lIGNvbXBvbmVudHNcbiAgICBjb25zdCBvcmNoZXN0cmF0b3JTZXJ2aWNlID0gdGhpcy5jcmVhdGVPcmNoZXN0cmF0b3JTZXJ2aWNlKHByb3BzKTtcbiAgICBjb25zdCByZWdpc3RyeVNlcnZpY2UgPSB0aGlzLmNyZWF0ZVJlZ2lzdHJ5U2VydmljZShwcm9wcyk7XG4gICAgY29uc3QgcHJveHlTZXJ2aWNlID0gdGhpcy5jcmVhdGVQcm94eVNlcnZpY2UocHJvcHMpO1xuXG4gICAgLy8gRGVwbG95IGNvbXB1dGUgbm9kZXNcbiAgICBjb25zdCBjb21wdXRlTm9kZXMgPSB0aGlzLmNyZWF0ZUNvbXB1dGVOb2Rlcyhwcm9wcyk7XG5cbiAgICAvLyBDcmVhdGUgQ2xvdWRXYXRjaCBEYXNoYm9hcmRcbiAgICB0aGlzLmNyZWF0ZURhc2hib2FyZChvcmNoZXN0cmF0b3JTZXJ2aWNlLCByZWdpc3RyeVNlcnZpY2UsIHByb3h5U2VydmljZSk7XG5cbiAgICAvLyBPdXRwdXRzXG4gICAgbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ09ic2lkaWFuRW5kcG9pbnQnLCB7XG4gICAgICB2YWx1ZTogdGhpcy5hbGIubG9hZEJhbGFuY2VyRG5zTmFtZSxcbiAgICAgIGRlc2NyaXB0aW9uOiAnT2JzaWRpYW4gR2F0ZXdheSBlbmRwb2ludCcsXG4gICAgfSk7XG5cbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnQ2x1c3Rlck5hbWUnLCB7XG4gICAgICB2YWx1ZTogdGhpcy5jbHVzdGVyLmNsdXN0ZXJOYW1lLFxuICAgICAgZGVzY3JpcHRpb246ICdFQ1MgQ2x1c3RlciBuYW1lJyxcbiAgICB9KTtcbiAgfVxuXG4gIHByaXZhdGUgY3JlYXRlT3JjaGVzdHJhdG9yU2VydmljZShwcm9wczogT2JzaWRpYW5IeWJyaWRTdGFja1Byb3BzKTogZWNzLkZhcmdhdGVTZXJ2aWNlIHtcbiAgICBjb25zdCB0YXNrRGVmaW5pdGlvbiA9IG5ldyBlY3MuRmFyZ2F0ZVRhc2tEZWZpbml0aW9uKHRoaXMsICdPcmNoZXN0cmF0b3JUYXNrJywge1xuICAgICAgY3B1OiBwcm9wcy5jb250cm9sUGxhbmU/Lm9yY2hlc3RyYXRvcj8uY3B1IHx8IDIwNDgsXG4gICAgICBtZW1vcnlMaW1pdE1pQjogcHJvcHMuY29udHJvbFBsYW5lPy5vcmNoZXN0cmF0b3I/Lm1lbW9yeSB8fCA0MDk2LFxuICAgIH0pO1xuXG4gICAgdGFza0RlZmluaXRpb24uYWRkQ29udGFpbmVyKCdvcmNoZXN0cmF0b3InLCB7XG4gICAgICBpbWFnZTogZWNzLkNvbnRhaW5lckltYWdlLmZyb21SZWdpc3RyeSgnb2JzaWRpYW4vb3JjaGVzdHJhdG9yOmxhdGVzdCcpLFxuICAgICAgbG9nZ2luZzogbmV3IGVjcy5Bd3NMb2dEcml2ZXIoe1xuICAgICAgICBzdHJlYW1QcmVmaXg6ICdvcmNoZXN0cmF0b3InLFxuICAgICAgICBsb2dSZXRlbnRpb246IGxvZ3MuUmV0ZW50aW9uRGF5cy5PTkVfV0VFSyxcbiAgICAgIH0pLFxuICAgICAgZW52aXJvbm1lbnQ6IHtcbiAgICAgICAgU0VSVklDRV9OQU1FOiAnb3JjaGVzdHJhdG9yJyxcbiAgICAgIH0sXG4gICAgICBwb3J0TWFwcGluZ3M6IFt7XG4gICAgICAgIGNvbnRhaW5lclBvcnQ6IDkwOTAsXG4gICAgICAgIHByb3RvY29sOiBlY3MuUHJvdG9jb2wuVENQLFxuICAgICAgfV0sXG4gICAgfSk7XG5cbiAgICBjb25zdCBzZXJ2aWNlID0gbmV3IGVjcy5GYXJnYXRlU2VydmljZSh0aGlzLCAnT3JjaGVzdHJhdG9yU2VydmljZScsIHtcbiAgICAgIGNsdXN0ZXI6IHRoaXMuY2x1c3RlcixcbiAgICAgIHRhc2tEZWZpbml0aW9uLFxuICAgICAgZGVzaXJlZENvdW50OiBwcm9wcy5jb250cm9sUGxhbmU/Lm9yY2hlc3RyYXRvcj8uZGVzaXJlZENvdW50IHx8IDIsXG4gICAgICBzZXJ2aWNlTmFtZTogJ29ic2lkaWFuLW9yY2hlc3RyYXRvcicsXG4gICAgfSk7XG5cbiAgICAvLyBBZGQgQUxCIHRhcmdldFxuICAgIGNvbnN0IHRhcmdldEdyb3VwID0gbmV3IGVsYnYyLkFwcGxpY2F0aW9uVGFyZ2V0R3JvdXAodGhpcywgJ09yY2hlc3RyYXRvclRhcmdldEdyb3VwJywge1xuICAgICAgdnBjOiB0aGlzLnZwYyxcbiAgICAgIHBvcnQ6IDkwOTAsXG4gICAgICBwcm90b2NvbDogZWxidjIuQXBwbGljYXRpb25Qcm90b2NvbC5IVFRQLFxuICAgICAgdGFyZ2V0VHlwZTogZWxidjIuVGFyZ2V0VHlwZS5JUCxcbiAgICAgIGhlYWx0aENoZWNrOiB7XG4gICAgICAgIHBhdGg6ICcvaGVhbHRoJyxcbiAgICAgICAgaW50ZXJ2YWw6IGNkay5EdXJhdGlvbi5zZWNvbmRzKDMwKSxcbiAgICAgIH0sXG4gICAgfSk7XG5cbiAgICBzZXJ2aWNlLmF0dGFjaFRvQXBwbGljYXRpb25UYXJnZXRHcm91cCh0YXJnZXRHcm91cCk7XG5cbiAgICB0aGlzLmFsYi5hZGRMaXN0ZW5lcignT3JjaGVzdHJhdG9yTGlzdGVuZXInLCB7XG4gICAgICBwb3J0OiA5MDkwLFxuICAgICAgZGVmYXVsdFRhcmdldEdyb3VwczogW3RhcmdldEdyb3VwXSxcbiAgICB9KTtcblxuICAgIHJldHVybiBzZXJ2aWNlO1xuICB9XG5cbiAgcHJpdmF0ZSBjcmVhdGVSZWdpc3RyeVNlcnZpY2UocHJvcHM6IE9ic2lkaWFuSHlicmlkU3RhY2tQcm9wcyk6IGVjcy5GYXJnYXRlU2VydmljZSB7XG4gICAgY29uc3QgdGFza0RlZmluaXRpb24gPSBuZXcgZWNzLkZhcmdhdGVUYXNrRGVmaW5pdGlvbih0aGlzLCAnUmVnaXN0cnlUYXNrJywge1xuICAgICAgY3B1OiBwcm9wcy5jb250cm9sUGxhbmU/LnJlZ2lzdHJ5Py5jcHUgfHwgMTAyNCxcbiAgICAgIG1lbW9yeUxpbWl0TWlCOiBwcm9wcy5jb250cm9sUGxhbmU/LnJlZ2lzdHJ5Py5tZW1vcnkgfHwgMjA0OCxcbiAgICB9KTtcblxuICAgIC8vIEFkZCBFRlMgdm9sdW1lIGZvciBpbWFnZSBjYWNoZVxuICAgIGNvbnN0IGZpbGVTeXN0ZW0gPSBuZXcgY2RrLmF3c19lZnMuRmlsZVN5c3RlbSh0aGlzLCAnUmVnaXN0cnlDYWNoZScsIHtcbiAgICAgIHZwYzogdGhpcy52cGMsXG4gICAgICBlbmNyeXB0ZWQ6IHRydWUsXG4gICAgICBwZXJmb3JtYW5jZU1vZGU6IGNkay5hd3NfZWZzLlBlcmZvcm1hbmNlTW9kZS5HRU5FUkFMX1BVUlBPU0UsXG4gICAgfSk7XG5cbiAgICB0YXNrRGVmaW5pdGlvbi5hZGRWb2x1bWUoe1xuICAgICAgbmFtZTogJ2NhY2hlJyxcbiAgICAgIGVmc1ZvbHVtZUNvbmZpZ3VyYXRpb246IHtcbiAgICAgICAgZmlsZVN5c3RlbUlkOiBmaWxlU3lzdGVtLmZpbGVTeXN0ZW1JZCxcbiAgICAgIH0sXG4gICAgfSk7XG5cbiAgICBjb25zdCBjb250YWluZXIgPSB0YXNrRGVmaW5pdGlvbi5hZGRDb250YWluZXIoJ3JlZ2lzdHJ5Jywge1xuICAgICAgaW1hZ2U6IGVjcy5Db250YWluZXJJbWFnZS5mcm9tUmVnaXN0cnkoJ29ic2lkaWFuL3JlZ2lzdHJ5OmxhdGVzdCcpLFxuICAgICAgbG9nZ2luZzogbmV3IGVjcy5Bd3NMb2dEcml2ZXIoe1xuICAgICAgICBzdHJlYW1QcmVmaXg6ICdyZWdpc3RyeScsXG4gICAgICAgIGxvZ1JldGVudGlvbjogbG9ncy5SZXRlbnRpb25EYXlzLk9ORV9XRUVLLFxuICAgICAgfSksXG4gICAgICBlbnZpcm9ubWVudDoge1xuICAgICAgICBTRVJWSUNFX05BTUU6ICdyZWdpc3RyeScsXG4gICAgICAgIENBQ0hFX1BBVEg6ICcvY2FjaGUnLFxuICAgICAgfSxcbiAgICAgIHBvcnRNYXBwaW5nczogW3tcbiAgICAgICAgY29udGFpbmVyUG9ydDogOTA5MSxcbiAgICAgICAgcHJvdG9jb2w6IGVjcy5Qcm90b2NvbC5UQ1AsXG4gICAgICB9XSxcbiAgICB9KTtcblxuICAgIGNvbnRhaW5lci5hZGRNb3VudFBvaW50cyh7XG4gICAgICBzb3VyY2VWb2x1bWU6ICdjYWNoZScsXG4gICAgICBjb250YWluZXJQYXRoOiAnL2NhY2hlJyxcbiAgICAgIHJlYWRPbmx5OiBmYWxzZSxcbiAgICB9KTtcblxuICAgIGNvbnN0IHNlcnZpY2UgPSBuZXcgZWNzLkZhcmdhdGVTZXJ2aWNlKHRoaXMsICdSZWdpc3RyeVNlcnZpY2UnLCB7XG4gICAgICBjbHVzdGVyOiB0aGlzLmNsdXN0ZXIsXG4gICAgICB0YXNrRGVmaW5pdGlvbixcbiAgICAgIGRlc2lyZWRDb3VudDogcHJvcHMuY29udHJvbFBsYW5lPy5yZWdpc3RyeT8uZGVzaXJlZENvdW50IHx8IDIsXG4gICAgICBzZXJ2aWNlTmFtZTogJ29ic2lkaWFuLXJlZ2lzdHJ5JyxcbiAgICB9KTtcblxuICAgIHJldHVybiBzZXJ2aWNlO1xuICB9XG5cbiAgcHJpdmF0ZSBjcmVhdGVQcm94eVNlcnZpY2UocHJvcHM6IE9ic2lkaWFuSHlicmlkU3RhY2tQcm9wcyk6IGVjcy5GYXJnYXRlU2VydmljZSB7XG4gICAgY29uc3QgdGFza0RlZmluaXRpb24gPSBuZXcgZWNzLkZhcmdhdGVUYXNrRGVmaW5pdGlvbih0aGlzLCAnUHJveHlUYXNrJywge1xuICAgICAgY3B1OiBwcm9wcy5jb250cm9sUGxhbmU/LnByb3h5Py5jcHUgfHwgNTEyLFxuICAgICAgbWVtb3J5TGltaXRNaUI6IHByb3BzLmNvbnRyb2xQbGFuZT8ucHJveHk/Lm1lbW9yeSB8fCAxMDI0LFxuICAgIH0pO1xuXG4gICAgdGFza0RlZmluaXRpb24uYWRkQ29udGFpbmVyKCdwcm94eScsIHtcbiAgICAgIGltYWdlOiBlY3MuQ29udGFpbmVySW1hZ2UuZnJvbVJlZ2lzdHJ5KCdvYnNpZGlhbi9wcm94eTpsYXRlc3QnKSxcbiAgICAgIGxvZ2dpbmc6IG5ldyBlY3MuQXdzTG9nRHJpdmVyKHtcbiAgICAgICAgc3RyZWFtUHJlZml4OiAncHJveHknLFxuICAgICAgICBsb2dSZXRlbnRpb246IGxvZ3MuUmV0ZW50aW9uRGF5cy5PTkVfV0VFSyxcbiAgICAgIH0pLFxuICAgICAgZW52aXJvbm1lbnQ6IHtcbiAgICAgICAgU0VSVklDRV9OQU1FOiAncHJveHknLFxuICAgICAgfSxcbiAgICAgIHBvcnRNYXBwaW5nczogW3tcbiAgICAgICAgY29udGFpbmVyUG9ydDogODA5MCxcbiAgICAgICAgcHJvdG9jb2w6IGVjcy5Qcm90b2NvbC5UQ1AsXG4gICAgICB9XSxcbiAgICB9KTtcblxuICAgIGNvbnN0IHNlcnZpY2UgPSBuZXcgZWNzLkZhcmdhdGVTZXJ2aWNlKHRoaXMsICdQcm94eVNlcnZpY2UnLCB7XG4gICAgICBjbHVzdGVyOiB0aGlzLmNsdXN0ZXIsXG4gICAgICB0YXNrRGVmaW5pdGlvbixcbiAgICAgIGRlc2lyZWRDb3VudDogcHJvcHMuY29udHJvbFBsYW5lPy5wcm94eT8uZGVzaXJlZENvdW50IHx8IDMsXG4gICAgICBzZXJ2aWNlTmFtZTogJ29ic2lkaWFuLXByb3h5JyxcbiAgICB9KTtcblxuICAgIHJldHVybiBzZXJ2aWNlO1xuICB9XG5cbiAgcHJpdmF0ZSBjcmVhdGVDb21wdXRlTm9kZXMocHJvcHM6IE9ic2lkaWFuSHlicmlkU3RhY2tQcm9wcyk6IGF1dG9zY2FsaW5nLkF1dG9TY2FsaW5nR3JvdXAge1xuICAgIC8vIFNlY3VyaXR5IEdyb3VwIGZvciBjb21wdXRlIG5vZGVzXG4gICAgY29uc3QgY29tcHV0ZVNlY3VyaXR5R3JvdXAgPSBuZXcgZWMyLlNlY3VyaXR5R3JvdXAodGhpcywgJ0NvbXB1dGVTZWN1cml0eUdyb3VwJywge1xuICAgICAgdnBjOiB0aGlzLnZwYyxcbiAgICAgIGRlc2NyaXB0aW9uOiAnU2VjdXJpdHkgZ3JvdXAgZm9yIE9ic2lkaWFuIGNvbXB1dGUgbm9kZXMnLFxuICAgICAgYWxsb3dBbGxPdXRib3VuZDogdHJ1ZSxcbiAgICB9KTtcblxuICAgIC8vIEFsbG93IGNvbW11bmljYXRpb24gZnJvbSBjb250cm9sIHBsYW5lXG4gICAgY29tcHV0ZVNlY3VyaXR5R3JvdXAuYWRkSW5ncmVzc1J1bGUoXG4gICAgICBlYzIuUGVlci5pcHY0KHRoaXMudnBjLnZwY0NpZHJCbG9jayksXG4gICAgICBlYzIuUG9ydC5hbGxUcmFmZmljKCksXG4gICAgICAnQWxsb3cgYWxsIHRyYWZmaWMgZnJvbSBWUEMnXG4gICAgKTtcblxuICAgIC8vIElBTSBSb2xlIGZvciBjb21wdXRlIG5vZGVzXG4gICAgY29uc3QgY29tcHV0ZVJvbGUgPSBuZXcgaWFtLlJvbGUodGhpcywgJ0NvbXB1dGVOb2RlUm9sZScsIHtcbiAgICAgIGFzc3VtZWRCeTogbmV3IGlhbS5TZXJ2aWNlUHJpbmNpcGFsKCdlYzIuYW1hem9uYXdzLmNvbScpLFxuICAgICAgbWFuYWdlZFBvbGljaWVzOiBbXG4gICAgICAgIGlhbS5NYW5hZ2VkUG9saWN5LmZyb21Bd3NNYW5hZ2VkUG9saWN5TmFtZSgnQW1hem9uU1NNTWFuYWdlZEluc3RhbmNlQ29yZScpLFxuICAgICAgXSxcbiAgICB9KTtcblxuICAgIC8vIEFkZCBEb2NrZXIgcGVybWlzc2lvbnNcbiAgICBjb21wdXRlUm9sZS5hZGRUb1BvbGljeShuZXcgaWFtLlBvbGljeVN0YXRlbWVudCh7XG4gICAgICBlZmZlY3Q6IGlhbS5FZmZlY3QuQUxMT1csXG4gICAgICBhY3Rpb25zOiBbXG4gICAgICAgICdlY3I6R2V0QXV0aG9yaXphdGlvblRva2VuJyxcbiAgICAgICAgJ2VjcjpCYXRjaENoZWNrTGF5ZXJBdmFpbGFiaWxpdHknLFxuICAgICAgICAnZWNyOkdldERvd25sb2FkVXJsRm9yTGF5ZXInLFxuICAgICAgICAnZWNyOkJhdGNoR2V0SW1hZ2UnLFxuICAgICAgXSxcbiAgICAgIHJlc291cmNlczogWycqJ10sXG4gICAgfSkpO1xuXG4gICAgLy8gVXNlciBEYXRhIGZvciBjb21wdXRlIG5vZGVzXG4gICAgY29uc3QgdXNlckRhdGEgPSBlYzIuVXNlckRhdGEuZm9yTGludXgoKTtcbiAgICB1c2VyRGF0YS5hZGRDb21tYW5kcyhcbiAgICAgICd5dW0gdXBkYXRlIC15JyxcbiAgICAgICd5dW0gaW5zdGFsbCAteSBkb2NrZXInLFxuICAgICAgJ3NlcnZpY2UgZG9ja2VyIHN0YXJ0JyxcbiAgICAgICd1c2VybW9kIC1hIC1HIGRvY2tlciBlYzItdXNlcicsXG4gICAgICBcbiAgICAgIC8vIENvbmZpZ3VyZSBEb2NrZXIgZm9yIGNvbnRhaW5lciBleGVjdXRpb25cbiAgICAgICdjYXQgPiAvZXRjL2RvY2tlci9kYWVtb24uanNvbiA8PCBFT0YnLFxuICAgICAgJ3snLFxuICAgICAgJyAgXCJsb2ctZHJpdmVyXCI6IFwianNvbi1maWxlXCIsJyxcbiAgICAgICcgIFwibG9nLW9wdHNcIjogeycsXG4gICAgICAnICAgIFwibWF4LXNpemVcIjogXCIxMG1cIiwnLFxuICAgICAgJyAgICBcIm1heC1maWxlXCI6IFwiM1wiJyxcbiAgICAgICcgIH0sJyxcbiAgICAgICcgIFwiZGVmYXVsdC11bGltaXRzXCI6IHsnLFxuICAgICAgJyAgICBcIm5vZmlsZVwiOiB7JyxcbiAgICAgICcgICAgICBcIk5hbWVcIjogXCJub2ZpbGVcIiwnLFxuICAgICAgJyAgICAgIFwiSGFyZFwiOiA2NTUzNiwnLFxuICAgICAgJyAgICAgIFwiU29mdFwiOiA2NTUzNicsXG4gICAgICAnICAgIH0nLFxuICAgICAgJyAgfScsXG4gICAgICAnfScsXG4gICAgICAnRU9GJyxcbiAgICAgIFxuICAgICAgJ3N5c3RlbWN0bCByZXN0YXJ0IGRvY2tlcidcbiAgICApO1xuXG4gICAgLy8gTGF1bmNoIFRlbXBsYXRlIGZvciBjb21wdXRlIG5vZGVzXG4gICAgY29uc3QgbGF1bmNoVGVtcGxhdGUgPSBuZXcgZWMyLkxhdW5jaFRlbXBsYXRlKHRoaXMsICdDb21wdXRlTGF1bmNoVGVtcGxhdGUnLCB7XG4gICAgICBpbnN0YW5jZVR5cGU6IHByb3BzLmNvbXB1dGVJbnN0YW5jZVR5cGUgfHwgZWMyLkluc3RhbmNlVHlwZS5vZihlYzIuSW5zdGFuY2VDbGFzcy5NNSwgZWMyLkluc3RhbmNlU2l6ZS5YTEFSR0UyKSxcbiAgICAgIG1hY2hpbmVJbWFnZTogZWMyLk1hY2hpbmVJbWFnZS5sYXRlc3RBbWF6b25MaW51eDIoKSxcbiAgICAgIHVzZXJEYXRhLFxuICAgICAgcm9sZTogY29tcHV0ZVJvbGUsXG4gICAgICBzZWN1cml0eUdyb3VwOiBjb21wdXRlU2VjdXJpdHlHcm91cCxcbiAgICAgIGJsb2NrRGV2aWNlczogW3tcbiAgICAgICAgZGV2aWNlTmFtZTogJy9kZXYveHZkYScsXG4gICAgICAgIHZvbHVtZTogZWMyLkJsb2NrRGV2aWNlVm9sdW1lLmVicygyMDAsIHtcbiAgICAgICAgICB2b2x1bWVUeXBlOiBlYzIuRWJzRGV2aWNlVm9sdW1lVHlwZS5HUDMsXG4gICAgICAgICAgZW5jcnlwdGVkOiB0cnVlLFxuICAgICAgICB9KSxcbiAgICAgIH1dLFxuICAgIH0pO1xuXG4gICAgLy8gQXV0byBTY2FsaW5nIEdyb3VwIGZvciBjb21wdXRlIG5vZGVzXG4gICAgY29uc3QgYXNnID0gbmV3IGF1dG9zY2FsaW5nLkF1dG9TY2FsaW5nR3JvdXAodGhpcywgJ0NvbXB1dGVBU0cnLCB7XG4gICAgICB2cGM6IHRoaXMudnBjLFxuICAgICAgbGF1bmNoVGVtcGxhdGUsXG4gICAgICBtaW5DYXBhY2l0eTogcHJvcHMubWluQ29tcHV0ZU5vZGVzIHx8IDMsXG4gICAgICBtYXhDYXBhY2l0eTogcHJvcHMubWF4Q29tcHV0ZU5vZGVzIHx8IDIwLFxuICAgICAgaGVhbHRoQ2hlY2s6IGF1dG9zY2FsaW5nLkhlYWx0aENoZWNrLmVjMih7XG4gICAgICAgIGdyYWNlOiBjZGsuRHVyYXRpb24ubWludXRlcyg1KSxcbiAgICAgIH0pLFxuICAgIH0pO1xuXG4gICAgLy8gQXV0by1zY2FsaW5nIGJhc2VkIG9uIENQVVxuICAgIGFzZy5zY2FsZU9uQ3B1VXRpbGl6YXRpb24oJ1NjYWxlT25DUFUnLCB7XG4gICAgICB0YXJnZXRVdGlsaXphdGlvblBlcmNlbnQ6IDcwLFxuICAgICAgY29vbGRvd246IGNkay5EdXJhdGlvbi5taW51dGVzKDUpLFxuICAgIH0pO1xuXG4gICAgcmV0dXJuIGFzZztcbiAgfVxuXG4gIHByaXZhdGUgY3JlYXRlRGFzaGJvYXJkKFxuICAgIG9yY2hlc3RyYXRvcjogZWNzLkZhcmdhdGVTZXJ2aWNlLFxuICAgIHJlZ2lzdHJ5OiBlY3MuRmFyZ2F0ZVNlcnZpY2UsXG4gICAgcHJveHk6IGVjcy5GYXJnYXRlU2VydmljZVxuICApOiBjbG91ZHdhdGNoLkRhc2hib2FyZCB7XG4gICAgY29uc3QgZGFzaGJvYXJkID0gbmV3IGNsb3Vkd2F0Y2guRGFzaGJvYXJkKHRoaXMsICdPYnNpZGlhbkRhc2hib2FyZCcsIHtcbiAgICAgIGRhc2hib2FyZE5hbWU6ICdvYnNpZGlhbi1wcm9kdWN0aW9uJyxcbiAgICB9KTtcblxuICAgIC8vIE9yY2hlc3RyYXRvciBtZXRyaWNzXG4gICAgZGFzaGJvYXJkLmFkZFdpZGdldHMoXG4gICAgICBuZXcgY2xvdWR3YXRjaC5HcmFwaFdpZGdldCh7XG4gICAgICAgIHRpdGxlOiAnT3JjaGVzdHJhdG9yIENQVSBVdGlsaXphdGlvbicsXG4gICAgICAgIGxlZnQ6IFtvcmNoZXN0cmF0b3IubWV0cmljQ3B1VXRpbGl6YXRpb24oKV0sXG4gICAgICB9KSxcbiAgICAgIG5ldyBjbG91ZHdhdGNoLkdyYXBoV2lkZ2V0KHtcbiAgICAgICAgdGl0bGU6ICdPcmNoZXN0cmF0b3IgTWVtb3J5IFV0aWxpemF0aW9uJyxcbiAgICAgICAgbGVmdDogW29yY2hlc3RyYXRvci5tZXRyaWNNZW1vcnlVdGlsaXphdGlvbigpXSxcbiAgICAgIH0pXG4gICAgKTtcblxuICAgIC8vIFJlZ2lzdHJ5IG1ldHJpY3NcbiAgICBkYXNoYm9hcmQuYWRkV2lkZ2V0cyhcbiAgICAgIG5ldyBjbG91ZHdhdGNoLkdyYXBoV2lkZ2V0KHtcbiAgICAgICAgdGl0bGU6ICdSZWdpc3RyeSBDUFUgVXRpbGl6YXRpb24nLFxuICAgICAgICBsZWZ0OiBbcmVnaXN0cnkubWV0cmljQ3B1VXRpbGl6YXRpb24oKV0sXG4gICAgICB9KSxcbiAgICAgIG5ldyBjbG91ZHdhdGNoLkdyYXBoV2lkZ2V0KHtcbiAgICAgICAgdGl0bGU6ICdSZWdpc3RyeSBNZW1vcnkgVXRpbGl6YXRpb24nLFxuICAgICAgICBsZWZ0OiBbcmVnaXN0cnkubWV0cmljTWVtb3J5VXRpbGl6YXRpb24oKV0sXG4gICAgICB9KVxuICAgICk7XG5cbiAgICAvLyBQcm94eSBtZXRyaWNzXG4gICAgZGFzaGJvYXJkLmFkZFdpZGdldHMoXG4gICAgICBuZXcgY2xvdWR3YXRjaC5HcmFwaFdpZGdldCh7XG4gICAgICAgIHRpdGxlOiAnUHJveHkgQ1BVIFV0aWxpemF0aW9uJyxcbiAgICAgICAgbGVmdDogW3Byb3h5Lm1ldHJpY0NwdVV0aWxpemF0aW9uKCldLFxuICAgICAgfSksXG4gICAgICBuZXcgY2xvdWR3YXRjaC5HcmFwaFdpZGdldCh7XG4gICAgICAgIHRpdGxlOiAnUHJveHkgTWVtb3J5IFV0aWxpemF0aW9uJyxcbiAgICAgICAgbGVmdDogW3Byb3h5Lm1ldHJpY01lbW9yeVV0aWxpemF0aW9uKCldLFxuICAgICAgfSlcbiAgICApO1xuXG4gICAgcmV0dXJuIGRhc2hib2FyZDtcbiAgfVxufSJdfQ==