import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import { Construct } from 'constructs';
export interface ObsidianEC2StackProps extends cdk.StackProps {
    instanceType?: ec2.InstanceType;
    minSize?: number;
    maxSize?: number;
    vpcId?: string;
}
export declare class ObsidianEC2Stack extends cdk.Stack {
    constructor(scope: Construct, id: string, props?: ObsidianEC2StackProps);
    private generateConfigYaml;
}
