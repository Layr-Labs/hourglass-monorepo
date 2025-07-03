import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as autoscaling from 'aws-cdk-lib/aws-autoscaling';
import * as s3 from 'aws-cdk-lib/aws-s3';
import { Construct } from 'constructs';
import { ObsidianArtifactsConstruct } from '../constructs/artifacts';
import { ObsidianBuildPipeline } from '../constructs/build-pipeline';
import { ObsidianBuildTrigger } from '../constructs/build-trigger';

export interface ObsidianEC2StackProps extends cdk.StackProps {
  instanceType?: ec2.InstanceType;
  minSize?: number;
  maxSize?: number;
  vpcId?: string;
}

export class ObsidianEC2Stack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: ObsidianEC2StackProps = {}) {
    super(scope, id, props);

    // Create artifacts bucket and build pipeline
    const artifacts = new ObsidianArtifactsConstruct(this, 'Artifacts');
    const buildPipeline = new ObsidianBuildPipeline(this, 'BuildPipeline', {
      artifactsBucket: artifacts.bucket,
    });
    const buildTrigger = new ObsidianBuildTrigger(this, 'BuildTrigger', {
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
    securityGroup.addIngressRule(
      ec2.Peer.anyIpv4(),
      ec2.Port.tcp(8080),
      'Allow HTTP traffic'
    );

    securityGroup.addIngressRule(
      ec2.Peer.anyIpv4(),
      ec2.Port.tcp(9090),
      'Allow gRPC traffic'
    );

    securityGroup.addIngressRule(
      ec2.Peer.anyIpv4(),
      ec2.Port.tcp(9091),
      'Allow metrics traffic'
    );

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
      'yum update -y',
      'yum install -y docker',
      'service docker start',
      'usermod -a -G docker ec2-user',
      
      // Install Docker Compose
      'curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose',
      'chmod +x /usr/local/bin/docker-compose',
      
      // Download and start Obsidian
      'mkdir -p /opt/obsidian',
      'cd /opt/obsidian',
      
      // Create configuration
      'cat > /opt/obsidian/config.yaml << EOF',
      this.generateConfigYaml(),
      'EOF',
      
      // Download Obsidian binary from S3
      `aws s3 cp s3://${artifacts.bucket.bucketName}/binaries/obsidian-linux-amd64-latest /opt/obsidian/obsidian`,
      'chmod +x /opt/obsidian/obsidian',
      
      // Verify binary downloaded successfully
      'if [ ! -f /opt/obsidian/obsidian ]; then',
      '  echo "ERROR: Failed to download Obsidian binary from S3"',
      '  exit 1',
      'fi',
      
      // Create systemd service
      'cat > /etc/systemd/system/obsidian.service << EOF',
      '[Unit]',
      'Description=Obsidian Gateway',
      'After=docker.service',
      'Requires=docker.service',
      '',
      '[Service]',
      'Type=simple',
      'ExecStart=/opt/obsidian/obsidian --config=/opt/obsidian/config.yaml',
      'Restart=always',
      'RestartSec=10',
      'StandardOutput=journal',
      'StandardError=journal',
      'SyslogIdentifier=obsidian',
      '',
      '[Install]',
      'WantedBy=multi-user.target',
      'EOF',
      
      // Start Obsidian
      'systemctl daemon-reload',
      'systemctl enable obsidian',
      'systemctl start obsidian'
    );

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

  private generateConfigYaml(): string {
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