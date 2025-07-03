import * as s3 from 'aws-cdk-lib/aws-s3';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';
export declare class ObsidianArtifactsConstruct extends Construct {
    readonly bucket: s3.Bucket;
    readonly bucketPolicy: iam.PolicyStatement;
    constructor(scope: Construct, id: string);
}
