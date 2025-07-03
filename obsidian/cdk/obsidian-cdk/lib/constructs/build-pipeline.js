"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ObsidianBuildPipeline = void 0;
const cdk = require("aws-cdk-lib");
const codebuild = require("aws-cdk-lib/aws-codebuild");
const constructs_1 = require("constructs");
class ObsidianBuildPipeline extends constructs_1.Construct {
    constructor(scope, id, props) {
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
    triggerBuild() {
        // This would be called from a Lambda or manually via AWS CLI
        // aws codebuild start-build --project-name obsidian-build
    }
}
exports.ObsidianBuildPipeline = ObsidianBuildPipeline;
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiYnVpbGQtcGlwZWxpbmUuanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyJidWlsZC1waXBlbGluZS50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7QUFBQSxtQ0FBbUM7QUFDbkMsdURBQXVEO0FBR3ZELDJDQUF1QztBQVF2QyxNQUFhLHFCQUFzQixTQUFRLHNCQUFTO0lBR2xELFlBQVksS0FBZ0IsRUFBRSxFQUFVLEVBQUUsS0FBeUI7UUFDakUsS0FBSyxDQUFDLEtBQUssRUFBRSxFQUFFLENBQUMsQ0FBQztRQUVqQiwyQkFBMkI7UUFDM0IsSUFBSSxDQUFDLE9BQU8sR0FBRyxJQUFJLFNBQVMsQ0FBQyxPQUFPLENBQUMsSUFBSSxFQUFFLHNCQUFzQixFQUFFO1lBQ2pFLFdBQVcsRUFBRSxnQkFBZ0I7WUFDN0IsV0FBVyxFQUFFLHdDQUF3QztZQUNyRCxNQUFNLEVBQUUsU0FBUyxDQUFDLE1BQU0sQ0FBQyxNQUFNLENBQUM7Z0JBQzlCLEtBQUssRUFBRSxXQUFXO2dCQUNsQixJQUFJLEVBQUUsS0FBSyxDQUFDLFVBQVUsSUFBSSxVQUFVO2dCQUNwQyxPQUFPLEVBQUUsSUFBSTtnQkFDYixjQUFjLEVBQUU7b0JBQ2QsU0FBUyxDQUFDLFdBQVcsQ0FBQyxTQUFTLENBQUMsU0FBUyxDQUFDLFdBQVcsQ0FBQyxJQUFJLENBQUMsQ0FBQyxXQUFXLENBQUMsS0FBSyxDQUFDLE1BQU0sSUFBSSxNQUFNLENBQUM7aUJBQ2hHO2FBQ0YsQ0FBQztZQUNGLFdBQVcsRUFBRTtnQkFDWCxVQUFVLEVBQUUsU0FBUyxDQUFDLGVBQWUsQ0FBQyxZQUFZO2dCQUNsRCxXQUFXLEVBQUUsU0FBUyxDQUFDLFdBQVcsQ0FBQyxNQUFNO2dCQUN6QyxvQkFBb0IsRUFBRTtvQkFDcEIsZ0JBQWdCLEVBQUU7d0JBQ2hCLEtBQUssRUFBRSxLQUFLLENBQUMsZUFBZSxDQUFDLFVBQVU7cUJBQ3hDO29CQUNELGNBQWMsRUFBRTt3QkFDZCxLQUFLLEVBQUUsR0FBRyxDQUFDLEtBQUssQ0FBQyxFQUFFLENBQUMsSUFBSSxDQUFDLENBQUMsT0FBTztxQkFDbEM7b0JBQ0QsVUFBVSxFQUFFO3dCQUNWLEtBQUssRUFBRSxHQUFHLENBQUMsS0FBSyxDQUFDLEVBQUUsQ0FBQyxJQUFJLENBQUMsQ0FBQyxNQUFNO3FCQUNqQztpQkFDRjthQUNGO1lBQ0QsU0FBUyxFQUFFLFNBQVMsQ0FBQyxTQUFTLENBQUMsVUFBVSxDQUFDO2dCQUN4QyxPQUFPLEVBQUUsS0FBSztnQkFDZCxNQUFNLEVBQUU7b0JBQ04sT0FBTyxFQUFFO3dCQUNQLGtCQUFrQixFQUFFOzRCQUNsQixNQUFNLEVBQUUsTUFBTTt5QkFDZjt3QkFDRCxRQUFRLEVBQUU7NEJBQ1IsaUNBQWlDOzRCQUNqQyxtQkFBbUI7NEJBQ25CLHNDQUFzQzs0QkFDdEMsZ0VBQWdFOzRCQUNoRSxpRUFBaUU7eUJBQ2xFO3FCQUNGO29CQUNELFNBQVMsRUFBRTt3QkFDVCxRQUFRLEVBQUU7NEJBQ1Isd0NBQXdDOzRCQUN4Qyx3REFBd0Q7NEJBQ3hELGtEQUFrRDs0QkFDbEQseUNBQXlDOzRCQUN6QyxtQ0FBbUM7eUJBQ3BDO3FCQUNGO29CQUNELEtBQUssRUFBRTt3QkFDTCxRQUFRLEVBQUU7NEJBQ1IsOEJBQThCOzRCQUM5QixhQUFhOzRCQUNiLFlBQVk7NEJBQ1osdUxBQXVMOzRCQUN2TCxnQ0FBZ0M7eUJBQ2pDO3FCQUNGO29CQUNELFVBQVUsRUFBRTt3QkFDVixRQUFRLEVBQUU7NEJBQ1IseUNBQXlDOzRCQUN6QywyQkFBMkI7NEJBQzNCLDhGQUE4Rjs0QkFDOUYseUJBQXlCOzRCQUN6Qiw0RkFBNEY7NEJBQzVGLHlDQUF5Qzs0QkFDekMsK0VBQStFOzRCQUMvRSw2RkFBNkY7NEJBQzdGLDJGQUEyRjs0QkFDM0YsdUZBQXVGO3lCQUN4RjtxQkFDRjtpQkFDRjtnQkFDRCxTQUFTLEVBQUU7b0JBQ1QsS0FBSyxFQUFFO3dCQUNMLHNCQUFzQjt3QkFDdEIsd0JBQXdCO3FCQUN6QjtvQkFDRCxJQUFJLEVBQUUsbUJBQW1CO2lCQUMxQjthQUNGLENBQUM7WUFDRixTQUFTLEVBQUUsU0FBUyxDQUFDLFNBQVMsQ0FBQyxFQUFFLENBQUM7Z0JBQ2hDLE1BQU0sRUFBRSxLQUFLLENBQUMsZUFBZTtnQkFDN0IsSUFBSSxFQUFFLGlCQUFpQjtnQkFDdkIsY0FBYyxFQUFFLEtBQUs7Z0JBQ3JCLFVBQVUsRUFBRSxJQUFJO2FBQ2pCLENBQUM7U0FDSCxDQUFDLENBQUM7UUFFSCxvQ0FBb0M7UUFDcEMsS0FBSyxDQUFDLGVBQWUsQ0FBQyxVQUFVLENBQUMsSUFBSSxDQUFDLE9BQU8sQ0FBQyxDQUFDO1FBRS9DLDBCQUEwQjtRQUMxQixJQUFJLEdBQUcsQ0FBQyxTQUFTLENBQUMsSUFBSSxFQUFFLGtCQUFrQixFQUFFO1lBQzFDLEtBQUssRUFBRSxJQUFJLENBQUMsT0FBTyxDQUFDLFdBQVc7WUFDL0IsV0FBVyxFQUFFLHdCQUF3QjtTQUN0QyxDQUFDLENBQUM7SUFDTCxDQUFDO0lBRUQsNkNBQTZDO0lBQ3RDLFlBQVk7UUFDakIsNkRBQTZEO1FBQzdELDBEQUEwRDtJQUM1RCxDQUFDO0NBQ0Y7QUFoSEQsc0RBZ0hDIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0ICogYXMgY2RrIGZyb20gJ2F3cy1jZGstbGliJztcbmltcG9ydCAqIGFzIGNvZGVidWlsZCBmcm9tICdhd3MtY2RrLWxpYi9hd3MtY29kZWJ1aWxkJztcbmltcG9ydCAqIGFzIHMzIGZyb20gJ2F3cy1jZGstbGliL2F3cy1zMyc7XG5pbXBvcnQgKiBhcyBpYW0gZnJvbSAnYXdzLWNkay1saWIvYXdzLWlhbSc7XG5pbXBvcnQgeyBDb25zdHJ1Y3QgfSBmcm9tICdjb25zdHJ1Y3RzJztcblxuZXhwb3J0IGludGVyZmFjZSBCdWlsZFBpcGVsaW5lUHJvcHMge1xuICBhcnRpZmFjdHNCdWNrZXQ6IHMzLkJ1Y2tldDtcbiAgZ2l0aHViUmVwbz86IHN0cmluZztcbiAgYnJhbmNoPzogc3RyaW5nO1xufVxuXG5leHBvcnQgY2xhc3MgT2JzaWRpYW5CdWlsZFBpcGVsaW5lIGV4dGVuZHMgQ29uc3RydWN0IHtcbiAgcHVibGljIHJlYWRvbmx5IHByb2plY3Q6IGNvZGVidWlsZC5Qcm9qZWN0O1xuXG4gIGNvbnN0cnVjdG9yKHNjb3BlOiBDb25zdHJ1Y3QsIGlkOiBzdHJpbmcsIHByb3BzOiBCdWlsZFBpcGVsaW5lUHJvcHMpIHtcbiAgICBzdXBlcihzY29wZSwgaWQpO1xuXG4gICAgLy8gQ3JlYXRlIENvZGVCdWlsZCBwcm9qZWN0XG4gICAgdGhpcy5wcm9qZWN0ID0gbmV3IGNvZGVidWlsZC5Qcm9qZWN0KHRoaXMsICdPYnNpZGlhbkJ1aWxkUHJvamVjdCcsIHtcbiAgICAgIHByb2plY3ROYW1lOiAnb2JzaWRpYW4tYnVpbGQnLFxuICAgICAgZGVzY3JpcHRpb246ICdCdWlsZCBhbmQgZGVwbG95IE9ic2lkaWFuIGJpbmFyeSB0byBTMycsXG4gICAgICBzb3VyY2U6IGNvZGVidWlsZC5Tb3VyY2UuZ2l0SHViKHtcbiAgICAgICAgb3duZXI6ICdob3VyZ2xhc3MnLFxuICAgICAgICByZXBvOiBwcm9wcy5naXRodWJSZXBvIHx8ICdvYnNpZGlhbicsXG4gICAgICAgIHdlYmhvb2s6IHRydWUsXG4gICAgICAgIHdlYmhvb2tGaWx0ZXJzOiBbXG4gICAgICAgICAgY29kZWJ1aWxkLkZpbHRlckdyb3VwLmluRXZlbnRPZihjb2RlYnVpbGQuRXZlbnRBY3Rpb24uUFVTSCkuYW5kQnJhbmNoSXMocHJvcHMuYnJhbmNoIHx8ICdtYWluJyksXG4gICAgICAgIF0sXG4gICAgICB9KSxcbiAgICAgIGVudmlyb25tZW50OiB7XG4gICAgICAgIGJ1aWxkSW1hZ2U6IGNvZGVidWlsZC5MaW51eEJ1aWxkSW1hZ2UuU1RBTkRBUkRfN18wLFxuICAgICAgICBjb21wdXRlVHlwZTogY29kZWJ1aWxkLkNvbXB1dGVUeXBlLk1FRElVTSxcbiAgICAgICAgZW52aXJvbm1lbnRWYXJpYWJsZXM6IHtcbiAgICAgICAgICBBUlRJRkFDVFNfQlVDS0VUOiB7XG4gICAgICAgICAgICB2YWx1ZTogcHJvcHMuYXJ0aWZhY3RzQnVja2V0LmJ1Y2tldE5hbWUsXG4gICAgICAgICAgfSxcbiAgICAgICAgICBBV1NfQUNDT1VOVF9JRDoge1xuICAgICAgICAgICAgdmFsdWU6IGNkay5TdGFjay5vZih0aGlzKS5hY2NvdW50LFxuICAgICAgICAgIH0sXG4gICAgICAgICAgQVdTX1JFR0lPTjoge1xuICAgICAgICAgICAgdmFsdWU6IGNkay5TdGFjay5vZih0aGlzKS5yZWdpb24sXG4gICAgICAgICAgfSxcbiAgICAgICAgfSxcbiAgICAgIH0sXG4gICAgICBidWlsZFNwZWM6IGNvZGVidWlsZC5CdWlsZFNwZWMuZnJvbU9iamVjdCh7XG4gICAgICAgIHZlcnNpb246ICcwLjInLFxuICAgICAgICBwaGFzZXM6IHtcbiAgICAgICAgICBpbnN0YWxsOiB7XG4gICAgICAgICAgICAncnVudGltZS12ZXJzaW9ucyc6IHtcbiAgICAgICAgICAgICAgZ29sYW5nOiAnMS4yMScsXG4gICAgICAgICAgICB9LFxuICAgICAgICAgICAgY29tbWFuZHM6IFtcbiAgICAgICAgICAgICAgJ2VjaG8gSW5zdGFsbGluZyBkZXBlbmRlbmNpZXMuLi4nLFxuICAgICAgICAgICAgICAnYXB0LWdldCB1cGRhdGUgLXknLFxuICAgICAgICAgICAgICAnYXB0LWdldCBpbnN0YWxsIC15IHByb3RvYnVmLWNvbXBpbGVyJyxcbiAgICAgICAgICAgICAgJ2dvIGluc3RhbGwgZ29vZ2xlLmdvbGFuZy5vcmcvcHJvdG9idWYvY21kL3Byb3RvYy1nZW4tZ29AbGF0ZXN0JyxcbiAgICAgICAgICAgICAgJ2dvIGluc3RhbGwgZ29vZ2xlLmdvbGFuZy5vcmcvZ3JwYy9jbWQvcHJvdG9jLWdlbi1nby1ncnBjQGxhdGVzdCcsXG4gICAgICAgICAgICBdLFxuICAgICAgICAgIH0sXG4gICAgICAgICAgcHJlX2J1aWxkOiB7XG4gICAgICAgICAgICBjb21tYW5kczogW1xuICAgICAgICAgICAgICAnZWNobyBQcmUtYnVpbGQgcGhhc2Ugc3RhcnRlZCBvbiBgZGF0ZWAnLFxuICAgICAgICAgICAgICAnZXhwb3J0IFZFUlNJT049JChnaXQgZGVzY3JpYmUgLS10YWdzIC0tYWx3YXlzIC0tZGlydHkpJyxcbiAgICAgICAgICAgICAgJ2V4cG9ydCBCVUlMRF9USU1FPSQoZGF0ZSAtdSArJVktJW0tJWRUJUg6JU06JVNaKScsXG4gICAgICAgICAgICAgICdleHBvcnQgR0lUX0NPTU1JVD0kKGdpdCByZXYtcGFyc2UgSEVBRCknLFxuICAgICAgICAgICAgICAnZWNobyBcIkJ1aWxkaW5nIHZlcnNpb246ICRWRVJTSU9OXCInLFxuICAgICAgICAgICAgXSxcbiAgICAgICAgICB9LFxuICAgICAgICAgIGJ1aWxkOiB7XG4gICAgICAgICAgICBjb21tYW5kczogW1xuICAgICAgICAgICAgICAnZWNobyBCdWlsZCBzdGFydGVkIG9uIGBkYXRlYCcsXG4gICAgICAgICAgICAgICdjZCBvYnNpZGlhbicsXG4gICAgICAgICAgICAgICdtYWtlIHByb3RvJyxcbiAgICAgICAgICAgICAgJ0dPT1M9bGludXggR09BUkNIPWFtZDY0IENHT19FTkFCTEVEPTAgZ28gYnVpbGQgLWxkZmxhZ3MgXCItWCBtYWluLlZlcnNpb249JFZFUlNJT04gLVggbWFpbi5CdWlsZFRpbWU9JEJVSUxEX1RJTUUgLVggbWFpbi5HaXRDb21taXQ9JEdJVF9DT01NSVRcIiAtbyBvYnNpZGlhbi1saW51eC1hbWQ2NCAuL2NtZC9vYnNpZGlhbicsXG4gICAgICAgICAgICAgICdlY2hvIEJ1aWxkIGNvbXBsZXRlZCBvbiBgZGF0ZWAnLFxuICAgICAgICAgICAgXSxcbiAgICAgICAgICB9LFxuICAgICAgICAgIHBvc3RfYnVpbGQ6IHtcbiAgICAgICAgICAgIGNvbW1hbmRzOiBbXG4gICAgICAgICAgICAgICdlY2hvIFBvc3QtYnVpbGQgcGhhc2Ugc3RhcnRlZCBvbiBgZGF0ZWAnLFxuICAgICAgICAgICAgICAnIyBVcGxvYWQgdmVyc2lvbmVkIGJpbmFyeScsXG4gICAgICAgICAgICAgICdhd3MgczMgY3Agb2JzaWRpYW4tbGludXgtYW1kNjQgczM6Ly8kQVJUSUZBQ1RTX0JVQ0tFVC9iaW5hcmllcy9vYnNpZGlhbi1saW51eC1hbWQ2NC0kVkVSU0lPTicsXG4gICAgICAgICAgICAgICcjIFVwZGF0ZSBsYXRlc3QgcG9pbnRlcicsXG4gICAgICAgICAgICAgICdhd3MgczMgY3Agb2JzaWRpYW4tbGludXgtYW1kNjQgczM6Ly8kQVJUSUZBQ1RTX0JVQ0tFVC9iaW5hcmllcy9vYnNpZGlhbi1saW51eC1hbWQ2NC1sYXRlc3QnLFxuICAgICAgICAgICAgICAnIyBDcmVhdGUgYW5kIHVwbG9hZCB0YXJiYWxsIHdpdGggY29uZmlnJyxcbiAgICAgICAgICAgICAgJ3RhciAtY3pmIG9ic2lkaWFuLSRWRVJTSU9OLnRhci5neiBvYnNpZGlhbi1saW51eC1hbWQ2NCBjb25maWcvcHJvZHVjdGlvbi55YW1sJyxcbiAgICAgICAgICAgICAgJ2F3cyBzMyBjcCBvYnNpZGlhbi0kVkVSU0lPTi50YXIuZ3ogczM6Ly8kQVJUSUZBQ1RTX0JVQ0tFVC9yZWxlYXNlcy9vYnNpZGlhbi0kVkVSU0lPTi50YXIuZ3onLFxuICAgICAgICAgICAgICAnYXdzIHMzIGNwIG9ic2lkaWFuLSRWRVJTSU9OLnRhci5neiBzMzovLyRBUlRJRkFDVFNfQlVDS0VUL3JlbGVhc2VzL29ic2lkaWFuLWxhdGVzdC50YXIuZ3onLFxuICAgICAgICAgICAgICAnZWNobyBcIkJpbmFyeSB1cGxvYWRlZCB0byBzMzovLyRBUlRJRkFDVFNfQlVDS0VUL2JpbmFyaWVzL29ic2lkaWFuLWxpbnV4LWFtZDY0LWxhdGVzdFwiJyxcbiAgICAgICAgICAgIF0sXG4gICAgICAgICAgfSxcbiAgICAgICAgfSxcbiAgICAgICAgYXJ0aWZhY3RzOiB7XG4gICAgICAgICAgZmlsZXM6IFtcbiAgICAgICAgICAgICdvYnNpZGlhbi1saW51eC1hbWQ2NCcsXG4gICAgICAgICAgICAnY29uZmlnL3Byb2R1Y3Rpb24ueWFtbCcsXG4gICAgICAgICAgXSxcbiAgICAgICAgICBuYW1lOiAnb2JzaWRpYW4tJFZFUlNJT04nLFxuICAgICAgICB9LFxuICAgICAgfSksXG4gICAgICBhcnRpZmFjdHM6IGNvZGVidWlsZC5BcnRpZmFjdHMuczMoe1xuICAgICAgICBidWNrZXQ6IHByb3BzLmFydGlmYWN0c0J1Y2tldCxcbiAgICAgICAgcGF0aDogJ2J1aWxkLWFydGlmYWN0cycsXG4gICAgICAgIGluY2x1ZGVCdWlsZElkOiBmYWxzZSxcbiAgICAgICAgcGFja2FnZVppcDogdHJ1ZSxcbiAgICAgIH0pLFxuICAgIH0pO1xuXG4gICAgLy8gR3JhbnQgcGVybWlzc2lvbnMgdG8gdXBsb2FkIHRvIFMzXG4gICAgcHJvcHMuYXJ0aWZhY3RzQnVja2V0LmdyYW50V3JpdGUodGhpcy5wcm9qZWN0KTtcblxuICAgIC8vIE91dHB1dCB0aGUgcHJvamVjdCBuYW1lXG4gICAgbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ0J1aWxkUHJvamVjdE5hbWUnLCB7XG4gICAgICB2YWx1ZTogdGhpcy5wcm9qZWN0LnByb2plY3ROYW1lLFxuICAgICAgZGVzY3JpcHRpb246ICdDb2RlQnVpbGQgcHJvamVjdCBuYW1lJyxcbiAgICB9KTtcbiAgfVxuXG4gIC8vIE1ldGhvZCB0byB0cmlnZ2VyIGEgYnVpbGQgcHJvZ3JhbW1hdGljYWxseVxuICBwdWJsaWMgdHJpZ2dlckJ1aWxkKCk6IHZvaWQge1xuICAgIC8vIFRoaXMgd291bGQgYmUgY2FsbGVkIGZyb20gYSBMYW1iZGEgb3IgbWFudWFsbHkgdmlhIEFXUyBDTElcbiAgICAvLyBhd3MgY29kZWJ1aWxkIHN0YXJ0LWJ1aWxkIC0tcHJvamVjdC1uYW1lIG9ic2lkaWFuLWJ1aWxkXG4gIH1cbn0iXX0=