import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as ecs from 'aws-cdk-lib/aws-ecs';
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2';
import { Construct } from 'constructs';
export interface ObsidianHybridStackProps extends cdk.StackProps {
    minComputeNodes?: number;
    maxComputeNodes?: number;
    computeInstanceType?: ec2.InstanceType;
    vpcId?: string;
    controlPlane?: {
        orchestrator: {
            desiredCount: number;
            cpu: number;
            memory: number;
        };
        registry: {
            desiredCount: number;
            cpu: number;
            memory: number;
        };
        proxy: {
            desiredCount: number;
            cpu: number;
            memory: number;
        };
    };
}
export declare class ObsidianHybridStack extends cdk.Stack {
    readonly vpc: ec2.IVpc;
    readonly cluster: ecs.Cluster;
    readonly alb: elbv2.ApplicationLoadBalancer;
    constructor(scope: Construct, id: string, props?: ObsidianHybridStackProps);
    private createOrchestratorService;
    private createRegistryService;
    private createProxyService;
    private createComputeNodes;
    private createDashboard;
}
