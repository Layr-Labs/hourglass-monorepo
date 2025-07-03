import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
export interface ObsidianECSStackProps extends cdk.StackProps {
    vpcId?: string;
    orchestratorConfig?: {
        desiredCount: number;
        cpu: number;
        memory: number;
    };
    registryConfig?: {
        desiredCount: number;
        cpu: number;
        memory: number;
    };
    proxyConfig?: {
        desiredCount: number;
        cpu: number;
        memory: number;
    };
}
export declare class ObsidianECSStack extends cdk.Stack {
    constructor(scope: Construct, id: string, props?: ObsidianECSStackProps);
}
