import * as codebuild from 'aws-cdk-lib/aws-codebuild';
import * as s3 from 'aws-cdk-lib/aws-s3';
import { Construct } from 'constructs';
export interface BuildPipelineProps {
    artifactsBucket: s3.Bucket;
    githubRepo?: string;
    branch?: string;
}
export declare class ObsidianBuildPipeline extends Construct {
    readonly project: codebuild.Project;
    constructor(scope: Construct, id: string, props: BuildPipelineProps);
    triggerBuild(): void;
}
