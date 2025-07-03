"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ObsidianSimpleStack = void 0;
const cdk = require("aws-cdk-lib");
const codebuild = require("aws-cdk-lib/aws-codebuild");
const s3 = require("aws-cdk-lib/aws-s3");
const iam = require("aws-cdk-lib/aws-iam");
const ec2 = require("aws-cdk-lib/aws-ec2");
const autoscaling = require("aws-cdk-lib/aws-autoscaling");
class ObsidianSimpleStack extends cdk.Stack {
    constructor(scope, id, props) {
        super(scope, id, props);
        // 🔹 Artifact bucket for Obsidian binaries
        const artifactBucket = new s3.Bucket(this, 'ObsidianArtifacts', {
            bucketName: `obsidian-artifacts-${this.account}-${this.region}`,
            removalPolicy: cdk.RemovalPolicy.RETAIN,
            versioned: true,
        });
        // 🔹 CodeBuild role
        const buildRole = new iam.Role(this, 'ObsidianBuildRole', {
            assumedBy: new iam.ServicePrincipal('codebuild.amazonaws.com'),
        });
        artifactBucket.grantReadWrite(buildRole);
        // 🔹 CodeBuild project to build Obsidian
        const buildProject = new codebuild.Project(this, 'ObsidianBuildProject', {
            role: buildRole,
            source: codebuild.Source.gitHub({
                owner: 'Layr-Labs',
                repo: 'hourglass-monorepo',
            }),
            environment: {
                buildImage: codebuild.LinuxBuildImage.STANDARD_7_0,
                computeType: codebuild.ComputeType.MEDIUM,
            },
            artifacts: codebuild.Artifacts.s3({
                bucket: artifactBucket,
                includeBuildId: false,
                packageZip: false,
                path: 'binaries',
                name: 'obsidian-linux-amd64',
            }),
            buildSpec: codebuild.BuildSpec.fromObject({
                version: '0.2',
                phases: {
                    install: {
                        'runtime-versions': {
                            golang: '1.21',
                        },
                        commands: [
                            'apt-get update -y',
                            'apt-get install -y protobuf-compiler',
                            'go install google.golang.org/protobuf/cmd/protoc-gen-go@latest',
                            'go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest',
                        ],
                    },
                    build: {
                        commands: [
                            'cd obsidian',
                            'make proto',
                            'GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o obsidian-linux-amd64 ./cmd/obsidian',
                        ],
                    },
                },
                artifacts: {
                    files: ['obsidian/obsidian-linux-amd64'],
                    'discard-paths': 'yes',
                },
            }),
        });
        // 🔹 VPC for EC2 instances
        const vpc = new ec2.Vpc(this, 'ObsidianVPC', {
            maxAzs: 2,
            natGateways: 1,
        });
        // 🔹 EC2 role with permissions
        const ec2Role = new iam.Role(this, 'ObsidianInstanceRole', {
            assumedBy: new iam.ServicePrincipal('ec2.amazonaws.com'),
            managedPolicies: [
                iam.ManagedPolicy.fromAwsManagedPolicyName('AmazonSSMManagedInstanceCore'),
            ],
        });
        // Grant S3 read access
        artifactBucket.grantRead(ec2Role);
        // Add ECR permissions for pulling container images
        ec2Role.addToPolicy(new iam.PolicyStatement({
            effect: iam.Effect.ALLOW,
            actions: [
                'ecr:GetAuthorizationToken',
                'ecr:BatchCheckLayerAvailability',
                'ecr:GetDownloadUrlForLayer',
                'ecr:BatchGetImage',
            ],
            resources: ['*'],
        }));
        // 🔹 Security group
        const securityGroup = new ec2.SecurityGroup(this, 'ObsidianSecurityGroup', {
            vpc,
            description: 'Security group for Obsidian instances',
        });
        securityGroup.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(8080), 'HTTP API');
        securityGroup.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(9090), 'gRPC API');
        // 🔹 User data script
        const userData = ec2.UserData.forLinux();
        userData.addCommands(
        // Install Docker
        'yum update -y', 'yum install -y docker aws-cli', 'service docker start', 'usermod -a -G docker ec2-user', 
        // Setup Obsidian
        'mkdir -p /opt/obsidian', 
        // Create default config
        'cat > /opt/obsidian/config.yaml << EOF', this.generateMinimalConfig(), 'EOF', 
        // Download Obsidian binary from S3
        `aws s3 cp s3://${artifactBucket.bucketName}/binaries/obsidian-linux-amd64 /opt/obsidian/obsidian || echo "Binary not yet available"`, 'chmod +x /opt/obsidian/obsidian || true', 
        // Create systemd service
        'cat > /etc/systemd/system/obsidian.service << EOF', '[Unit]', 'Description=Obsidian Gateway', 'After=docker.service', '', '[Service]', 'Type=simple', 'ExecStart=/opt/obsidian/obsidian --config=/opt/obsidian/config.yaml', 'Restart=always', 'StandardOutput=journal', '', '[Install]', 'WantedBy=multi-user.target', 'EOF', 'systemctl daemon-reload', 'systemctl enable obsidian', '# Start only if binary exists', '[ -f /opt/obsidian/obsidian ] && systemctl start obsidian || echo "Waiting for binary"');
        // 🔹 Launch template
        const launchTemplate = new ec2.LaunchTemplate(this, 'ObsidianLaunchTemplate', {
            instanceType: ec2.InstanceType.of(ec2.InstanceClass.M5, ec2.InstanceSize.LARGE),
            machineImage: ec2.MachineImage.latestAmazonLinux2(),
            userData,
            role: ec2Role,
            securityGroup,
        });
        // 🔹 Auto scaling group
        const asg = new autoscaling.AutoScalingGroup(this, 'ObsidianASG', {
            vpc,
            launchTemplate,
            minCapacity: 1,
            maxCapacity: 1,
        });
        // 🔹 Outputs
        new cdk.CfnOutput(this, 'BucketName', {
            value: artifactBucket.bucketName,
            description: 'S3 bucket containing Obsidian binaries',
        });
        new cdk.CfnOutput(this, 'BuildProjectName', {
            value: buildProject.projectName,
            description: 'CodeBuild project name',
        });
        new cdk.CfnOutput(this, 'BuildCommand', {
            value: `aws codebuild start-build --project-name ${buildProject.projectName}`,
            description: 'Command to trigger a build',
        });
        new cdk.CfnOutput(this, 'InstanceUpdateCommand', {
            value: `aws ssm send-command --document-name "AWS-RunShellScript" --targets "Key=tag:aws:autoscaling:groupName,Values=${asg.autoScalingGroupName}" --parameters 'commands=["aws s3 cp s3://${artifactBucket.bucketName}/binaries/obsidian-linux-amd64 /opt/obsidian/obsidian && chmod +x /opt/obsidian/obsidian && systemctl restart obsidian"]'`,
            description: 'Command to update Obsidian on running instances',
        });
    }
    generateMinimalConfig() {
        return `server:
  port: 8080
  grpcPort: 9090

orchestrator:
  resources:
    maxCpu: "4"
    maxMemory: "8Gi" 
    maxContainers: 10

registry:
  registries:
    - name: "ecr"
      type: "aws-ecr"
      region: "${this.region}"
      credentialSource: "iam-role"

proxy:
  backends: []

logging:
  level: "info"`;
    }
}
exports.ObsidianSimpleStack = ObsidianSimpleStack;
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoib2JzaWRpYW4tc2ltcGxlLXN0YWNrLmpzIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsib2JzaWRpYW4tc2ltcGxlLXN0YWNrLnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiI7OztBQUFBLG1DQUFtQztBQUNuQyx1REFBdUQ7QUFDdkQseUNBQXlDO0FBQ3pDLDJDQUEyQztBQUMzQywyQ0FBMkM7QUFDM0MsMkRBQTJEO0FBRzNELE1BQWEsbUJBQW9CLFNBQVEsR0FBRyxDQUFDLEtBQUs7SUFDaEQsWUFBWSxLQUFnQixFQUFFLEVBQVUsRUFBRSxLQUFzQjtRQUM5RCxLQUFLLENBQUMsS0FBSyxFQUFFLEVBQUUsRUFBRSxLQUFLLENBQUMsQ0FBQztRQUV4QiwyQ0FBMkM7UUFDM0MsTUFBTSxjQUFjLEdBQUcsSUFBSSxFQUFFLENBQUMsTUFBTSxDQUFDLElBQUksRUFBRSxtQkFBbUIsRUFBRTtZQUM5RCxVQUFVLEVBQUUsc0JBQXNCLElBQUksQ0FBQyxPQUFPLElBQUksSUFBSSxDQUFDLE1BQU0sRUFBRTtZQUMvRCxhQUFhLEVBQUUsR0FBRyxDQUFDLGFBQWEsQ0FBQyxNQUFNO1lBQ3ZDLFNBQVMsRUFBRSxJQUFJO1NBQ2hCLENBQUMsQ0FBQztRQUVILG9CQUFvQjtRQUNwQixNQUFNLFNBQVMsR0FBRyxJQUFJLEdBQUcsQ0FBQyxJQUFJLENBQUMsSUFBSSxFQUFFLG1CQUFtQixFQUFFO1lBQ3hELFNBQVMsRUFBRSxJQUFJLEdBQUcsQ0FBQyxnQkFBZ0IsQ0FBQyx5QkFBeUIsQ0FBQztTQUMvRCxDQUFDLENBQUM7UUFDSCxjQUFjLENBQUMsY0FBYyxDQUFDLFNBQVMsQ0FBQyxDQUFDO1FBRXpDLHlDQUF5QztRQUN6QyxNQUFNLFlBQVksR0FBRyxJQUFJLFNBQVMsQ0FBQyxPQUFPLENBQUMsSUFBSSxFQUFFLHNCQUFzQixFQUFFO1lBQ3ZFLElBQUksRUFBRSxTQUFTO1lBQ2YsTUFBTSxFQUFFLFNBQVMsQ0FBQyxNQUFNLENBQUMsTUFBTSxDQUFDO2dCQUM5QixL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LENBQUM7WUFDZCxXQUFXLEVBQUUsQ0FBQztTQUNmLENBQUMsQ0FBQztRQUVILGFBQWE7UUFDYixJQUFJLEdBQUcsQ0FBQyxTQUFTLENBQUMsSUFBSSxFQUFFLFlBQVksRUFBRTtZQUNwQyxLQUFLLEVBQUUsY0FBYyxDQUFDLFVBQVU7WUFDaEMsV0FBVyxFQUFFLHdDQUF3QztTQUN0RCxDQUFDLENBQUM7UUFFSCxJQUFJLEdBQUcsQ0FBQyxTQUFTLENBQUMsSUFBSSxFQUFFLGtCQUFrQixFQUFFO1lBQzFDLEtBQUssRUFBRSxZQUFZLENBQUMsV0FBVztZQUMvQixXQUFXLEVBQUUsd0JBQXdCO1NBQ3RDLENBQUMsQ0FBQztRQUVILElBQUksR0FBRyxDQUFDLFNBQVMsQ0FBQyxJQUFJLEVBQUUsY0FBYyxFQUFFO1lBQ3RDLEtBQUssRUFBRSw0Q0FBNEMsWUFBWSxDQUFDLFdBQVcsRUFBRTtZQUM3RSxXQUFXLEVBQUUsNEJBQTRCO1NBQzFDLENBQUMsQ0FBQztRQUVILElBQUksR0FBRyxDQUFDLFNBQVMsQ0FBQyxJQUFJLEVBQUUsdUJBQXVCLEVBQUU7WUFDL0MsS0FBSyxFQUFFLGlIQUFpSCxHQUFHLENBQUMsb0JBQW9CLDZDQUE2QyxjQUFjLENBQUMsVUFBVSwySEFBMkg7WUFDalYsV0FBVyxFQUFFLGlEQUFpRDtTQUMvRCxDQUFDLENBQUM7SUFDTCxDQUFDO0lBRU8scUJBQXFCO1FBQzNCLE9BQU87Ozs7Ozs7Ozs7Ozs7O2lCQWNNLElBQUksQ0FBQyxNQUFNOzs7Ozs7O2dCQU9aLENBQUM7SUFDZixDQUFDO0NBQ0Y7QUF6TkQsa0RBeU5DIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0ICogYXMgY2RrIGZyb20gJ2F3cy1jZGstbGliJztcbmltcG9ydCAqIGFzIGNvZGVidWlsZCBmcm9tICdhd3MtY2RrLWxpYi9hd3MtY29kZWJ1aWxkJztcbmltcG9ydCAqIGFzIHMzIGZyb20gJ2F3cy1jZGstbGliL2F3cy1zMyc7XG5pbXBvcnQgKiBhcyBpYW0gZnJvbSAnYXdzLWNkay1saWIvYXdzLWlhbSc7XG5pbXBvcnQgKiBhcyBlYzIgZnJvbSAnYXdzLWNkay1saWIvYXdzLWVjMic7XG5pbXBvcnQgKiBhcyBhdXRvc2NhbGluZyBmcm9tICdhd3MtY2RrLWxpYi9hd3MtYXV0b3NjYWxpbmcnO1xuaW1wb3J0IHsgQ29uc3RydWN0IH0gZnJvbSAnY29uc3RydWN0cyc7XG5cbmV4cG9ydCBjbGFzcyBPYnNpZGlhblNpbXBsZVN0YWNrIGV4dGVuZHMgY2RrLlN0YWNrIHtcbiAgY29uc3RydWN0b3Ioc2NvcGU6IENvbnN0cnVjdCwgaWQ6IHN0cmluZywgcHJvcHM/OiBjZGsuU3RhY2tQcm9wcykge1xuICAgIHN1cGVyKHNjb3BlLCBpZCwgcHJvcHMpO1xuXG4gICAgLy8g8J+UuSBBcnRpZmFjdCBidWNrZXQgZm9yIE9ic2lkaWFuIGJpbmFyaWVzXG4gICAgY29uc3QgYXJ0aWZhY3RCdWNrZXQgPSBuZXcgczMuQnVja2V0KHRoaXMsICdPYnNpZGlhbkFydGlmYWN0cycsIHtcbiAgICAgIGJ1Y2tldE5hbWU6IGBvYnNpZGlhbi1hcnRpZmFjdHMtJHt0aGlzLmFjY291bnR9LSR7dGhpcy5yZWdpb259YCxcbiAgICAgIHJlbW92YWxQb2xpY3k6IGNkay5SZW1vdmFsUG9saWN5LlJFVEFJTixcbiAgICAgIHZlcnNpb25lZDogdHJ1ZSxcbiAgICB9KTtcblxuICAgIC8vIPCflLkgQ29kZUJ1aWxkIHJvbGVcbiAgICBjb25zdCBidWlsZFJvbGUgPSBuZXcgaWFtLlJvbGUodGhpcywgJ09ic2lkaWFuQnVpbGRSb2xlJywge1xuICAgICAgYXNzdW1lZEJ5OiBuZXcgaWFtLlNlcnZpY2VQcmluY2lwYWwoJ2NvZGVidWlsZC5hbWF6b25hd3MuY29tJyksXG4gICAgfSk7XG4gICAgYXJ0aWZhY3RCdWNrZXQuZ3JhbnRSZWFkV3JpdGUoYnVpbGRSb2xlKTtcblxuICAgIC8vIPCflLkgQ29kZUJ1aWxkIHByb2plY3QgdG8gYnVpbGQgT2JzaWRpYW5cbiAgICBjb25zdCBidWlsZFByb2plY3QgPSBuZXcgY29kZWJ1aWxkLlByb2plY3QodGhpcywgJ09ic2lkaWFuQnVpbGRQcm9qZWN0Jywge1xuICAgICAgcm9sZTogYnVpbGRSb2xlLFxuICAgICAgc291cmNlOiBjb2RlYnVpbGQuU291cmNlLmdpdEh1Yih7XG4gICAgICAgIG93bmVyOiAnTGF5ci1MYWJzJyxcbiAgICAgICAgcmVwbzogJ2hvdXJnbGFzcy1tb25vcmVwbycsXG4gICAgICB9KSxcbiAgICAgIGVudmlyb25tZW50OiB7XG4gICAgICAgIGJ1aWxkSW1hZ2U6IGNvZGVidWlsZC5MaW51eEJ1aWxkSW1hZ2UuU1RBTkRBUkRfN18wLFxuICAgICAgICBjb21wdXRlVHlwZTogY29kZWJ1aWxkLkNvbXB1dGVUeXBlLk1FRElVTSxcbiAgICAgIH0sXG4gICAgICBhcnRpZmFjdHM6IGNvZGVidWlsZC5BcnRpZmFjdHMuczMoe1xuICAgICAgICBidWNrZXQ6IGFydGlmYWN0QnVja2V0LFxuICAgICAgICBpbmNsdWRlQnVpbGRJZDogZmFsc2UsXG4gICAgICAgIHBhY2thZ2VaaXA6IGZhbHNlLFxuICAgICAgICBwYXRoOiAnYmluYXJpZXMnLFxuICAgICAgICBuYW1lOiAnb2JzaWRpYW4tbGludXgtYW1kNjQnLFxuICAgICAgfSksXG4gICAgICBidWlsZFNwZWM6IGNvZGVidWlsZC5CdWlsZFNwZWMuZnJvbU9iamVjdCh7XG4gICAgICAgIHZlcnNpb246ICcwLjInLFxuICAgICAgICBwaGFzZXM6IHtcbiAgICAgICAgICBpbnN0YWxsOiB7XG4gICAgICAgICAgICAncnVudGltZS12ZXJzaW9ucyc6IHtcbiAgICAgICAgICAgICAgZ29sYW5nOiAnMS4yMScsXG4gICAgICAgICAgICB9LFxuICAgICAgICAgICAgY29tbWFuZHM6IFtcbiAgICAgICAgICAgICAgJ2FwdC1nZXQgdXBkYXRlIC15JyxcbiAgICAgICAgICAgICAgJ2FwdC1nZXQgaW5zdGFsbCAteSBwcm90b2J1Zi1jb21waWxlcicsXG4gICAgICAgICAgICAgICdnbyBpbnN0YWxsIGdvb2dsZS5nb2xhbmcub3JnL3Byb3RvYnVmL2NtZC9wcm90b2MtZ2VuLWdvQGxhdGVzdCcsXG4gICAgICAgICAgICAgICdnbyBpbnN0YWxsIGdvb2dsZS5nb2xhbmcub3JnL2dycGMvY21kL3Byb3RvYy1nZW4tZ28tZ3JwY0BsYXRlc3QnLFxuICAgICAgICAgICAgXSxcbiAgICAgICAgICB9LFxuICAgICAgICAgIGJ1aWxkOiB7XG4gICAgICAgICAgICBjb21tYW5kczogW1xuICAgICAgICAgICAgICAnY2Qgb2JzaWRpYW4nLFxuICAgICAgICAgICAgICAnbWFrZSBwcm90bycsXG4gICAgICAgICAgICAgICdHT09TPWxpbnV4IEdPQVJDSD1hbWQ2NCBDR09fRU5BQkxFRD0wIGdvIGJ1aWxkIC1vIG9ic2lkaWFuLWxpbnV4LWFtZDY0IC4vY21kL29ic2lkaWFuJyxcbiAgICAgICAgICAgIF0sXG4gICAgICAgICAgfSxcbiAgICAgICAgfSxcbiAgICAgICAgYXJ0aWZhY3RzOiB7XG4gICAgICAgICAgZmlsZXM6IFsnb2JzaWRpYW4vb2JzaWRpYW4tbGludXgtYW1kNjQnXSxcbiAgICAgICAgICAnZGlzY2FyZC1wYXRocyc6ICd5ZXMnLFxuICAgICAgICB9LFxuICAgICAgfSksXG4gICAgfSk7XG5cbiAgICAvLyDwn5S5IFZQQyBmb3IgRUMyIGluc3RhbmNlc1xuICAgIGNvbnN0IHZwYyA9IG5ldyBlYzIuVnBjKHRoaXMsICdPYnNpZGlhblZQQycsIHtcbiAgICAgIG1heEF6czogMixcbiAgICAgIG5hdEdhdGV3YXlzOiAxLFxuICAgIH0pO1xuXG4gICAgLy8g8J+UuSBFQzIgcm9sZSB3aXRoIHBlcm1pc3Npb25zXG4gICAgY29uc3QgZWMyUm9sZSA9IG5ldyBpYW0uUm9sZSh0aGlzLCAnT2JzaWRpYW5JbnN0YW5jZVJvbGUnLCB7XG4gICAgICBhc3N1bWVkQnk6IG5ldyBpYW0uU2VydmljZVByaW5jaXBhbCgnZWMyLmFtYXpvbmF3cy5jb20nKSxcbiAgICAgIG1hbmFnZWRQb2xpY2llczogW1xuICAgICAgICBpYW0uTWFuYWdlZFBvbGljeS5mcm9tQXdzTWFuYWdlZFBvbGljeU5hbWUoJ0FtYXpvblNTTU1hbmFnZWRJbnN0YW5jZUNvcmUnKSxcbiAgICAgIF0sXG4gICAgfSk7XG4gICAgXG4gICAgLy8gR3JhbnQgUzMgcmVhZCBhY2Nlc3NcbiAgICBhcnRpZmFjdEJ1Y2tldC5ncmFudFJlYWQoZWMyUm9sZSk7XG4gICAgXG4gICAgLy8gQWRkIEVDUiBwZXJtaXNzaW9ucyBmb3IgcHVsbGluZyBjb250YWluZXIgaW1hZ2VzXG4gICAgZWMyUm9sZS5hZGRUb1BvbGljeShuZXcgaWFtLlBvbGljeVN0YXRlbWVudCh7XG4gICAgICBlZmZlY3Q6IGlhbS5FZmZlY3QuQUxMT1csXG4gICAgICBhY3Rpb25zOiBbXG4gICAgICAgICdlY3I6R2V0QXV0aG9yaXphdGlvblRva2VuJyxcbiAgICAgICAgJ2VjcjpCYXRjaENoZWNrTGF5ZXJBdmFpbGFiaWxpdHknLFxuICAgICAgICAnZWNyOkdldERvd25sb2FkVXJsRm9yTGF5ZXInLFxuICAgICAgICAnZWNyOkJhdGNoR2V0SW1hZ2UnLFxuICAgICAgXSxcbiAgICAgIHJlc291cmNlczogWycqJ10sXG4gICAgfSkpO1xuXG4gICAgLy8g8J+UuSBTZWN1cml0eSBncm91cFxuICAgIGNvbnN0IHNlY3VyaXR5R3JvdXAgPSBuZXcgZWMyLlNlY3VyaXR5R3JvdXAodGhpcywgJ09ic2lkaWFuU2VjdXJpdHlHcm91cCcsIHtcbiAgICAgIHZwYyxcbiAgICAgIGRlc2NyaXB0aW9uOiAnU2VjdXJpdHkgZ3JvdXAgZm9yIE9ic2lkaWFuIGluc3RhbmNlcycsXG4gICAgfSk7XG4gICAgXG4gICAgc2VjdXJpdHlHcm91cC5hZGRJbmdyZXNzUnVsZShcbiAgICAgIGVjMi5QZWVyLmFueUlwdjQoKSxcbiAgICAgIGVjMi5Qb3J0LnRjcCg4MDgwKSxcbiAgICAgICdIVFRQIEFQSSdcbiAgICApO1xuICAgIFxuICAgIHNlY3VyaXR5R3JvdXAuYWRkSW5ncmVzc1J1bGUoXG4gICAgICBlYzIuUGVlci5hbnlJcHY0KCksXG4gICAgICBlYzIuUG9ydC50Y3AoOTA5MCksXG4gICAgICAnZ1JQQyBBUEknXG4gICAgKTtcblxuICAgIC8vIPCflLkgVXNlciBkYXRhIHNjcmlwdFxuICAgIGNvbnN0IHVzZXJEYXRhID0gZWMyLlVzZXJEYXRhLmZvckxpbnV4KCk7XG4gICAgdXNlckRhdGEuYWRkQ29tbWFuZHMoXG4gICAgICAvLyBJbnN0YWxsIERvY2tlclxuICAgICAgJ3l1bSB1cGRhdGUgLXknLFxuICAgICAgJ3l1bSBpbnN0YWxsIC15IGRvY2tlciBhd3MtY2xpJyxcbiAgICAgICdzZXJ2aWNlIGRvY2tlciBzdGFydCcsXG4gICAgICAndXNlcm1vZCAtYSAtRyBkb2NrZXIgZWMyLXVzZXInLFxuICAgICAgXG4gICAgICAvLyBTZXR1cCBPYnNpZGlhblxuICAgICAgJ21rZGlyIC1wIC9vcHQvb2JzaWRpYW4nLFxuICAgICAgXG4gICAgICAvLyBDcmVhdGUgZGVmYXVsdCBjb25maWdcbiAgICAgICdjYXQgPiAvb3B0L29ic2lkaWFuL2NvbmZpZy55YW1sIDw8IEVPRicsXG4gICAgICB0aGlzLmdlbmVyYXRlTWluaW1hbENvbmZpZygpLFxuICAgICAgJ0VPRicsXG4gICAgICBcbiAgICAgIC8vIERvd25sb2FkIE9ic2lkaWFuIGJpbmFyeSBmcm9tIFMzXG4gICAgICBgYXdzIHMzIGNwIHMzOi8vJHthcnRpZmFjdEJ1Y2tldC5idWNrZXROYW1lfS9iaW5hcmllcy9vYnNpZGlhbi1saW51eC1hbWQ2NCAvb3B0L29ic2lkaWFuL29ic2lkaWFuIHx8IGVjaG8gXCJCaW5hcnkgbm90IHlldCBhdmFpbGFibGVcImAsXG4gICAgICAnY2htb2QgK3ggL29wdC9vYnNpZGlhbi9vYnNpZGlhbiB8fCB0cnVlJyxcbiAgICAgIFxuICAgICAgLy8gQ3JlYXRlIHN5c3RlbWQgc2VydmljZVxuICAgICAgJ2NhdCA+IC9ldGMvc3lzdGVtZC9zeXN0ZW0vb2JzaWRpYW4uc2VydmljZSA8PCBFT0YnLFxuICAgICAgJ1tVbml0XScsXG4gICAgICAnRGVzY3JpcHRpb249T2JzaWRpYW4gR2F0ZXdheScsXG4gICAgICAnQWZ0ZXI9ZG9ja2VyLnNlcnZpY2UnLFxuICAgICAgJycsXG4gICAgICAnW1NlcnZpY2VdJyxcbiAgICAgICdUeXBlPXNpbXBsZScsXG4gICAgICAnRXhlY1N0YXJ0PS9vcHQvb2JzaWRpYW4vb2JzaWRpYW4gLS1jb25maWc9L29wdC9vYnNpZGlhbi9jb25maWcueWFtbCcsXG4gICAgICAnUmVzdGFydD1hbHdheXMnLFxuICAgICAgJ1N0YW5kYXJkT3V0cHV0PWpvdXJuYWwnLFxuICAgICAgJycsXG4gICAgICAnW0luc3RhbGxdJyxcbiAgICAgICdXYW50ZWRCeT1tdWx0aS11c2VyLnRhcmdldCcsXG4gICAgICAnRU9GJyxcbiAgICAgIFxuICAgICAgJ3N5c3RlbWN0bCBkYWVtb24tcmVsb2FkJyxcbiAgICAgICdzeXN0ZW1jdGwgZW5hYmxlIG9ic2lkaWFuJyxcbiAgICAgICcjIFN0YXJ0IG9ubHkgaWYgYmluYXJ5IGV4aXN0cycsXG4gICAgICAnWyAtZiAvb3B0L29ic2lkaWFuL29ic2lkaWFuIF0gJiYgc3lzdGVtY3RsIHN0YXJ0IG9ic2lkaWFuIHx8IGVjaG8gXCJXYWl0aW5nIGZvciBiaW5hcnlcIidcbiAgICApO1xuXG4gICAgLy8g8J+UuSBMYXVuY2ggdGVtcGxhdGVcbiAgICBjb25zdCBsYXVuY2hUZW1wbGF0ZSA9IG5ldyBlYzIuTGF1bmNoVGVtcGxhdGUodGhpcywgJ09ic2lkaWFuTGF1bmNoVGVtcGxhdGUnLCB7XG4gICAgICBpbnN0YW5jZVR5cGU6IGVjMi5JbnN0YW5jZVR5cGUub2YoZWMyLkluc3RhbmNlQ2xhc3MuTTUsIGVjMi5JbnN0YW5jZVNpemUuTEFSR0UpLFxuICAgICAgbWFjaGluZUltYWdlOiBlYzIuTWFjaGluZUltYWdlLmxhdGVzdEFtYXpvbkxpbnV4MigpLFxuICAgICAgdXNlckRhdGEsXG4gICAgICByb2xlOiBlYzJSb2xlLFxuICAgICAgc2VjdXJpdHlHcm91cCxcbiAgICB9KTtcblxuICAgIC8vIPCflLkgQXV0byBzY2FsaW5nIGdyb3VwXG4gICAgY29uc3QgYXNnID0gbmV3IGF1dG9zY2FsaW5nLkF1dG9TY2FsaW5nR3JvdXAodGhpcywgJ09ic2lkaWFuQVNHJywge1xuICAgICAgdnBjLFxuICAgICAgbGF1bmNoVGVtcGxhdGUsXG4gICAgICBtaW5DYXBhY2l0eTogMSxcbiAgICAgIG1heENhcGFjaXR5OiAxLFxuICAgIH0pO1xuXG4gICAgLy8g8J+UuSBPdXRwdXRzXG4gICAgbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ0J1Y2tldE5hbWUnLCB7XG4gICAgICB2YWx1ZTogYXJ0aWZhY3RCdWNrZXQuYnVja2V0TmFtZSxcbiAgICAgIGRlc2NyaXB0aW9uOiAnUzMgYnVja2V0IGNvbnRhaW5pbmcgT2JzaWRpYW4gYmluYXJpZXMnLFxuICAgIH0pO1xuXG4gICAgbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ0J1aWxkUHJvamVjdE5hbWUnLCB7XG4gICAgICB2YWx1ZTogYnVpbGRQcm9qZWN0LnByb2plY3ROYW1lLFxuICAgICAgZGVzY3JpcHRpb246ICdDb2RlQnVpbGQgcHJvamVjdCBuYW1lJyxcbiAgICB9KTtcblxuICAgIG5ldyBjZGsuQ2ZuT3V0cHV0KHRoaXMsICdCdWlsZENvbW1hbmQnLCB7XG4gICAgICB2YWx1ZTogYGF3cyBjb2RlYnVpbGQgc3RhcnQtYnVpbGQgLS1wcm9qZWN0LW5hbWUgJHtidWlsZFByb2plY3QucHJvamVjdE5hbWV9YCxcbiAgICAgIGRlc2NyaXB0aW9uOiAnQ29tbWFuZCB0byB0cmlnZ2VyIGEgYnVpbGQnLFxuICAgIH0pO1xuXG4gICAgbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ0luc3RhbmNlVXBkYXRlQ29tbWFuZCcsIHtcbiAgICAgIHZhbHVlOiBgYXdzIHNzbSBzZW5kLWNvbW1hbmQgLS1kb2N1bWVudC1uYW1lIFwiQVdTLVJ1blNoZWxsU2NyaXB0XCIgLS10YXJnZXRzIFwiS2V5PXRhZzphd3M6YXV0b3NjYWxpbmc6Z3JvdXBOYW1lLFZhbHVlcz0ke2FzZy5hdXRvU2NhbGluZ0dyb3VwTmFtZX1cIiAtLXBhcmFtZXRlcnMgJ2NvbW1hbmRzPVtcImF3cyBzMyBjcCBzMzovLyR7YXJ0aWZhY3RCdWNrZXQuYnVja2V0TmFtZX0vYmluYXJpZXMvb2JzaWRpYW4tbGludXgtYW1kNjQgL29wdC9vYnNpZGlhbi9vYnNpZGlhbiAmJiBjaG1vZCAreCAvb3B0L29ic2lkaWFuL29ic2lkaWFuICYmIHN5c3RlbWN0bCByZXN0YXJ0IG9ic2lkaWFuXCJdJ2AsXG4gICAgICBkZXNjcmlwdGlvbjogJ0NvbW1hbmQgdG8gdXBkYXRlIE9ic2lkaWFuIG9uIHJ1bm5pbmcgaW5zdGFuY2VzJyxcbiAgICB9KTtcbiAgfVxuXG4gIHByaXZhdGUgZ2VuZXJhdGVNaW5pbWFsQ29uZmlnKCk6IHN0cmluZyB7XG4gICAgcmV0dXJuIGBzZXJ2ZXI6XG4gIHBvcnQ6IDgwODBcbiAgZ3JwY1BvcnQ6IDkwOTBcblxub3JjaGVzdHJhdG9yOlxuICByZXNvdXJjZXM6XG4gICAgbWF4Q3B1OiBcIjRcIlxuICAgIG1heE1lbW9yeTogXCI4R2lcIiBcbiAgICBtYXhDb250YWluZXJzOiAxMFxuXG5yZWdpc3RyeTpcbiAgcmVnaXN0cmllczpcbiAgICAtIG5hbWU6IFwiZWNyXCJcbiAgICAgIHR5cGU6IFwiYXdzLWVjclwiXG4gICAgICByZWdpb246IFwiJHt0aGlzLnJlZ2lvbn1cIlxuICAgICAgY3JlZGVudGlhbFNvdXJjZTogXCJpYW0tcm9sZVwiXG5cbnByb3h5OlxuICBiYWNrZW5kczogW11cblxubG9nZ2luZzpcbiAgbGV2ZWw6IFwiaW5mb1wiYDtcbiAgfVxufSJdfQ==