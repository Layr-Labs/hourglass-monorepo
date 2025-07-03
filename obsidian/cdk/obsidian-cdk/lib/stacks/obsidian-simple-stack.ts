import * as cdk from 'aws-cdk-lib';
import * as codebuild from 'aws-cdk-lib/aws-codebuild';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as autoscaling from 'aws-cdk-lib/aws-autoscaling';
import { Construct } from 'constructs';

export class ObsidianSimpleStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
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
    
    securityGroup.addIngressRule(
      ec2.Peer.anyIpv4(),
      ec2.Port.tcp(8080),
      'HTTP API'
    );
    
    securityGroup.addIngressRule(
      ec2.Peer.anyIpv4(),
      ec2.Port.tcp(9090),
      'gRPC API'
    );

    // 🔹 User data script
    const userData = ec2.UserData.forLinux();
    userData.addCommands(
      // Install Docker
      'yum update -y',
      'yum install -y docker aws-cli',
      'service docker start',
      'usermod -a -G docker ec2-user',
      
      // Setup Obsidian
      'mkdir -p /opt/obsidian',
      
      // Create default config
      'cat > /opt/obsidian/config.yaml << EOF',
      this.generateMinimalConfig(),
      'EOF',
      
      // Download Obsidian binary from S3
      `aws s3 cp s3://${artifactBucket.bucketName}/binaries/obsidian-linux-amd64 /opt/obsidian/obsidian || echo "Binary not yet available"`,
      'chmod +x /opt/obsidian/obsidian || true',
      
      // Create systemd service
      'cat > /etc/systemd/system/obsidian.service << EOF',
      '[Unit]',
      'Description=Obsidian Gateway',
      'After=docker.service',
      '',
      '[Service]',
      'Type=simple',
      'ExecStart=/opt/obsidian/obsidian --config=/opt/obsidian/config.yaml',
      'Restart=always',
      'StandardOutput=journal',
      '',
      '[Install]',
      'WantedBy=multi-user.target',
      'EOF',
      
      'systemctl daemon-reload',
      'systemctl enable obsidian',
      '# Start only if binary exists',
      '[ -f /opt/obsidian/obsidian ] && systemctl start obsidian || echo "Waiting for binary"'
    );

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

  private generateMinimalConfig(): string {
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