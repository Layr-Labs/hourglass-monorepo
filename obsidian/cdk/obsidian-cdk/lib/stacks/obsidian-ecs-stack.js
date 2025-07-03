"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ObsidianECSStack = void 0;
const cdk = require("aws-cdk-lib");
const ec2 = require("aws-cdk-lib/aws-ec2");
const ecs = require("aws-cdk-lib/aws-ecs");
const iam = require("aws-cdk-lib/aws-iam");
const elbv2 = require("aws-cdk-lib/aws-elasticloadbalancingv2");
const logs = require("aws-cdk-lib/aws-logs");
class ObsidianECSStack extends cdk.Stack {
    constructor(scope, id, props = {}) {
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
exports.ObsidianECSStack = ObsidianECSStack;
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoib2JzaWRpYW4tZWNzLXN0YWNrLmpzIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsib2JzaWRpYW4tZWNzLXN0YWNrLnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiI7OztBQUFBLG1DQUFtQztBQUNuQywyQ0FBMkM7QUFDM0MsMkNBQTJDO0FBQzNDLDJDQUEyQztBQUMzQyxnRUFBZ0U7QUFDaEUsNkNBQTZDO0FBc0I3QyxNQUFhLGdCQUFpQixTQUFRLEdBQUcsQ0FBQyxLQUFLO0lBQzdDLFlBQVksS0FBZ0IsRUFBRSxFQUFVLEVBQUUsUUFBK0IsRUFBRTtRQUN6RSxLQUFLLENBQUMsS0FBSyxFQUFFLEVBQUUsRUFBRSxLQUFLLENBQUMsQ0FBQztRQUV4QixNQUFNO1FBQ04sTUFBTSxHQUFHLEdBQUcsS0FBSyxDQUFDLEtBQUs7WUFDckIsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxHQUFHLENBQUMsVUFBVSxDQUFDLElBQUksRUFBRSxLQUFLLEVBQUUsRUFBRSxLQUFLLEVBQUUsS0FBSyxDQUFDLEtBQUssRUFBRSxDQUFDO1lBQ3pELENBQUMsQ0FBQyxJQUFJLEdBQUcsQ0FBQyxHQUFHLENBQUMsSUFBSSxFQUFFLGFBQWEsRUFBRTtnQkFDL0IsTUFBTSxFQUFFLENBQUM7Z0JBQ1QsV0FBVyxFQUFFLENBQUM7YUFDZixDQUFDLENBQUM7UUFFUCxjQUFjO1FBQ2QsTUFBTSxPQUFPLEdBQUcsSUFBSSxHQUFHLENBQUMsT0FBTyxDQUFDLElBQUksRUFBRSxpQkFBaUIsRUFBRTtZQUN2RCxHQUFHO1lBQ0gsaUJBQWlCLEVBQUUsSUFBSTtTQUN4QixDQUFDLENBQUM7UUFFSCw0QkFBNEI7UUFDNUIsTUFBTSxHQUFHLEdBQUcsSUFBSSxLQUFLLENBQUMsdUJBQXVCLENBQUMsSUFBSSxFQUFFLGFBQWEsRUFBRTtZQUNqRSxHQUFHO1lBQ0gsY0FBYyxFQUFFLElBQUk7U0FDckIsQ0FBQyxDQUFDO1FBRUgsMERBQTBEO1FBQzFELE1BQU0sY0FBYyxHQUFHLElBQUksR0FBRyxD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QUFJLEdBQUcsQ0FBQyxTQUFTLENBQUMsSUFBSSxFQUFFLGFBQWEsRUFBRTtZQUNyQyxLQUFLLEVBQUUsT0FBTyxDQUFDLFdBQVc7WUFDMUIsV0FBVyxFQUFFLGtCQUFrQjtTQUNoQyxDQUFDLENBQUM7SUFDTCxDQUFDO0NBQ0Y7QUFwTkQsNENBb05DIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0ICogYXMgY2RrIGZyb20gJ2F3cy1jZGstbGliJztcbmltcG9ydCAqIGFzIGVjMiBmcm9tICdhd3MtY2RrLWxpYi9hd3MtZWMyJztcbmltcG9ydCAqIGFzIGVjcyBmcm9tICdhd3MtY2RrLWxpYi9hd3MtZWNzJztcbmltcG9ydCAqIGFzIGlhbSBmcm9tICdhd3MtY2RrLWxpYi9hd3MtaWFtJztcbmltcG9ydCAqIGFzIGVsYnYyIGZyb20gJ2F3cy1jZGstbGliL2F3cy1lbGFzdGljbG9hZGJhbGFuY2luZ3YyJztcbmltcG9ydCAqIGFzIGxvZ3MgZnJvbSAnYXdzLWNkay1saWIvYXdzLWxvZ3MnO1xuaW1wb3J0IHsgQ29uc3RydWN0IH0gZnJvbSAnY29uc3RydWN0cyc7XG5cbmV4cG9ydCBpbnRlcmZhY2UgT2JzaWRpYW5FQ1NTdGFja1Byb3BzIGV4dGVuZHMgY2RrLlN0YWNrUHJvcHMge1xuICB2cGNJZD86IHN0cmluZztcbiAgb3JjaGVzdHJhdG9yQ29uZmlnPzoge1xuICAgIGRlc2lyZWRDb3VudDogbnVtYmVyO1xuICAgIGNwdTogbnVtYmVyO1xuICAgIG1lbW9yeTogbnVtYmVyO1xuICB9O1xuICByZWdpc3RyeUNvbmZpZz86IHtcbiAgICBkZXNpcmVkQ291bnQ6IG51bWJlcjtcbiAgICBjcHU6IG51bWJlcjtcbiAgICBtZW1vcnk6IG51bWJlcjtcbiAgfTtcbiAgcHJveHlDb25maWc/OiB7XG4gICAgZGVzaXJlZENvdW50OiBudW1iZXI7XG4gICAgY3B1OiBudW1iZXI7XG4gICAgbWVtb3J5OiBudW1iZXI7XG4gIH07XG59XG5cbmV4cG9ydCBjbGFzcyBPYnNpZGlhbkVDU1N0YWNrIGV4dGVuZHMgY2RrLlN0YWNrIHtcbiAgY29uc3RydWN0b3Ioc2NvcGU6IENvbnN0cnVjdCwgaWQ6IHN0cmluZywgcHJvcHM6IE9ic2lkaWFuRUNTU3RhY2tQcm9wcyA9IHt9KSB7XG4gICAgc3VwZXIoc2NvcGUsIGlkLCBwcm9wcyk7XG5cbiAgICAvLyBWUENcbiAgICBjb25zdCB2cGMgPSBwcm9wcy52cGNJZCBcbiAgICAgID8gZWMyLlZwYy5mcm9tTG9va3VwKHRoaXMsICdWUEMnLCB7IHZwY0lkOiBwcm9wcy52cGNJZCB9KVxuICAgICAgOiBuZXcgZWMyLlZwYyh0aGlzLCAnT2JzaWRpYW5WUEMnLCB7XG4gICAgICAgICAgbWF4QXpzOiAyLFxuICAgICAgICAgIG5hdEdhdGV3YXlzOiAxLFxuICAgICAgICB9KTtcblxuICAgIC8vIEVDUyBDbHVzdGVyXG4gICAgY29uc3QgY2x1c3RlciA9IG5ldyBlY3MuQ2x1c3Rlcih0aGlzLCAnT2JzaWRpYW5DbHVzdGVyJywge1xuICAgICAgdnBjLFxuICAgICAgY29udGFpbmVySW5zaWdodHM6IHRydWUsXG4gICAgfSk7XG5cbiAgICAvLyBBcHBsaWNhdGlvbiBMb2FkIEJhbGFuY2VyXG4gICAgY29uc3QgYWxiID0gbmV3IGVsYnYyLkFwcGxpY2F0aW9uTG9hZEJhbGFuY2VyKHRoaXMsICdPYnNpZGlhbkFMQicsIHtcbiAgICAgIHZwYyxcbiAgICAgIGludGVybmV0RmFjaW5nOiB0cnVlLFxuICAgIH0pO1xuXG4gICAgLy8gR2F0ZXdheSBUYXNrIERlZmluaXRpb24gKGFsbCBzZXJ2aWNlcyBpbiBvbmUgY29udGFpbmVyKVxuICAgIGNvbnN0IGdhdGV3YXlUYXNrRGVmID0gbmV3IGVjcy5GYXJnYXRlVGFza0RlZmluaXRpb24odGhpcywgJ0dhdGV3YXlUYXNrJywge1xuICAgICAgY3B1OiA0MDk2LFxuICAgICAgbWVtb3J5TGltaXRNaUI6IDgxOTIsXG4gICAgfSk7XG5cbiAgICAvLyBJQU0gcGVybWlzc2lvbnMgZm9yIHRoZSB0YXNrXG4gICAgZ2F0ZXdheVRhc2tEZWYuYWRkVG9UYXNrUm9sZVBvbGljeShuZXcgaWFtLlBvbGljeVN0YXRlbWVudCh7XG4gICAgICBlZmZlY3Q6IGlhbS5FZmZlY3QuQUxMT1csXG4gICAgICBhY3Rpb25zOiBbXG4gICAgICAgICdlY3I6R2V0QXV0aG9yaXphdGlvblRva2VuJyxcbiAgICAgICAgJ2VjcjpCYXRjaENoZWNrTGF5ZXJBdmFpbGFiaWxpdHknLFxuICAgICAgICAnZWNyOkdldERvd25sb2FkVXJsRm9yTGF5ZXInLFxuICAgICAgICAnZWNyOkJhdGNoR2V0SW1hZ2UnLFxuICAgICAgXSxcbiAgICAgIHJlc291cmNlczogWycqJ10sXG4gICAgfSkpO1xuXG4gICAgZ2F0ZXdheVRhc2tEZWYuYWRkVG9UYXNrUm9sZVBvbGljeShuZXcgaWFtLlBvbGljeVN0YXRlbWVudCh7XG4gICAgICBlZmZlY3Q6IGlhbS5FZmZlY3QuQUxMT1csXG4gICAgICBhY3Rpb25zOiBbXG4gICAgICAgICdzZWNyZXRzbWFuYWdlcjpHZXRTZWNyZXRWYWx1ZScsXG4gICAgICBdLFxuICAgICAgcmVzb3VyY2VzOiBbYGFybjphd3M6c2VjcmV0c21hbmFnZXI6JHt0aGlzLnJlZ2lvbn06JHt0aGlzLmFjY291bnR9OnNlY3JldDpvYnNpZGlhbi8qYF0sXG4gICAgfSkpO1xuXG4gICAgLy8gR2F0ZXdheSBDb250YWluZXJcbiAgICBjb25zdCBjb250YWluZXIgPSBnYXRld2F5VGFza0RlZi5hZGRDb250YWluZXIoJ2dhdGV3YXknLCB7XG4gICAgICBpbWFnZTogZWNzLkNvbnRhaW5lckltYWdlLmZyb21SZWdpc3RyeSgnb2JzaWRpYW4vZ2F0ZXdheTpsYXRlc3QnKSxcbiAgICAgIGxvZ2dpbmc6IG5ldyBlY3MuQXdzTG9nRHJpdmVyKHtcbiAgICAgICAgc3RyZWFtUHJlZml4OiAnZ2F0ZXdheScsXG4gICAgICAgIGxvZ1JldGVudGlvbjogbG9ncy5SZXRlbnRpb25EYXlzLk9ORV9XRUVLLFxuICAgICAgfSksXG4gICAgICBlbnZpcm9ubWVudDoge1xuICAgICAgICBFTlZJUk9OTUVOVDogJ3N0YWdpbmcnLFxuICAgICAgICBBV1NfUkVHSU9OOiB0aGlzLnJlZ2lvbixcbiAgICAgIH0sXG4gICAgICBwb3J0TWFwcGluZ3M6IFtcbiAgICAgICAge1xuICAgICAgICAgIGNvbnRhaW5lclBvcnQ6IDgwODAsXG4gICAgICAgICAgcHJvdG9jb2w6IGVjcy5Qcm90b2NvbC5UQ1AsXG4gICAgICAgIH0sXG4gICAgICAgIHtcbiAgICAgICAgICBjb250YWluZXJQb3J0OiA5MDkwLFxuICAgICAgICAgIHByb3RvY29sOiBlY3MuUHJvdG9jb2wuVENQLFxuICAgICAgICB9LFxuICAgICAgICB7XG4gICAgICAgICAgY29udGFpbmVyUG9ydDogOTA5MSxcbiAgICAgICAgICBwcm90b2NvbDogZWNzLlByb3RvY29sLlRDUCxcbiAgICAgICAgfSxcbiAgICAgIF0sXG4gICAgICBoZWFsdGhDaGVjazoge1xuICAgICAgICBjb21tYW5kOiBbJ0NNRC1TSEVMTCcsICdjdXJsIC1mIGh0dHA6Ly9sb2NhbGhvc3Q6ODA4MC9oZWFsdGggfHwgZXhpdCAxJ10sXG4gICAgICAgIGludGVydmFsOiBjZGsuRHVyYXRpb24uc2Vjb25kcygzMCksXG4gICAgICAgIHRpbWVvdXQ6IGNkay5EdXJhdGlvbi5zZWNvbmRzKDUpLFxuICAgICAgICByZXRyaWVzOiAzLFxuICAgICAgfSxcbiAgICB9KTtcblxuICAgIC8vIEdhdGV3YXkgU2VydmljZVxuICAgIGNvbnN0IHNlcnZpY2UgPSBuZXcgZWNzLkZhcmdhdGVTZXJ2aWNlKHRoaXMsICdHYXRld2F5U2VydmljZScsIHtcbiAgICAgIGNsdXN0ZXIsXG4gICAgICB0YXNrRGVmaW5pdGlvbjogZ2F0ZXdheVRhc2tEZWYsXG4gICAgICBkZXNpcmVkQ291bnQ6IDIsXG4gICAgICBzZXJ2aWNlTmFtZTogJ29ic2lkaWFuLWdhdGV3YXknLFxuICAgICAgYXNzaWduUHVibGljSXA6IGZhbHNlLFxuICAgIH0pO1xuXG4gICAgLy8gQ29uZmlndXJlIEFMQiB0YXJnZXQgZ3JvdXBzIGFuZCBsaXN0ZW5lcnNcbiAgICBcbiAgICAvLyBIVFRQIHRhcmdldCBncm91cFxuICAgIGNvbnN0IGh0dHBUYXJnZXRHcm91cCA9IG5ldyBlbGJ2Mi5BcHBsaWNhdGlvblRhcmdldEdyb3VwKHRoaXMsICdIdHRwVGFyZ2V0R3JvdXAnLCB7XG4gICAgICB2cGMsXG4gICAgICBwb3J0OiA4MDgwLFxuICAgICAgcHJvdG9jb2w6IGVsYnYyLkFwcGxpY2F0aW9uUHJvdG9jb2wuSFRUUCxcbiAgICAgIHRhcmdldFR5cGU6IGVsYnYyLlRhcmdldFR5cGUuSVAsXG4gICAgICBoZWFsdGhDaGVjazoge1xuICAgICAgICBwYXRoOiAnL2hlYWx0aCcsXG4gICAgICAgIGludGVydmFsOiBjZGsuRHVyYXRpb24uc2Vjb25kcygzMCksXG4gICAgICAgIHRpbWVvdXQ6IGNkay5EdXJhdGlvbi5zZWNvbmRzKDUpLFxuICAgICAgICBoZWFsdGh5VGhyZXNob2xkQ291bnQ6IDIsXG4gICAgICAgIHVuaGVhbHRoeVRocmVzaG9sZENvdW50OiAzLFxuICAgICAgfSxcbiAgICB9KTtcblxuICAgIC8vIGdSUEMgdGFyZ2V0IGdyb3VwXG4gICAgY29uc3QgZ3JwY1RhcmdldEdyb3VwID0gbmV3IGVsYnYyLkFwcGxpY2F0aW9uVGFyZ2V0R3JvdXAodGhpcywgJ0dycGNUYXJnZXRHcm91cCcsIHtcbiAgICAgIHZwYyxcbiAgICAgIHBvcnQ6IDkwOTAsXG4gICAgICBwcm90b2NvbDogZWxidjIuQXBwbGljYXRpb25Qcm90b2NvbC5IVFRQLFxuICAgICAgdGFyZ2V0VHlwZTogZWxidjIuVGFyZ2V0VHlwZS5JUCxcbiAgICAgIHByb3RvY29sVmVyc2lvbjogZWxidjIuQXBwbGljYXRpb25Qcm90b2NvbFZlcnNpb24uR1JQQyxcbiAgICAgIGhlYWx0aENoZWNrOiB7XG4gICAgICAgIGVuYWJsZWQ6IHRydWUsXG4gICAgICAgIHByb3RvY29sOiBlbGJ2Mi5Qcm90b2NvbC5IVFRQLFxuICAgICAgICBwYXRoOiAnL2dycGMuaGVhbHRoLnYxLkhlYWx0aC9DaGVjaycsXG4gICAgICAgIGludGVydmFsOiBjZGsuRHVyYXRpb24uc2Vjb25kcygzMCksXG4gICAgICAgIHRpbWVvdXQ6IGNkay5EdXJhdGlvbi5zZWNvbmRzKDUpLFxuICAgICAgICBoZWFsdGh5VGhyZXNob2xkQ291bnQ6IDIsXG4gICAgICAgIHVuaGVhbHRoeVRocmVzaG9sZENvdW50OiAzLFxuICAgICAgfSxcbiAgICB9KTtcblxuICAgIC8vIE1ldHJpY3MgdGFyZ2V0IGdyb3VwXG4gICAgY29uc3QgbWV0cmljc1RhcmdldEdyb3VwID0gbmV3IGVsYnYyLkFwcGxpY2F0aW9uVGFyZ2V0R3JvdXAodGhpcywgJ01ldHJpY3NUYXJnZXRHcm91cCcsIHtcbiAgICAgIHZwYyxcbiAgICAgIHBvcnQ6IDkwOTEsXG4gICAgICBwcm90b2NvbDogZWxidjIuQXBwbGljYXRpb25Qcm90b2NvbC5IVFRQLFxuICAgICAgdGFyZ2V0VHlwZTogZWxidjIuVGFyZ2V0VHlwZS5JUCxcbiAgICAgIGhlYWx0aENoZWNrOiB7XG4gICAgICAgIHBhdGg6ICcvbWV0cmljcycsXG4gICAgICAgIGludGVydmFsOiBjZGsuRHVyYXRpb24uc2Vjb25kcygzMCksXG4gICAgICAgIHRpbWVvdXQ6IGNkay5EdXJhdGlvbi5zZWNvbmRzKDUpLFxuICAgICAgICBoZWFsdGh5VGhyZXNob2xkQ291bnQ6IDIsXG4gICAgICAgIHVuaGVhbHRoeVRocmVzaG9sZENvdW50OiAzLFxuICAgICAgfSxcbiAgICB9KTtcblxuICAgIC8vIEF0dGFjaCBzZXJ2aWNlIHRvIHRhcmdldCBncm91cHNcbiAgICBzZXJ2aWNlLmF0dGFjaFRvQXBwbGljYXRpb25UYXJnZXRHcm91cChodHRwVGFyZ2V0R3JvdXApO1xuICAgIHNlcnZpY2UuYXR0YWNoVG9BcHBsaWNhdGlvblRhcmdldEdyb3VwKGdycGNUYXJnZXRHcm91cCk7XG4gICAgc2VydmljZS5hdHRhY2hUb0FwcGxpY2F0aW9uVGFyZ2V0R3JvdXAobWV0cmljc1RhcmdldEdyb3VwKTtcblxuICAgIC8vIEFMQiBMaXN0ZW5lcnNcbiAgICBhbGIuYWRkTGlzdGVuZXIoJ0h0dHBMaXN0ZW5lcicsIHtcbiAgICAgIHBvcnQ6IDgwLFxuICAgICAgZGVmYXVsdFRhcmdldEdyb3VwczogW2h0dHBUYXJnZXRHcm91cF0sXG4gICAgfSk7XG5cbiAgICBhbGIuYWRkTGlzdGVuZXIoJ0dycGNMaXN0ZW5lcicsIHtcbiAgICAgIHBvcnQ6IDkwOTAsXG4gICAgICBkZWZhdWx0VGFyZ2V0R3JvdXBzOiBbZ3JwY1RhcmdldEdyb3VwXSxcbiAgICB9KTtcblxuICAgIGFsYi5hZGRMaXN0ZW5lcignTWV0cmljc0xpc3RlbmVyJywge1xuICAgICAgcG9ydDogOTA5MSxcbiAgICAgIGRlZmF1bHRUYXJnZXRHcm91cHM6IFttZXRyaWNzVGFyZ2V0R3JvdXBdLFxuICAgIH0pO1xuXG4gICAgLy8gQXV0by1zY2FsaW5nXG4gICAgY29uc3Qgc2NhbGluZyA9IHNlcnZpY2UuYXV0b1NjYWxlVGFza0NvdW50KHtcbiAgICAgIG1pbkNhcGFjaXR5OiAyLFxuICAgICAgbWF4Q2FwYWNpdHk6IDEwLFxuICAgIH0pO1xuXG4gICAgc2NhbGluZy5zY2FsZU9uQ3B1VXRpbGl6YXRpb24oJ0NwdVNjYWxpbmcnLCB7XG4gICAgICB0YXJnZXRVdGlsaXphdGlvblBlcmNlbnQ6IDcwLFxuICAgICAgc2NhbGVJbkNvb2xkb3duOiBjZGsuRHVyYXRpb24uc2Vjb25kcyg2MCksXG4gICAgICBzY2FsZU91dENvb2xkb3duOiBjZGsuRHVyYXRpb24uc2Vjb25kcyg2MCksXG4gICAgfSk7XG5cbiAgICBzY2FsaW5nLnNjYWxlT25NZW1vcnlVdGlsaXphdGlvbignTWVtb3J5U2NhbGluZycsIHtcbiAgICAgIHRhcmdldFV0aWxpemF0aW9uUGVyY2VudDogODAsXG4gICAgICBzY2FsZUluQ29vbGRvd246IGNkay5EdXJhdGlvbi5zZWNvbmRzKDYwKSxcbiAgICAgIHNjYWxlT3V0Q29vbGRvd246IGNkay5EdXJhdGlvbi5zZWNvbmRzKDYwKSxcbiAgICB9KTtcblxuICAgIC8vIE91dHB1dHNcbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnT2JzaWRpYW5FbmRwb2ludCcsIHtcbiAgICAgIHZhbHVlOiBhbGIubG9hZEJhbGFuY2VyRG5zTmFtZSxcbiAgICAgIGRlc2NyaXB0aW9uOiAnT2JzaWRpYW4gR2F0ZXdheSBlbmRwb2ludCcsXG4gICAgfSk7XG5cbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnSHR0cFVSTCcsIHtcbiAgICAgIHZhbHVlOiBgaHR0cDovLyR7YWxiLmxvYWRCYWxhbmNlckRuc05hbWV9YCxcbiAgICAgIGRlc2NyaXB0aW9uOiAnSFRUUCBlbmRwb2ludCcsXG4gICAgfSk7XG5cbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnR3JwY0VuZHBvaW50Jywge1xuICAgICAgdmFsdWU6IGAke2FsYi5sb2FkQmFsYW5jZXJEbnNOYW1lfTo5MDkwYCxcbiAgICAgIGRlc2NyaXB0aW9uOiAnZ1JQQyBlbmRwb2ludCcsXG4gICAgfSk7XG5cbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnTWV0cmljc0VuZHBvaW50Jywge1xuICAgICAgdmFsdWU6IGBodHRwOi8vJHthbGIubG9hZEJhbGFuY2VyRG5zTmFtZX06OTA5MS9tZXRyaWNzYCxcbiAgICAgIGRlc2NyaXB0aW9uOiAnUHJvbWV0aGV1cyBtZXRyaWNzIGVuZHBvaW50JyxcbiAgICB9KTtcblxuICAgIG5ldyBjZGsuQ2ZuT3V0cHV0KHRoaXMsICdDbHVzdGVyTmFtZScsIHtcbiAgICAgIHZhbHVlOiBjbHVzdGVyLmNsdXN0ZXJOYW1lLFxuICAgICAgZGVzY3JpcHRpb246ICdFQ1MgQ2x1c3RlciBuYW1lJyxcbiAgICB9KTtcblxuICAgIG5ldyBjZGsuQ2ZuT3V0cHV0KHRoaXMsICdTZXJ2aWNlTmFtZScsIHtcbiAgICAgIHZhbHVlOiBzZXJ2aWNlLnNlcnZpY2VOYW1lLFxuICAgICAgZGVzY3JpcHRpb246ICdFQ1MgU2VydmljZSBuYW1lJyxcbiAgICB9KTtcbiAgfVxufSJdfQ==