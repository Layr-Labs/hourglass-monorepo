import * as cdk from 'aws-cdk-lib';
import * as codebuild from 'aws-cdk-lib/aws-codebuild';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';

export interface BuildPipelineProps {
  artifactsBucket: s3.Bucket;
  githubRepo?: string;
  branch?: string;
}

export class ObsidianBuildPipeline extends Construct {
  public readonly project: codebuild.Project;

  constructor(scope: Construct, id: string, props: BuildPipelineProps) {
    super(scope, id);

    // Create CodeBuild project
    this.project = new codebuild.Project(this, 'ObsidianBuildProject', {
      projectName: 'obsidian-build',
      description: 'Build and deploy Obsidian binary to S3',
      source: codebuild.Source.gitHub({
        owner: 'hourglass',
        repo: props.githubRepo || 'obsidian',
        webhook: true,
        webhookFilters: [
          codebuild.FilterGroup.inEventOf(codebuild.EventAction.PUSH).andBranchIs(props.branch || 'main'),
        ],
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.STANDARD_7_0,
        computeType: codebuild.ComputeType.MEDIUM,
        environmentVariables: {
          ARTIFACTS_BUCKET: {
            value: props.artifactsBucket.bucketName,
          },
          AWS_ACCOUNT_ID: {
            value: cdk.Stack.of(this).account,
          },
          AWS_REGION: {
            value: cdk.Stack.of(this).region,
          },
        },
      },
      buildSpec: codebuild.BuildSpec.fromObject({
        version: '0.2',
        phases: {
          install: {
            'runtime-versions': {
              golang: '1.21',
            },
            commands: [
              'echo Installing dependencies...',
              'apt-get update -y',
              'apt-get install -y protobuf-compiler',
              'go install google.golang.org/protobuf/cmd/protoc-gen-go@latest',
              'go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest',
            ],
          },
          pre_build: {
            commands: [
              'echo Pre-build phase started on `date`',
              'export VERSION=$(git describe --tags --always --dirty)',
              'export BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)',
              'export GIT_COMMIT=$(git rev-parse HEAD)',
              'echo "Building version: $VERSION"',
            ],
          },
          build: {
            commands: [
              'echo Build started on `date`',
              'cd obsidian',
              'make proto',
              'GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" -o obsidian-linux-amd64 ./cmd/obsidian',
              'echo Build completed on `date`',
            ],
          },
          post_build: {
            commands: [
              'echo Post-build phase started on `date`',
              '# Upload versioned binary',
              'aws s3 cp obsidian-linux-amd64 s3://$ARTIFACTS_BUCKET/binaries/obsidian-linux-amd64-$VERSION',
              '# Update latest pointer',
              'aws s3 cp obsidian-linux-amd64 s3://$ARTIFACTS_BUCKET/binaries/obsidian-linux-amd64-latest',
              '# Create and upload tarball with config',
              'tar -czf obsidian-$VERSION.tar.gz obsidian-linux-amd64 config/production.yaml',
              'aws s3 cp obsidian-$VERSION.tar.gz s3://$ARTIFACTS_BUCKET/releases/obsidian-$VERSION.tar.gz',
              'aws s3 cp obsidian-$VERSION.tar.gz s3://$ARTIFACTS_BUCKET/releases/obsidian-latest.tar.gz',
              'echo "Binary uploaded to s3://$ARTIFACTS_BUCKET/binaries/obsidian-linux-amd64-latest"',
            ],
          },
        },
        artifacts: {
          files: [
            'obsidian-linux-amd64',
            'config/production.yaml',
          ],
          name: 'obsidian-$VERSION',
        },
      }),
      artifacts: codebuild.Artifacts.s3({
        bucket: props.artifactsBucket,
        path: 'build-artifacts',
        includeBuildId: false,
        packageZip: true,
      }),
    });

    // Grant permissions to upload to S3
    props.artifactsBucket.grantWrite(this.project);

    // Output the project name
    new cdk.CfnOutput(this, 'BuildProjectName', {
      value: this.project.projectName,
      description: 'CodeBuild project name',
    });
  }

  // Method to trigger a build programmatically
  public triggerBuild(): void {
    // This would be called from a Lambda or manually via AWS CLI
    // aws codebuild start-build --project-name obsidian-build
  }
}