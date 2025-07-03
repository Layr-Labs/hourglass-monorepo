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
                branchOrRef: 'obsidian', // Your test branch
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
                            // Use specific versions compatible with Go 1.21
                            'go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.32.0',
                            'go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0',
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
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoib2JzaWRpYW4tc2ltcGxlLXN0YWNrLmpzIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsib2JzaWRpYW4tc2ltcGxlLXN0YWNrLnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiI7OztBQUFBLG1DQUFtQztBQUNuQyx1REFBdUQ7QUFDdkQseUNBQXlDO0FBQ3pDLDJDQUEyQztBQUMzQywyQ0FBMkM7QUFDM0MsMkRBQTJEO0FBRzNELE1BQWEsbUJBQW9CLFNBQVEsR0FBRyxDQUFDLEtBQUs7SUFDaEQsWUFBWSxLQUFnQixFQUFFLEVBQVUsRUFBRSxLQUFzQjtRQUM5RCxLQUFLLENBQUMsS0FBSyxFQUFFLEVBQUUsRUFBRSxLQUFLLENBQUMsQ0FBQztRQUV4QiwyQ0FBMkM7UUFDM0MsTUFBTSxjQUFjLEdBQUcsSUFBSSxFQUFFLENBQUMsTUFBTSxDQUFDLElBQUksRUFBRSxtQkFBbUIsRUFBRTtZQUM5RCxVQUFVLEVBQUUsc0JBQXNCLElBQUksQ0FBQyxPQUFPLElBQUksSUFBSSxDQUFDLE1BQU0sRUFBRTtZQUMvRCxhQUFhLEVBQUUsR0FBRyxDQUFDLGFBQWEsQ0FBQyxNQUFNO1lBQ3ZDLFNBQVMsRUFBRSxJQUFJO1NBQ2hCLENBQUMsQ0FBQztRQUVILG9CQUFvQjtRQUNwQixNQUFNLFNBQVMsR0FBRyxJQUFJLEdBQUcsQ0FBQyxJQUFJLENBQUMsSUFBSSxFQUFFLG1CQUFtQixFQUFFO1lBQ3hELFNBQVMsRUFBRSxJQUFJLEdBQUcsQ0FBQyxnQkFBZ0IsQ0FBQyx5QkFBeUIsQ0FBQztTQUMvRCxDQUFDLENBQUM7UUFDSCxjQUFjLENBQUMsY0FBYyxDQUFDLFNBQVMsQ0FBQyxDQUFDO1FBRXpDLHlDQUF5QztRQUN6QyxNQUFNLFlBQVksR0FBRyxJQUFJLFNBQVMsQ0FBQyxPQUFPLENBQUMsSUFBSSxFQUFFLHNCQUFzQixFQUFFO1lBQ3ZFLElBQUksRUFBRSxTQUFTO1lBQ2YsTUFBTSxFQUFFLFNBQVMsQ0FBQyxNQUFNLENBQUMsTUFBTSxDQUFDO2dCQUM5QixL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QUUsYUFBYSxFQUFFO1lBQ2hFLEdBQUc7WUFDSCxjQUFjO1lBQ2QsV0FBVyxFQUFFLENBQUM7WUFDZCxXQUFXLEVBQUUsQ0FBQztTQUNmLENBQUMsQ0FBQztRQUVILGFBQWE7UUFDYixJQUFJLEdBQUcsQ0FBQyxTQUFTLENBQUMsSUFBSSxFQUFFLFlBQVksRUFBRTtZQUNwQyxLQUFLLEVBQUUsY0FBYyxDQUFDLFVBQVU7WUFDaEMsV0FBVyxFQUFFLHdDQUF3QztTQUN0RCxDQUFDLENBQUM7UUFFSCxJQUFJLEdBQUcsQ0FBQyxTQUFTLENBQUMsSUFBSSxFQUFFLGtCQUFrQixFQUFFO1lBQzFDLEtBQUssRUFBRSxZQUFZLENBQUMsV0FBVztZQUMvQixXQUFXLEVBQUUsd0JBQXdCO1NBQ3RDLENBQUMsQ0FBQztRQUVILElBQUksR0FBRyxDQUFDLFNBQVMsQ0FBQyxJQUFJLEVBQUUsY0FBYyxFQUFFO1lBQ3RDLEtBQUssRUFBRSw0Q0FBNEMsWUFBWSxDQUFDLFdBQVcsRUFBRTtZQUM3RSxXQUFXLEVBQUUsNEJBQTRCO1NBQzFDLENBQUMsQ0FBQztRQUVILElBQUksR0FBRyxDQUFDLFNBQVMsQ0FBQyxJQUFJLEVBQUUsdUJBQXVCLEVBQUU7WUFDL0MsS0FBSyxFQUFFLGlIQUFpSCxHQUFHLENBQUMsb0JBQW9CLDZDQUE2QyxjQUFjLENBQUMsVUFBVSwySEFBMkg7WUFDalYsV0FBVyxFQUFFLGlEQUFpRDtTQUMvRCxDQUFDLENBQUM7SUFDTCxDQUFDO0lBRU8scUJBQXFCO1FBQzNCLE9BQU87Ozs7Ozs7Ozs7Ozs7O2lCQWNNLElBQUksQ0FBQyxNQUFNOzs7Ozs7O2dCQU9aLENBQUM7SUFDZixDQUFDO0NBQ0Y7QUEzTkQsa0RBMk5DIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0ICogYXMgY2RrIGZyb20gJ2F3cy1jZGstbGliJztcbmltcG9ydCAqIGFzIGNvZGVidWlsZCBmcm9tICdhd3MtY2RrLWxpYi9hd3MtY29kZWJ1aWxkJztcbmltcG9ydCAqIGFzIHMzIGZyb20gJ2F3cy1jZGstbGliL2F3cy1zMyc7XG5pbXBvcnQgKiBhcyBpYW0gZnJvbSAnYXdzLWNkay1saWIvYXdzLWlhbSc7XG5pbXBvcnQgKiBhcyBlYzIgZnJvbSAnYXdzLWNkay1saWIvYXdzLWVjMic7XG5pbXBvcnQgKiBhcyBhdXRvc2NhbGluZyBmcm9tICdhd3MtY2RrLWxpYi9hd3MtYXV0b3NjYWxpbmcnO1xuaW1wb3J0IHsgQ29uc3RydWN0IH0gZnJvbSAnY29uc3RydWN0cyc7XG5cbmV4cG9ydCBjbGFzcyBPYnNpZGlhblNpbXBsZVN0YWNrIGV4dGVuZHMgY2RrLlN0YWNrIHtcbiAgY29uc3RydWN0b3Ioc2NvcGU6IENvbnN0cnVjdCwgaWQ6IHN0cmluZywgcHJvcHM/OiBjZGsuU3RhY2tQcm9wcykge1xuICAgIHN1cGVyKHNjb3BlLCBpZCwgcHJvcHMpO1xuXG4gICAgLy8g8J+UuSBBcnRpZmFjdCBidWNrZXQgZm9yIE9ic2lkaWFuIGJpbmFyaWVzXG4gICAgY29uc3QgYXJ0aWZhY3RCdWNrZXQgPSBuZXcgczMuQnVja2V0KHRoaXMsICdPYnNpZGlhbkFydGlmYWN0cycsIHtcbiAgICAgIGJ1Y2tldE5hbWU6IGBvYnNpZGlhbi1hcnRpZmFjdHMtJHt0aGlzLmFjY291bnR9LSR7dGhpcy5yZWdpb259YCxcbiAgICAgIHJlbW92YWxQb2xpY3k6IGNkay5SZW1vdmFsUG9saWN5LlJFVEFJTixcbiAgICAgIHZlcnNpb25lZDogdHJ1ZSxcbiAgICB9KTtcblxuICAgIC8vIPCflLkgQ29kZUJ1aWxkIHJvbGVcbiAgICBjb25zdCBidWlsZFJvbGUgPSBuZXcgaWFtLlJvbGUodGhpcywgJ09ic2lkaWFuQnVpbGRSb2xlJywge1xuICAgICAgYXNzdW1lZEJ5OiBuZXcgaWFtLlNlcnZpY2VQcmluY2lwYWwoJ2NvZGVidWlsZC5hbWF6b25hd3MuY29tJyksXG4gICAgfSk7XG4gICAgYXJ0aWZhY3RCdWNrZXQuZ3JhbnRSZWFkV3JpdGUoYnVpbGRSb2xlKTtcblxuICAgIC8vIPCflLkgQ29kZUJ1aWxkIHByb2plY3QgdG8gYnVpbGQgT2JzaWRpYW5cbiAgICBjb25zdCBidWlsZFByb2plY3QgPSBuZXcgY29kZWJ1aWxkLlByb2plY3QodGhpcywgJ09ic2lkaWFuQnVpbGRQcm9qZWN0Jywge1xuICAgICAgcm9sZTogYnVpbGRSb2xlLFxuICAgICAgc291cmNlOiBjb2RlYnVpbGQuU291cmNlLmdpdEh1Yih7XG4gICAgICAgIG93bmVyOiAnTGF5ci1MYWJzJyxcbiAgICAgICAgcmVwbzogJ2hvdXJnbGFzcy1tb25vcmVwbycsXG4gICAgICAgIGJyYW5jaE9yUmVmOiAnb2JzaWRpYW4nLCAvLyBZb3VyIHRlc3QgYnJhbmNoXG4gICAgICB9KSxcbiAgICAgIGVudmlyb25tZW50OiB7XG4gICAgICAgIGJ1aWxkSW1hZ2U6IGNvZGVidWlsZC5MaW51eEJ1aWxkSW1hZ2UuU1RBTkRBUkRfN18wLFxuICAgICAgICBjb21wdXRlVHlwZTogY29kZWJ1aWxkLkNvbXB1dGVUeXBlLk1FRElVTSxcbiAgICAgIH0sXG4gICAgICBhcnRpZmFjdHM6IGNvZGVidWlsZC5BcnRpZmFjdHMuczMoe1xuICAgICAgICBidWNrZXQ6IGFydGlmYWN0QnVja2V0LFxuICAgICAgICBpbmNsdWRlQnVpbGRJZDogZmFsc2UsXG4gICAgICAgIHBhY2thZ2VaaXA6IGZhbHNlLFxuICAgICAgICBwYXRoOiAnYmluYXJpZXMnLFxuICAgICAgICBuYW1lOiAnb2JzaWRpYW4tbGludXgtYW1kNjQnLFxuICAgICAgfSksXG4gICAgICBidWlsZFNwZWM6IGNvZGVidWlsZC5CdWlsZFNwZWMuZnJvbU9iamVjdCh7XG4gICAgICAgIHZlcnNpb246ICcwLjInLFxuICAgICAgICBwaGFzZXM6IHtcbiAgICAgICAgICBpbnN0YWxsOiB7XG4gICAgICAgICAgICAncnVudGltZS12ZXJzaW9ucyc6IHtcbiAgICAgICAgICAgICAgZ29sYW5nOiAnMS4yMScsXG4gICAgICAgICAgICB9LFxuICAgICAgICAgICAgY29tbWFuZHM6IFtcbiAgICAgICAgICAgICAgJ2FwdC1nZXQgdXBkYXRlIC15JyxcbiAgICAgICAgICAgICAgJ2FwdC1nZXQgaW5zdGFsbCAteSBwcm90b2J1Zi1jb21waWxlcicsXG4gICAgICAgICAgICAgIC8vIFVzZSBzcGVjaWZpYyB2ZXJzaW9ucyBjb21wYXRpYmxlIHdpdGggR28gMS4yMVxuICAgICAgICAgICAgICAnZ28gaW5zdGFsbCBnb29nbGUuZ29sYW5nLm9yZy9wcm90b2J1Zi9jbWQvcHJvdG9jLWdlbi1nb0B2MS4zMi4wJyxcbiAgICAgICAgICAgICAgJ2dvIGluc3RhbGwgZ29vZ2xlLmdvbGFuZy5vcmcvZ3JwYy9jbWQvcHJvdG9jLWdlbi1nby1ncnBjQHYxLjMuMCcsXG4gICAgICAgICAgICBdLFxuICAgICAgICAgIH0sXG4gICAgICAgICAgYnVpbGQ6IHtcbiAgICAgICAgICAgIGNvbW1hbmRzOiBbXG4gICAgICAgICAgICAgICdjZCBvYnNpZGlhbicsXG4gICAgICAgICAgICAgICdtYWtlIHByb3RvJyxcbiAgICAgICAgICAgICAgJ0dPT1M9bGludXggR09BUkNIPWFtZDY0IENHT19FTkFCTEVEPTAgZ28gYnVpbGQgLW8gb2JzaWRpYW4tbGludXgtYW1kNjQgLi9jbWQvb2JzaWRpYW4nLFxuICAgICAgICAgICAgXSxcbiAgICAgICAgICB9LFxuICAgICAgICB9LFxuICAgICAgICBhcnRpZmFjdHM6IHtcbiAgICAgICAgICBmaWxlczogWydvYnNpZGlhbi9vYnNpZGlhbi1saW51eC1hbWQ2NCddLFxuICAgICAgICAgICdkaXNjYXJkLXBhdGhzJzogJ3llcycsXG4gICAgICAgIH0sXG4gICAgICB9KSxcbiAgICB9KTtcblxuICAgIC8vIPCflLkgVlBDIGZvciBFQzIgaW5zdGFuY2VzXG4gICAgY29uc3QgdnBjID0gbmV3IGVjMi5WcGModGhpcywgJ09ic2lkaWFuVlBDJywge1xuICAgICAgbWF4QXpzOiAyLFxuICAgICAgbmF0R2F0ZXdheXM6IDEsXG4gICAgfSk7XG5cbiAgICAvLyDwn5S5IEVDMiByb2xlIHdpdGggcGVybWlzc2lvbnNcbiAgICBjb25zdCBlYzJSb2xlID0gbmV3IGlhbS5Sb2xlKHRoaXMsICdPYnNpZGlhbkluc3RhbmNlUm9sZScsIHtcbiAgICAgIGFzc3VtZWRCeTogbmV3IGlhbS5TZXJ2aWNlUHJpbmNpcGFsKCdlYzIuYW1hem9uYXdzLmNvbScpLFxuICAgICAgbWFuYWdlZFBvbGljaWVzOiBbXG4gICAgICAgIGlhbS5NYW5hZ2VkUG9saWN5LmZyb21Bd3NNYW5hZ2VkUG9saWN5TmFtZSgnQW1hem9uU1NNTWFuYWdlZEluc3RhbmNlQ29yZScpLFxuICAgICAgXSxcbiAgICB9KTtcbiAgICBcbiAgICAvLyBHcmFudCBTMyByZWFkIGFjY2Vzc1xuICAgIGFydGlmYWN0QnVja2V0LmdyYW50UmVhZChlYzJSb2xlKTtcbiAgICBcbiAgICAvLyBBZGQgRUNSIHBlcm1pc3Npb25zIGZvciBwdWxsaW5nIGNvbnRhaW5lciBpbWFnZXNcbiAgICBlYzJSb2xlLmFkZFRvUG9saWN5KG5ldyBpYW0uUG9saWN5U3RhdGVtZW50KHtcbiAgICAgIGVmZmVjdDogaWFtLkVmZmVjdC5BTExPVyxcbiAgICAgIGFjdGlvbnM6IFtcbiAgICAgICAgJ2VjcjpHZXRBdXRob3JpemF0aW9uVG9rZW4nLFxuICAgICAgICAnZWNyOkJhdGNoQ2hlY2tMYXllckF2YWlsYWJpbGl0eScsXG4gICAgICAgICdlY3I6R2V0RG93bmxvYWRVcmxGb3JMYXllcicsXG4gICAgICAgICdlY3I6QmF0Y2hHZXRJbWFnZScsXG4gICAgICBdLFxuICAgICAgcmVzb3VyY2VzOiBbJyonXSxcbiAgICB9KSk7XG5cbiAgICAvLyDwn5S5IFNlY3VyaXR5IGdyb3VwXG4gICAgY29uc3Qgc2VjdXJpdHlHcm91cCA9IG5ldyBlYzIuU2VjdXJpdHlHcm91cCh0aGlzLCAnT2JzaWRpYW5TZWN1cml0eUdyb3VwJywge1xuICAgICAgdnBjLFxuICAgICAgZGVzY3JpcHRpb246ICdTZWN1cml0eSBncm91cCBmb3IgT2JzaWRpYW4gaW5zdGFuY2VzJyxcbiAgICB9KTtcbiAgICBcbiAgICBzZWN1cml0eUdyb3VwLmFkZEluZ3Jlc3NSdWxlKFxuICAgICAgZWMyLlBlZXIuYW55SXB2NCgpLFxuICAgICAgZWMyLlBvcnQudGNwKDgwODApLFxuICAgICAgJ0hUVFAgQVBJJ1xuICAgICk7XG4gICAgXG4gICAgc2VjdXJpdHlHcm91cC5hZGRJbmdyZXNzUnVsZShcbiAgICAgIGVjMi5QZWVyLmFueUlwdjQoKSxcbiAgICAgIGVjMi5Qb3J0LnRjcCg5MDkwKSxcbiAgICAgICdnUlBDIEFQSSdcbiAgICApO1xuXG4gICAgLy8g8J+UuSBVc2VyIGRhdGEgc2NyaXB0XG4gICAgY29uc3QgdXNlckRhdGEgPSBlYzIuVXNlckRhdGEuZm9yTGludXgoKTtcbiAgICB1c2VyRGF0YS5hZGRDb21tYW5kcyhcbiAgICAgIC8vIEluc3RhbGwgRG9ja2VyXG4gICAgICAneXVtIHVwZGF0ZSAteScsXG4gICAgICAneXVtIGluc3RhbGwgLXkgZG9ja2VyIGF3cy1jbGknLFxuICAgICAgJ3NlcnZpY2UgZG9ja2VyIHN0YXJ0JyxcbiAgICAgICd1c2VybW9kIC1hIC1HIGRvY2tlciBlYzItdXNlcicsXG4gICAgICBcbiAgICAgIC8vIFNldHVwIE9ic2lkaWFuXG4gICAgICAnbWtkaXIgLXAgL29wdC9vYnNpZGlhbicsXG4gICAgICBcbiAgICAgIC8vIENyZWF0ZSBkZWZhdWx0IGNvbmZpZ1xuICAgICAgJ2NhdCA+IC9vcHQvb2JzaWRpYW4vY29uZmlnLnlhbWwgPDwgRU9GJyxcbiAgICAgIHRoaXMuZ2VuZXJhdGVNaW5pbWFsQ29uZmlnKCksXG4gICAgICAnRU9GJyxcbiAgICAgIFxuICAgICAgLy8gRG93bmxvYWQgT2JzaWRpYW4gYmluYXJ5IGZyb20gUzNcbiAgICAgIGBhd3MgczMgY3AgczM6Ly8ke2FydGlmYWN0QnVja2V0LmJ1Y2tldE5hbWV9L2JpbmFyaWVzL29ic2lkaWFuLWxpbnV4LWFtZDY0IC9vcHQvb2JzaWRpYW4vb2JzaWRpYW4gfHwgZWNobyBcIkJpbmFyeSBub3QgeWV0IGF2YWlsYWJsZVwiYCxcbiAgICAgICdjaG1vZCAreCAvb3B0L29ic2lkaWFuL29ic2lkaWFuIHx8IHRydWUnLFxuICAgICAgXG4gICAgICAvLyBDcmVhdGUgc3lzdGVtZCBzZXJ2aWNlXG4gICAgICAnY2F0ID4gL2V0Yy9zeXN0ZW1kL3N5c3RlbS9vYnNpZGlhbi5zZXJ2aWNlIDw8IEVPRicsXG4gICAgICAnW1VuaXRdJyxcbiAgICAgICdEZXNjcmlwdGlvbj1PYnNpZGlhbiBHYXRld2F5JyxcbiAgICAgICdBZnRlcj1kb2NrZXIuc2VydmljZScsXG4gICAgICAnJyxcbiAgICAgICdbU2VydmljZV0nLFxuICAgICAgJ1R5cGU9c2ltcGxlJyxcbiAgICAgICdFeGVjU3RhcnQ9L29wdC9vYnNpZGlhbi9vYnNpZGlhbiAtLWNvbmZpZz0vb3B0L29ic2lkaWFuL2NvbmZpZy55YW1sJyxcbiAgICAgICdSZXN0YXJ0PWFsd2F5cycsXG4gICAgICAnU3RhbmRhcmRPdXRwdXQ9am91cm5hbCcsXG4gICAgICAnJyxcbiAgICAgICdbSW5zdGFsbF0nLFxuICAgICAgJ1dhbnRlZEJ5PW11bHRpLXVzZXIudGFyZ2V0JyxcbiAgICAgICdFT0YnLFxuICAgICAgXG4gICAgICAnc3lzdGVtY3RsIGRhZW1vbi1yZWxvYWQnLFxuICAgICAgJ3N5c3RlbWN0bCBlbmFibGUgb2JzaWRpYW4nLFxuICAgICAgJyMgU3RhcnQgb25seSBpZiBiaW5hcnkgZXhpc3RzJyxcbiAgICAgICdbIC1mIC9vcHQvb2JzaWRpYW4vb2JzaWRpYW4gXSAmJiBzeXN0ZW1jdGwgc3RhcnQgb2JzaWRpYW4gfHwgZWNobyBcIldhaXRpbmcgZm9yIGJpbmFyeVwiJ1xuICAgICk7XG5cbiAgICAvLyDwn5S5IExhdW5jaCB0ZW1wbGF0ZVxuICAgIGNvbnN0IGxhdW5jaFRlbXBsYXRlID0gbmV3IGVjMi5MYXVuY2hUZW1wbGF0ZSh0aGlzLCAnT2JzaWRpYW5MYXVuY2hUZW1wbGF0ZScsIHtcbiAgICAgIGluc3RhbmNlVHlwZTogZWMyLkluc3RhbmNlVHlwZS5vZihlYzIuSW5zdGFuY2VDbGFzcy5NNSwgZWMyLkluc3RhbmNlU2l6ZS5MQVJHRSksXG4gICAgICBtYWNoaW5lSW1hZ2U6IGVjMi5NYWNoaW5lSW1hZ2UubGF0ZXN0QW1hem9uTGludXgyKCksXG4gICAgICB1c2VyRGF0YSxcbiAgICAgIHJvbGU6IGVjMlJvbGUsXG4gICAgICBzZWN1cml0eUdyb3VwLFxuICAgIH0pO1xuXG4gICAgLy8g8J+UuSBBdXRvIHNjYWxpbmcgZ3JvdXBcbiAgICBjb25zdCBhc2cgPSBuZXcgYXV0b3NjYWxpbmcuQXV0b1NjYWxpbmdHcm91cCh0aGlzLCAnT2JzaWRpYW5BU0cnLCB7XG4gICAgICB2cGMsXG4gICAgICBsYXVuY2hUZW1wbGF0ZSxcbiAgICAgIG1pbkNhcGFjaXR5OiAxLFxuICAgICAgbWF4Q2FwYWNpdHk6IDEsXG4gICAgfSk7XG5cbiAgICAvLyDwn5S5IE91dHB1dHNcbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnQnVja2V0TmFtZScsIHtcbiAgICAgIHZhbHVlOiBhcnRpZmFjdEJ1Y2tldC5idWNrZXROYW1lLFxuICAgICAgZGVzY3JpcHRpb246ICdTMyBidWNrZXQgY29udGFpbmluZyBPYnNpZGlhbiBiaW5hcmllcycsXG4gICAgfSk7XG5cbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnQnVpbGRQcm9qZWN0TmFtZScsIHtcbiAgICAgIHZhbHVlOiBidWlsZFByb2plY3QucHJvamVjdE5hbWUsXG4gICAgICBkZXNjcmlwdGlvbjogJ0NvZGVCdWlsZCBwcm9qZWN0IG5hbWUnLFxuICAgIH0pO1xuXG4gICAgbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ0J1aWxkQ29tbWFuZCcsIHtcbiAgICAgIHZhbHVlOiBgYXdzIGNvZGVidWlsZCBzdGFydC1idWlsZCAtLXByb2plY3QtbmFtZSAke2J1aWxkUHJvamVjdC5wcm9qZWN0TmFtZX1gLFxuICAgICAgZGVzY3JpcHRpb246ICdDb21tYW5kIHRvIHRyaWdnZXIgYSBidWlsZCcsXG4gICAgfSk7XG5cbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnSW5zdGFuY2VVcGRhdGVDb21tYW5kJywge1xuICAgICAgdmFsdWU6IGBhd3Mgc3NtIHNlbmQtY29tbWFuZCAtLWRvY3VtZW50LW5hbWUgXCJBV1MtUnVuU2hlbGxTY3JpcHRcIiAtLXRhcmdldHMgXCJLZXk9dGFnOmF3czphdXRvc2NhbGluZzpncm91cE5hbWUsVmFsdWVzPSR7YXNnLmF1dG9TY2FsaW5nR3JvdXBOYW1lfVwiIC0tcGFyYW1ldGVycyAnY29tbWFuZHM9W1wiYXdzIHMzIGNwIHMzOi8vJHthcnRpZmFjdEJ1Y2tldC5idWNrZXROYW1lfS9iaW5hcmllcy9vYnNpZGlhbi1saW51eC1hbWQ2NCAvb3B0L29ic2lkaWFuL29ic2lkaWFuICYmIGNobW9kICt4IC9vcHQvb2JzaWRpYW4vb2JzaWRpYW4gJiYgc3lzdGVtY3RsIHJlc3RhcnQgb2JzaWRpYW5cIl0nYCxcbiAgICAgIGRlc2NyaXB0aW9uOiAnQ29tbWFuZCB0byB1cGRhdGUgT2JzaWRpYW4gb24gcnVubmluZyBpbnN0YW5jZXMnLFxuICAgIH0pO1xuICB9XG5cbiAgcHJpdmF0ZSBnZW5lcmF0ZU1pbmltYWxDb25maWcoKTogc3RyaW5nIHtcbiAgICByZXR1cm4gYHNlcnZlcjpcbiAgcG9ydDogODA4MFxuICBncnBjUG9ydDogOTA5MFxuXG5vcmNoZXN0cmF0b3I6XG4gIHJlc291cmNlczpcbiAgICBtYXhDcHU6IFwiNFwiXG4gICAgbWF4TWVtb3J5OiBcIjhHaVwiIFxuICAgIG1heENvbnRhaW5lcnM6IDEwXG5cbnJlZ2lzdHJ5OlxuICByZWdpc3RyaWVzOlxuICAgIC0gbmFtZTogXCJlY3JcIlxuICAgICAgdHlwZTogXCJhd3MtZWNyXCJcbiAgICAgIHJlZ2lvbjogXCIke3RoaXMucmVnaW9ufVwiXG4gICAgICBjcmVkZW50aWFsU291cmNlOiBcImlhbS1yb2xlXCJcblxucHJveHk6XG4gIGJhY2tlbmRzOiBbXVxuXG5sb2dnaW5nOlxuICBsZXZlbDogXCJpbmZvXCJgO1xuICB9XG59Il19