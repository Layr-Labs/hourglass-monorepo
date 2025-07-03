import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as codebuild from 'aws-cdk-lib/aws-codebuild';
import { Construct } from 'constructs';
export interface BuildTriggerProps {
    buildProject: codebuild.Project;
}
export declare class ObsidianBuildTrigger extends Construct {
    readonly triggerFunction: lambda.Function;
    constructor(scope: Construct, id: string, props: BuildTriggerProps);
}
