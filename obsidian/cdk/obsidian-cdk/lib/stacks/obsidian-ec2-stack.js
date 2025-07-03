"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ObsidianEC2Stack = void 0;
const cdk = require("aws-cdk-lib");
const ec2 = require("aws-cdk-lib/aws-ec2");
const iam = require("aws-cdk-lib/aws-iam");
const autoscaling = require("aws-cdk-lib/aws-autoscaling");
const artifacts_1 = require("../constructs/artifacts");
const build_pipeline_1 = require("../constructs/build-pipeline");
const build_trigger_1 = require("../constructs/build-trigger");
class ObsidianEC2Stack extends cdk.Stack {
    constructor(scope, id, props = {}) {
        super(scope, id, props);
        // Create artifacts bucket and build pipeline
        const artifacts = new artifacts_1.ObsidianArtifactsConstruct(this, 'Artifacts');
        const buildPipeline = new build_pipeline_1.ObsidianBuildPipeline(this, 'BuildPipeline', {
            artifactsBucket: artifacts.bucket,
        });
        const buildTrigger = new build_trigger_1.ObsidianBuildTrigger(this, 'BuildTrigger', {
            buildProject: buildPipeline.project,
        });
        // VPC
        const vpc = props.vpcId
            ? ec2.Vpc.fromLookup(this, 'VPC', { vpcId: props.vpcId })
            : new ec2.Vpc(this, 'ObsidianVPC', {
                maxAzs: 2,
                natGateways: 1,
            });
        // Security Group
        const securityGroup = new ec2.SecurityGroup(this, 'ObsidianSecurityGroup', {
            vpc,
            description: 'Security group for Obsidian instances',
            allowAllOutbound: true,
        });
        // Allow inbound traffic on Obsidian ports
        securityGroup.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(8080), 'Allow HTTP traffic');
        securityGroup.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(9090), 'Allow gRPC traffic');
        securityGroup.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(9091), 'Allow metrics traffic');
        // IAM Role
        const role = new iam.Role(this, 'ObsidianInstanceRole', {
            assumedBy: new iam.ServicePrincipal('ec2.amazonaws.com'),
            managedPolicies: [
                iam.ManagedPolicy.fromAwsManagedPolicyName('AmazonSSMManagedInstanceCore'),
            ],
        });
        // Add ECR permissions
        role.addToPolicy(new iam.PolicyStatement({
            effect: iam.Effect.ALLOW,
            actions: [
                'ecr:GetAuthorizationToken',
                'ecr:BatchCheckLayerAvailability',
                'ecr:GetDownloadUrlForLayer',
                'ecr:BatchGetImage',
                'ecr:DescribeImages',
            ],
            resources: ['*'],
        }));
        // Add Secrets Manager permissions
        role.addToPolicy(new iam.PolicyStatement({
            effect: iam.Effect.ALLOW,
            actions: [
                'secretsmanager:GetSecretValue',
                'secretsmanager:DescribeSecret',
            ],
            resources: [`arn:aws:secretsmanager:${this.region}:${this.account}:secret:obsidian/*`],
        }));
        // Add CloudWatch permissions
        role.addToPolicy(new iam.PolicyStatement({
            effect: iam.Effect.ALLOW,
            actions: [
                'cloudwatch:PutMetricData',
                'logs:CreateLogGroup',
                'logs:CreateLogStream',
                'logs:PutLogEvents',
            ],
            resources: ['*'],
        }));
        // Add S3 permissions to download from artifacts bucket
        role.addToPolicy(artifacts.bucketPolicy);
        // User Data script
        const userData = ec2.UserData.forLinux();
        userData.addCommands(
        // Update and install dependencies
        'yum update -y', 'yum install -y docker', 'service docker start', 'usermod -a -G docker ec2-user', 
        // Install Docker Compose
        'curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose', 'chmod +x /usr/local/bin/docker-compose', 
        // Download and start Obsidian
        'mkdir -p /opt/obsidian', 'cd /opt/obsidian', 
        // Create configuration
        'cat > /opt/obsidian/config.yaml << EOF', this.generateConfigYaml(), 'EOF', 
        // Download Obsidian binary from S3
        `aws s3 cp s3://${artifacts.bucket.bucketName}/binaries/obsidian-linux-amd64-latest /opt/obsidian/obsidian`, 'chmod +x /opt/obsidian/obsidian', 
        // Verify binary downloaded successfully
        'if [ ! -f /opt/obsidian/obsidian ]; then', '  echo "ERROR: Failed to download Obsidian binary from S3"', '  exit 1', 'fi', 
        // Create systemd service
        'cat > /etc/systemd/system/obsidian.service << EOF', '[Unit]', 'Description=Obsidian Gateway', 'After=docker.service', 'Requires=docker.service', '', '[Service]', 'Type=simple', 'ExecStart=/opt/obsidian/obsidian --config=/opt/obsidian/config.yaml', 'Restart=always', 'RestartSec=10', 'StandardOutput=journal', 'StandardError=journal', 'SyslogIdentifier=obsidian', '', '[Install]', 'WantedBy=multi-user.target', 'EOF', 
        // Start Obsidian
        'systemctl daemon-reload', 'systemctl enable obsidian', 'systemctl start obsidian');
        // Launch Template
        const launchTemplate = new ec2.LaunchTemplate(this, 'ObsidianLaunchTemplate', {
            instanceType: props.instanceType || ec2.InstanceType.of(ec2.InstanceClass.M5, ec2.InstanceSize.XLARGE2),
            machineImage: ec2.MachineImage.latestAmazonLinux2(),
            userData,
            role,
            securityGroup,
            blockDevices: [{
                    deviceName: '/dev/xvda',
                    volume: ec2.BlockDeviceVolume.ebs(100, {
                        volumeType: ec2.EbsDeviceVolumeType.GP3,
                        encrypted: true,
                    }),
                }],
        });
        // Auto Scaling Group
        const asg = new autoscaling.AutoScalingGroup(this, 'ObsidianASG', {
            vpc,
            launchTemplate,
            minCapacity: props.minSize || 1,
            maxCapacity: props.maxSize || 1,
            healthCheck: autoscaling.HealthCheck.ec2({
                grace: cdk.Duration.minutes(5),
            }),
        });
        // Outputs
        new cdk.CfnOutput(this, 'ObsidianEndpoint', {
            value: `http://${asg.connections.securityGroups[0].securityGroupId}.${this.region}.elb.amazonaws.com:8080`,
            description: 'Obsidian Gateway endpoint',
        });
        new cdk.CfnOutput(this, 'SecurityGroupId', {
            value: securityGroup.securityGroupId,
            description: 'Security group ID for Obsidian instances',
        });
    }
    generateConfigYaml() {
        return `server:
  port: 8080
  grpcPort: 9090
  maxConnections: 1000
  readTimeout: 30s
  writeTimeout: 30s
  shutdownTimeout: 30s

orchestrator:
  resources:
    maxCpu: "4"
    maxMemory: "8Gi"
    maxDisk: "50Gi"
    maxContainers: 10
  queue:
    maxQueueSize: 1000
    taskTimeout: 5m
    retryPolicy:
      maxRetries: 3
      backoffMultiplier: 2
  container:
    runtime: "docker"
    network: "bridge"
    pullPolicy: "IfNotPresent"

registry:
  registries:
    - name: "ecr"
      type: "aws-ecr"
      region: "${this.region}"
      credentialSource: "iam-role"

proxy:
  backends: []

logging:
  level: "info"
  format: "json"
  outputPath: "/var/log/obsidian/obsidian.log"

monitoring:
  metricsEnabled: true
  metricsPort: 9091`;
    }
}
exports.ObsidianEC2Stack = ObsidianEC2Stack;
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoib2JzaWRpYW4tZWMyLXN0YWNrLmpzIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsib2JzaWRpYW4tZWMyLXN0YWNrLnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiI7OztBQUFBLG1DQUFtQztBQUNuQywyQ0FBMkM7QUFDM0MsMkNBQTJDO0FBQzNDLDJEQUEyRDtBQUczRCx1REFBcUU7QUFDckUsaUVBQXFFO0FBQ3JFLCtEQUFtRTtBQVNuRSxNQUFhLGdCQUFpQixTQUFRLEdBQUcsQ0FBQyxLQUFLO0lBQzdDLFlBQVksS0FBZ0IsRUFBRSxFQUFVLEVBQUUsUUFBK0IsRUFBRTtRQUN6RSxLQUFLLENBQUMsS0FBSyxFQUFFLEVBQUUsRUFBRSxLQUFLLENBQUMsQ0FBQztRQUV4Qiw2Q0FBNkM7UUFDN0MsTUFBTSxTQUFTLEdBQUcsSUFBSSxzQ0FBMEIsQ0FBQyxJQUFJLEVBQUUsV0FBVyxDQUFDLENBQUM7UUFDcEUsTUFBTSxhQUFhLEdBQUcsSUFBSSxzQ0FBcUIsQ0FBQyxJQUFJLEVBQUUsZUFBZSxFQUFFO1lBQ3JFLGVBQWUsRUFBRSxTQUFTLENBQUMsTUFBTTtTQUNsQyxDQUFDLENBQUM7UUFDSCxNQUFNLFlBQVksR0FBRyxJQUFJLG9DQUFvQixDQUFDLElBQUksRUFBRSxjQUFjLEVBQUU7WUFDbEUsWUFBWSxFQUFFLGFBQWEsQ0FBQyxPQUFPO1NBQ3BDLENBQUMsQ0FBQztRQUVILE1BQU07UUFDTixNQUFNLEdBQUcsR0FBRyxLQUFLLENBQUMsS0FBSztZQUNyQixDQUFDLENBQUMsR0FBRyxDQUFDLEdBQUcsQ0FBQyxVQUFVLENBQUMsSUFBSSxFQUFFLEtBQUssRUFBRSxFQUFFLEtBQUssRUFBRSxLQUFLLENBQUMsS0FBSyxFQUFFLENBQUM7WUFDekQsQ0FBQyxDQUFDLElBQUksR0FBRyxDQUFDLEdBQUcsQ0FBQyxJQUFJLEVBQUUsYUFBYSxFQUFFO2dCQUMvQixNQUFNLEVBQUUsQ0FBQztnQkFDVCxXQUFXLEVBQUUsQ0FBQzthQUNmLENBQUMsQ0FBQztR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LEVBQUUsSUFBSTtxQkFDaEIsQ0FBQztpQkFDSCxDQUFDO1NBQ0gsQ0FBQyxDQUFDO1FBRUgscUJBQXFCO1FBQ3JCLE1BQU0sR0FBRyxHQUFHLElBQUksV0FBVyxDQUFDLGdCQUFnQixDQUFDLElBQUksRUFBRSxhQUFhLEVBQUU7WUFDaEUsR0FBRztZQUNILGNBQWM7WUFDZCxXQUFXLEVBQUUsS0FBSyxDQUFDLE9BQU8sSUFBSSxDQUFDO1lBQy9CLFdBQVcsRUFBRSxLQUFLLENBQUMsT0FBTyxJQUFJLENBQUM7WUFDL0IsV0FBVyxFQUFFLFdBQVcsQ0FBQyxXQUFXLENBQUMsR0FBRyxDQUFDO2dCQUN2QyxLQUFLLEVBQUUsR0FBRyxDQUFDLFFBQVEsQ0FBQyxPQUFPLENBQUMsQ0FBQyxDQUFDO2FBQy9CLENBQUM7U0FDSCxDQUFDLENBQUM7UUFFSCxVQUFVO1FBQ1YsSUFBSSxHQUFHLENBQUMsU0FBUyxDQUFDLElBQUksRUFBRSxrQkFBa0IsRUFBRTtZQUMxQyxLQUFLLEVBQUUsVUFBVSxHQUFHLENBQUMsV0FBVyxDQUFDLGNBQWMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxlQUFlLElBQUksSUFBSSxDQUFDLE1BQU0seUJBQXlCO1lBQzFHLFdBQVcsRUFBRSwyQkFBMkI7U0FDekMsQ0FBQyxDQUFDO1FBRUgsSUFBSSxHQUFHLENBQUMsU0FBUyxDQUFDLElBQUksRUFBRSxpQkFBaUIsRUFBRTtZQUN6QyxLQUFLLEVBQUUsYUFBYSxDQUFDLGVBQWU7WUFDcEMsV0FBVyxFQUFFLDBDQUEwQztTQUN4RCxDQUFDLENBQUM7SUFDTCxDQUFDO0lBRU8sa0JBQWtCO1FBQ3hCLE9BQU87Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7O2lCQTZCTSxJQUFJLENBQUMsTUFBTTs7Ozs7Ozs7Ozs7OztvQkFhUixDQUFDO0lBQ25CLENBQUM7Q0FDRjtBQTNPRCw0Q0EyT0MiLCJzb3VyY2VzQ29udGVudCI6WyJpbXBvcnQgKiBhcyBjZGsgZnJvbSAnYXdzLWNkay1saWInO1xuaW1wb3J0ICogYXMgZWMyIGZyb20gJ2F3cy1jZGstbGliL2F3cy1lYzInO1xuaW1wb3J0ICogYXMgaWFtIGZyb20gJ2F3cy1jZGstbGliL2F3cy1pYW0nO1xuaW1wb3J0ICogYXMgYXV0b3NjYWxpbmcgZnJvbSAnYXdzLWNkay1saWIvYXdzLWF1dG9zY2FsaW5nJztcbmltcG9ydCAqIGFzIHMzIGZyb20gJ2F3cy1jZGstbGliL2F3cy1zMyc7XG5pbXBvcnQgeyBDb25zdHJ1Y3QgfSBmcm9tICdjb25zdHJ1Y3RzJztcbmltcG9ydCB7IE9ic2lkaWFuQXJ0aWZhY3RzQ29uc3RydWN0IH0gZnJvbSAnLi4vY29uc3RydWN0cy9hcnRpZmFjdHMnO1xuaW1wb3J0IHsgT2JzaWRpYW5CdWlsZFBpcGVsaW5lIH0gZnJvbSAnLi4vY29uc3RydWN0cy9idWlsZC1waXBlbGluZSc7XG5pbXBvcnQgeyBPYnNpZGlhbkJ1aWxkVHJpZ2dlciB9IGZyb20gJy4uL2NvbnN0cnVjdHMvYnVpbGQtdHJpZ2dlcic7XG5cbmV4cG9ydCBpbnRlcmZhY2UgT2JzaWRpYW5FQzJTdGFja1Byb3BzIGV4dGVuZHMgY2RrLlN0YWNrUHJvcHMge1xuICBpbnN0YW5jZVR5cGU/OiBlYzIuSW5zdGFuY2VUeXBlO1xuICBtaW5TaXplPzogbnVtYmVyO1xuICBtYXhTaXplPzogbnVtYmVyO1xuICB2cGNJZD86IHN0cmluZztcbn1cblxuZXhwb3J0IGNsYXNzIE9ic2lkaWFuRUMyU3RhY2sgZXh0ZW5kcyBjZGsuU3RhY2sge1xuICBjb25zdHJ1Y3RvcihzY29wZTogQ29uc3RydWN0LCBpZDogc3RyaW5nLCBwcm9wczogT2JzaWRpYW5FQzJTdGFja1Byb3BzID0ge30pIHtcbiAgICBzdXBlcihzY29wZSwgaWQsIHByb3BzKTtcblxuICAgIC8vIENyZWF0ZSBhcnRpZmFjdHMgYnVja2V0IGFuZCBidWlsZCBwaXBlbGluZVxuICAgIGNvbnN0IGFydGlmYWN0cyA9IG5ldyBPYnNpZGlhbkFydGlmYWN0c0NvbnN0cnVjdCh0aGlzLCAnQXJ0aWZhY3RzJyk7XG4gICAgY29uc3QgYnVpbGRQaXBlbGluZSA9IG5ldyBPYnNpZGlhbkJ1aWxkUGlwZWxpbmUodGhpcywgJ0J1aWxkUGlwZWxpbmUnLCB7XG4gICAgICBhcnRpZmFjdHNCdWNrZXQ6IGFydGlmYWN0cy5idWNrZXQsXG4gICAgfSk7XG4gICAgY29uc3QgYnVpbGRUcmlnZ2VyID0gbmV3IE9ic2lkaWFuQnVpbGRUcmlnZ2VyKHRoaXMsICdCdWlsZFRyaWdnZXInLCB7XG4gICAgICBidWlsZFByb2plY3Q6IGJ1aWxkUGlwZWxpbmUucHJvamVjdCxcbiAgICB9KTtcblxuICAgIC8vIFZQQ1xuICAgIGNvbnN0IHZwYyA9IHByb3BzLnZwY0lkIFxuICAgICAgPyBlYzIuVnBjLmZyb21Mb29rdXAodGhpcywgJ1ZQQycsIHsgdnBjSWQ6IHByb3BzLnZwY0lkIH0pXG4gICAgICA6IG5ldyBlYzIuVnBjKHRoaXMsICdPYnNpZGlhblZQQycsIHtcbiAgICAgICAgICBtYXhBenM6IDIsXG4gICAgICAgICAgbmF0R2F0ZXdheXM6IDEsXG4gICAgICAgIH0pO1xuXG4gICAgLy8gU2VjdXJpdHkgR3JvdXBcbiAgICBjb25zdCBzZWN1cml0eUdyb3VwID0gbmV3IGVjMi5TZWN1cml0eUdyb3VwKHRoaXMsICdPYnNpZGlhblNlY3VyaXR5R3JvdXAnLCB7XG4gICAgICB2cGMsXG4gICAgICBkZXNjcmlwdGlvbjogJ1NlY3VyaXR5IGdyb3VwIGZvciBPYnNpZGlhbiBpbnN0YW5jZXMnLFxuICAgICAgYWxsb3dBbGxPdXRib3VuZDogdHJ1ZSxcbiAgICB9KTtcblxuICAgIC8vIEFsbG93IGluYm91bmQgdHJhZmZpYyBvbiBPYnNpZGlhbiBwb3J0c1xuICAgIHNlY3VyaXR5R3JvdXAuYWRkSW5ncmVzc1J1bGUoXG4gICAgICBlYzIuUGVlci5hbnlJcHY0KCksXG4gICAgICBlYzIuUG9ydC50Y3AoODA4MCksXG4gICAgICAnQWxsb3cgSFRUUCB0cmFmZmljJ1xuICAgICk7XG5cbiAgICBzZWN1cml0eUdyb3VwLmFkZEluZ3Jlc3NSdWxlKFxuICAgICAgZWMyLlBlZXIuYW55SXB2NCgpLFxuICAgICAgZWMyLlBvcnQudGNwKDkwOTApLFxuICAgICAgJ0FsbG93IGdSUEMgdHJhZmZpYydcbiAgICApO1xuXG4gICAgc2VjdXJpdHlHcm91cC5hZGRJbmdyZXNzUnVsZShcbiAgICAgIGVjMi5QZWVyLmFueUlwdjQoKSxcbiAgICAgIGVjMi5Qb3J0LnRjcCg5MDkxKSxcbiAgICAgICdBbGxvdyBtZXRyaWNzIHRyYWZmaWMnXG4gICAgKTtcblxuICAgIC8vIElBTSBSb2xlXG4gICAgY29uc3Qgcm9sZSA9IG5ldyBpYW0uUm9sZSh0aGlzLCAnT2JzaWRpYW5JbnN0YW5jZVJvbGUnLCB7XG4gICAgICBhc3N1bWVkQnk6IG5ldyBpYW0uU2VydmljZVByaW5jaXBhbCgnZWMyLmFtYXpvbmF3cy5jb20nKSxcbiAgICAgIG1hbmFnZWRQb2xpY2llczogW1xuICAgICAgICBpYW0uTWFuYWdlZFBvbGljeS5mcm9tQXdzTWFuYWdlZFBvbGljeU5hbWUoJ0FtYXpvblNTTU1hbmFnZWRJbnN0YW5jZUNvcmUnKSxcbiAgICAgIF0sXG4gICAgfSk7XG5cbiAgICAvLyBBZGQgRUNSIHBlcm1pc3Npb25zXG4gICAgcm9sZS5hZGRUb1BvbGljeShuZXcgaWFtLlBvbGljeVN0YXRlbWVudCh7XG4gICAgICBlZmZlY3Q6IGlhbS5FZmZlY3QuQUxMT1csXG4gICAgICBhY3Rpb25zOiBbXG4gICAgICAgICdlY3I6R2V0QXV0aG9yaXphdGlvblRva2VuJyxcbiAgICAgICAgJ2VjcjpCYXRjaENoZWNrTGF5ZXJBdmFpbGFiaWxpdHknLFxuICAgICAgICAnZWNyOkdldERvd25sb2FkVXJsRm9yTGF5ZXInLFxuICAgICAgICAnZWNyOkJhdGNoR2V0SW1hZ2UnLFxuICAgICAgICAnZWNyOkRlc2NyaWJlSW1hZ2VzJyxcbiAgICAgIF0sXG4gICAgICByZXNvdXJjZXM6IFsnKiddLFxuICAgIH0pKTtcblxuICAgIC8vIEFkZCBTZWNyZXRzIE1hbmFnZXIgcGVybWlzc2lvbnNcbiAgICByb2xlLmFkZFRvUG9saWN5KG5ldyBpYW0uUG9saWN5U3RhdGVtZW50KHtcbiAgICAgIGVmZmVjdDogaWFtLkVmZmVjdC5BTExPVyxcbiAgICAgIGFjdGlvbnM6IFtcbiAgICAgICAgJ3NlY3JldHNtYW5hZ2VyOkdldFNlY3JldFZhbHVlJyxcbiAgICAgICAgJ3NlY3JldHNtYW5hZ2VyOkRlc2NyaWJlU2VjcmV0JyxcbiAgICAgIF0sXG4gICAgICByZXNvdXJjZXM6IFtgYXJuOmF3czpzZWNyZXRzbWFuYWdlcjoke3RoaXMucmVnaW9ufToke3RoaXMuYWNjb3VudH06c2VjcmV0Om9ic2lkaWFuLypgXSxcbiAgICB9KSk7XG5cbiAgICAvLyBBZGQgQ2xvdWRXYXRjaCBwZXJtaXNzaW9uc1xuICAgIHJvbGUuYWRkVG9Qb2xpY3kobmV3IGlhbS5Qb2xpY3lTdGF0ZW1lbnQoe1xuICAgICAgZWZmZWN0OiBpYW0uRWZmZWN0LkFMTE9XLFxuICAgICAgYWN0aW9uczogW1xuICAgICAgICAnY2xvdWR3YXRjaDpQdXRNZXRyaWNEYXRhJyxcbiAgICAgICAgJ2xvZ3M6Q3JlYXRlTG9nR3JvdXAnLFxuICAgICAgICAnbG9nczpDcmVhdGVMb2dTdHJlYW0nLFxuICAgICAgICAnbG9nczpQdXRMb2dFdmVudHMnLFxuICAgICAgXSxcbiAgICAgIHJlc291cmNlczogWycqJ10sXG4gICAgfSkpO1xuXG4gICAgLy8gQWRkIFMzIHBlcm1pc3Npb25zIHRvIGRvd25sb2FkIGZyb20gYXJ0aWZhY3RzIGJ1Y2tldFxuICAgIHJvbGUuYWRkVG9Qb2xpY3koYXJ0aWZhY3RzLmJ1Y2tldFBvbGljeSk7XG5cbiAgICAvLyBVc2VyIERhdGEgc2NyaXB0XG4gICAgY29uc3QgdXNlckRhdGEgPSBlYzIuVXNlckRhdGEuZm9yTGludXgoKTtcbiAgICB1c2VyRGF0YS5hZGRDb21tYW5kcyhcbiAgICAgIC8vIFVwZGF0ZSBhbmQgaW5zdGFsbCBkZXBlbmRlbmNpZXNcbiAgICAgICd5dW0gdXBkYXRlIC15JyxcbiAgICAgICd5dW0gaW5zdGFsbCAteSBkb2NrZXInLFxuICAgICAgJ3NlcnZpY2UgZG9ja2VyIHN0YXJ0JyxcbiAgICAgICd1c2VybW9kIC1hIC1HIGRvY2tlciBlYzItdXNlcicsXG4gICAgICBcbiAgICAgIC8vIEluc3RhbGwgRG9ja2VyIENvbXBvc2VcbiAgICAgICdjdXJsIC1MIFwiaHR0cHM6Ly9naXRodWIuY29tL2RvY2tlci9jb21wb3NlL3JlbGVhc2VzL2xhdGVzdC9kb3dubG9hZC9kb2NrZXItY29tcG9zZS0kKHVuYW1lIC1zKS0kKHVuYW1lIC1tKVwiIC1vIC91c3IvbG9jYWwvYmluL2RvY2tlci1jb21wb3NlJyxcbiAgICAgICdjaG1vZCAreCAvdXNyL2xvY2FsL2Jpbi9kb2NrZXItY29tcG9zZScsXG4gICAgICBcbiAgICAgIC8vIERvd25sb2FkIGFuZCBzdGFydCBPYnNpZGlhblxuICAgICAgJ21rZGlyIC1wIC9vcHQvb2JzaWRpYW4nLFxuICAgICAgJ2NkIC9vcHQvb2JzaWRpYW4nLFxuICAgICAgXG4gICAgICAvLyBDcmVhdGUgY29uZmlndXJhdGlvblxuICAgICAgJ2NhdCA+IC9vcHQvb2JzaWRpYW4vY29uZmlnLnlhbWwgPDwgRU9GJyxcbiAgICAgIHRoaXMuZ2VuZXJhdGVDb25maWdZYW1sKCksXG4gICAgICAnRU9GJyxcbiAgICAgIFxuICAgICAgLy8gRG93bmxvYWQgT2JzaWRpYW4gYmluYXJ5IGZyb20gUzNcbiAgICAgIGBhd3MgczMgY3AgczM6Ly8ke2FydGlmYWN0cy5idWNrZXQuYnVja2V0TmFtZX0vYmluYXJpZXMvb2JzaWRpYW4tbGludXgtYW1kNjQtbGF0ZXN0IC9vcHQvb2JzaWRpYW4vb2JzaWRpYW5gLFxuICAgICAgJ2NobW9kICt4IC9vcHQvb2JzaWRpYW4vb2JzaWRpYW4nLFxuICAgICAgXG4gICAgICAvLyBWZXJpZnkgYmluYXJ5IGRvd25sb2FkZWQgc3VjY2Vzc2Z1bGx5XG4gICAgICAnaWYgWyAhIC1mIC9vcHQvb2JzaWRpYW4vb2JzaWRpYW4gXTsgdGhlbicsXG4gICAgICAnICBlY2hvIFwiRVJST1I6IEZhaWxlZCB0byBkb3dubG9hZCBPYnNpZGlhbiBiaW5hcnkgZnJvbSBTM1wiJyxcbiAgICAgICcgIGV4aXQgMScsXG4gICAgICAnZmknLFxuICAgICAgXG4gICAgICAvLyBDcmVhdGUgc3lzdGVtZCBzZXJ2aWNlXG4gICAgICAnY2F0ID4gL2V0Yy9zeXN0ZW1kL3N5c3RlbS9vYnNpZGlhbi5zZXJ2aWNlIDw8IEVPRicsXG4gICAgICAnW1VuaXRdJyxcbiAgICAgICdEZXNjcmlwdGlvbj1PYnNpZGlhbiBHYXRld2F5JyxcbiAgICAgICdBZnRlcj1kb2NrZXIuc2VydmljZScsXG4gICAgICAnUmVxdWlyZXM9ZG9ja2VyLnNlcnZpY2UnLFxuICAgICAgJycsXG4gICAgICAnW1NlcnZpY2VdJyxcbiAgICAgICdUeXBlPXNpbXBsZScsXG4gICAgICAnRXhlY1N0YXJ0PS9vcHQvb2JzaWRpYW4vb2JzaWRpYW4gLS1jb25maWc9L29wdC9vYnNpZGlhbi9jb25maWcueWFtbCcsXG4gICAgICAnUmVzdGFydD1hbHdheXMnLFxuICAgICAgJ1Jlc3RhcnRTZWM9MTAnLFxuICAgICAgJ1N0YW5kYXJkT3V0cHV0PWpvdXJuYWwnLFxuICAgICAgJ1N0YW5kYXJkRXJyb3I9am91cm5hbCcsXG4gICAgICAnU3lzbG9nSWRlbnRpZmllcj1vYnNpZGlhbicsXG4gICAgICAnJyxcbiAgICAgICdbSW5zdGFsbF0nLFxuICAgICAgJ1dhbnRlZEJ5PW11bHRpLXVzZXIudGFyZ2V0JyxcbiAgICAgICdFT0YnLFxuICAgICAgXG4gICAgICAvLyBTdGFydCBPYnNpZGlhblxuICAgICAgJ3N5c3RlbWN0bCBkYWVtb24tcmVsb2FkJyxcbiAgICAgICdzeXN0ZW1jdGwgZW5hYmxlIG9ic2lkaWFuJyxcbiAgICAgICdzeXN0ZW1jdGwgc3RhcnQgb2JzaWRpYW4nXG4gICAgKTtcblxuICAgIC8vIExhdW5jaCBUZW1wbGF0ZVxuICAgIGNvbnN0IGxhdW5jaFRlbXBsYXRlID0gbmV3IGVjMi5MYXVuY2hUZW1wbGF0ZSh0aGlzLCAnT2JzaWRpYW5MYXVuY2hUZW1wbGF0ZScsIHtcbiAgICAgIGluc3RhbmNlVHlwZTogcHJvcHMuaW5zdGFuY2VUeXBlIHx8IGVjMi5JbnN0YW5jZVR5cGUub2YoZWMyLkluc3RhbmNlQ2xhc3MuTTUsIGVjMi5JbnN0YW5jZVNpemUuWExBUkdFMiksXG4gICAgICBtYWNoaW5lSW1hZ2U6IGVjMi5NYWNoaW5lSW1hZ2UubGF0ZXN0QW1hem9uTGludXgyKCksXG4gICAgICB1c2VyRGF0YSxcbiAgICAgIHJvbGUsXG4gICAgICBzZWN1cml0eUdyb3VwLFxuICAgICAgYmxvY2tEZXZpY2VzOiBbe1xuICAgICAgICBkZXZpY2VOYW1lOiAnL2Rldi94dmRhJyxcbiAgICAgICAgdm9sdW1lOiBlYzIuQmxvY2tEZXZpY2VWb2x1bWUuZWJzKDEwMCwge1xuICAgICAgICAgIHZvbHVtZVR5cGU6IGVjMi5FYnNEZXZpY2VWb2x1bWVUeXBlLkdQMyxcbiAgICAgICAgICBlbmNyeXB0ZWQ6IHRydWUsXG4gICAgICAgIH0pLFxuICAgICAgfV0sXG4gICAgfSk7XG5cbiAgICAvLyBBdXRvIFNjYWxpbmcgR3JvdXBcbiAgICBjb25zdCBhc2cgPSBuZXcgYXV0b3NjYWxpbmcuQXV0b1NjYWxpbmdHcm91cCh0aGlzLCAnT2JzaWRpYW5BU0cnLCB7XG4gICAgICB2cGMsXG4gICAgICBsYXVuY2hUZW1wbGF0ZSxcbiAgICAgIG1pbkNhcGFjaXR5OiBwcm9wcy5taW5TaXplIHx8IDEsXG4gICAgICBtYXhDYXBhY2l0eTogcHJvcHMubWF4U2l6ZSB8fCAxLFxuICAgICAgaGVhbHRoQ2hlY2s6IGF1dG9zY2FsaW5nLkhlYWx0aENoZWNrLmVjMih7XG4gICAgICAgIGdyYWNlOiBjZGsuRHVyYXRpb24ubWludXRlcyg1KSxcbiAgICAgIH0pLFxuICAgIH0pO1xuXG4gICAgLy8gT3V0cHV0c1xuICAgIG5ldyBjZGsuQ2ZuT3V0cHV0KHRoaXMsICdPYnNpZGlhbkVuZHBvaW50Jywge1xuICAgICAgdmFsdWU6IGBodHRwOi8vJHthc2cuY29ubmVjdGlvbnMuc2VjdXJpdHlHcm91cHNbMF0uc2VjdXJpdHlHcm91cElkfS4ke3RoaXMucmVnaW9ufS5lbGIuYW1hem9uYXdzLmNvbTo4MDgwYCxcbiAgICAgIGRlc2NyaXB0aW9uOiAnT2JzaWRpYW4gR2F0ZXdheSBlbmRwb2ludCcsXG4gICAgfSk7XG5cbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnU2VjdXJpdHlHcm91cElkJywge1xuICAgICAgdmFsdWU6IHNlY3VyaXR5R3JvdXAuc2VjdXJpdHlHcm91cElkLFxuICAgICAgZGVzY3JpcHRpb246ICdTZWN1cml0eSBncm91cCBJRCBmb3IgT2JzaWRpYW4gaW5zdGFuY2VzJyxcbiAgICB9KTtcbiAgfVxuXG4gIHByaXZhdGUgZ2VuZXJhdGVDb25maWdZYW1sKCk6IHN0cmluZyB7XG4gICAgcmV0dXJuIGBzZXJ2ZXI6XG4gIHBvcnQ6IDgwODBcbiAgZ3JwY1BvcnQ6IDkwOTBcbiAgbWF4Q29ubmVjdGlvbnM6IDEwMDBcbiAgcmVhZFRpbWVvdXQ6IDMwc1xuICB3cml0ZVRpbWVvdXQ6IDMwc1xuICBzaHV0ZG93blRpbWVvdXQ6IDMwc1xuXG5vcmNoZXN0cmF0b3I6XG4gIHJlc291cmNlczpcbiAgICBtYXhDcHU6IFwiNFwiXG4gICAgbWF4TWVtb3J5OiBcIjhHaVwiXG4gICAgbWF4RGlzazogXCI1MEdpXCJcbiAgICBtYXhDb250YWluZXJzOiAxMFxuICBxdWV1ZTpcbiAgICBtYXhRdWV1ZVNpemU6IDEwMDBcbiAgICB0YXNrVGltZW91dDogNW1cbiAgICByZXRyeVBvbGljeTpcbiAgICAgIG1heFJldHJpZXM6IDNcbiAgICAgIGJhY2tvZmZNdWx0aXBsaWVyOiAyXG4gIGNvbnRhaW5lcjpcbiAgICBydW50aW1lOiBcImRvY2tlclwiXG4gICAgbmV0d29yazogXCJicmlkZ2VcIlxuICAgIHB1bGxQb2xpY3k6IFwiSWZOb3RQcmVzZW50XCJcblxucmVnaXN0cnk6XG4gIHJlZ2lzdHJpZXM6XG4gICAgLSBuYW1lOiBcImVjclwiXG4gICAgICB0eXBlOiBcImF3cy1lY3JcIlxuICAgICAgcmVnaW9uOiBcIiR7dGhpcy5yZWdpb259XCJcbiAgICAgIGNyZWRlbnRpYWxTb3VyY2U6IFwiaWFtLXJvbGVcIlxuXG5wcm94eTpcbiAgYmFja2VuZHM6IFtdXG5cbmxvZ2dpbmc6XG4gIGxldmVsOiBcImluZm9cIlxuICBmb3JtYXQ6IFwianNvblwiXG4gIG91dHB1dFBhdGg6IFwiL3Zhci9sb2cvb2JzaWRpYW4vb2JzaWRpYW4ubG9nXCJcblxubW9uaXRvcmluZzpcbiAgbWV0cmljc0VuYWJsZWQ6IHRydWVcbiAgbWV0cmljc1BvcnQ6IDkwOTFgO1xuICB9XG59Il19