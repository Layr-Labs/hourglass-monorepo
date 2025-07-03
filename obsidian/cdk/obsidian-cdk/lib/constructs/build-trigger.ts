import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as codebuild from 'aws-cdk-lib/aws-codebuild';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';

export interface BuildTriggerProps {
  buildProject: codebuild.Project;
}

export class ObsidianBuildTrigger extends Construct {
  public readonly triggerFunction: lambda.Function;

  constructor(scope: Construct, id: string, props: BuildTriggerProps) {
    super(scope, id);

    // Create Lambda function to trigger builds
    this.triggerFunction = new lambda.Function(this, 'BuildTriggerFunction', {
      runtime: lambda.Runtime.NODEJS_18_X,
      handler: 'index.handler',
      code: lambda.Code.fromInline(`
        const AWS = require('aws-sdk');
        const codebuild = new AWS.CodeBuild();
        
        exports.handler = async (event) => {
          const params = {
            projectName: '${props.buildProject.projectName}',
            sourceVersion: event.branch || 'main'
          };
          
          try {
            const result = await codebuild.startBuild(params).promise();
            return {
              statusCode: 200,
              body: JSON.stringify({
                message: 'Build started successfully',
                buildId: result.build.id,
                buildArn: result.build.arn
              })
            };
          } catch (error) {
            return {
              statusCode: 500,
              body: JSON.stringify({
                message: 'Failed to start build',
                error: error.message
              })
            };
          }
        };
      `),
      timeout: cdk.Duration.seconds(30),
      functionName: 'obsidian-build-trigger',
    });

    // Grant permissions to trigger builds
    this.triggerFunction.addToRolePolicy(new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: [
        'codebuild:StartBuild',
        'codebuild:BatchGetBuilds',
      ],
      resources: [props.buildProject.projectArn],
    }));

    // Output the function name for easy invocation
    new cdk.CfnOutput(this, 'BuildTriggerFunctionName', {
      value: this.triggerFunction.functionName,
      description: 'Lambda function to trigger builds',
    });

    // Output command to trigger build
    new cdk.CfnOutput(this, 'TriggerBuildCommand', {
      value: `aws lambda invoke --function-name ${this.triggerFunction.functionName} --payload '{"branch":"main"}' response.json`,
      description: 'Command to trigger a build',
    });
  }
}