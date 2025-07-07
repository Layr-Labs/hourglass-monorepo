import * as cdk from 'aws-cdk-lib';
import * as ecr from 'aws-cdk-lib/aws-ecr';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';
import { ReleaseMonitor } from '../constructs/release-monitor';
import { LambdaFactory } from '../constructs/lambda-factory';
import { PerformerRegistry } from '../constructs/performer-registry';

interface QuartzStackProps extends cdk.StackProps {
  releaseManagerAddress: string;
  rpcEndpoint: string;
  eksNodeRoleArn?: string;
}

export class QuartzStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: QuartzStackProps) {
    super(scope, id, props);

    // Create ECR repository for performer images
    const performerRepository = new ecr.Repository(this, 'PerformerRepository', {
      repositoryName: 'quartz-performer',
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      autoDeleteImages: true,
      imageScanOnPush: true,
      lifecycleRules: [{
        maxImageCount: 10,
        description: 'Keep only 10 most recent images',
      }],
    });

    // Create performer registry
    const registry = new PerformerRegistry(this, 'PerformerRegistry');

    // Create Lambda factory with Step Functions
    const factory = new LambdaFactory(this, 'LambdaFactory', {
      ecrRepository: performerRepository,
    });

    // Create release monitor
    const monitor = new ReleaseMonitor(this, 'ReleaseMonitor', {
      releaseManagerAddress: props.releaseManagerAddress,
      rpcEndpoint: props.rpcEndpoint,
      stateMachine: factory.stateMachine,
    });

    // Grant ECR permissions to EKS nodes if provided
    if (props.eksNodeRoleArn) {
      const eksNodeRole = iam.Role.fromRoleArn(this, 'EKSNodeRole', props.eksNodeRoleArn);
      performerRepository.grantPullPush(eksNodeRole);
    }

    // Outputs
    new cdk.CfnOutput(this, 'ECRRepositoryURI', {
      value: performerRepository.repositoryUri,
      description: 'ECR repository URI for performer images',
    });

    new cdk.CfnOutput(this, 'RegistryAPIEndpoint', {
      value: registry.api.url,
      description: 'API endpoint for performer registry',
    });

    new cdk.CfnOutput(this, 'StateMachineArn', {
      value: factory.stateMachine.stateMachineArn,
      description: 'Step Functions state machine ARN',
    });

    new cdk.CfnOutput(this, 'MonitorFunctionName', {
      value: monitor.indexerFunction.functionName,
      description: 'Release monitor Lambda function name',
    });

    // Add tags
    cdk.Tags.of(this).add('Component', 'quartz');
    cdk.Tags.of(this).add('ManagedBy', 'cdk');
  }
}