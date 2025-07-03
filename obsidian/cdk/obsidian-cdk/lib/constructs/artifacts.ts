import * as cdk from 'aws-cdk-lib';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';

export class ObsidianArtifactsConstruct extends Construct {
  public readonly bucket: s3.Bucket;
  public readonly bucketPolicy: iam.PolicyStatement;

  constructor(scope: Construct, id: string) {
    super(scope, id);

    // Create S3 bucket for Obsidian artifacts
    this.bucket = new s3.Bucket(this, 'ArtifactsBucket', {
      bucketName: `obsidian-artifacts-${cdk.Stack.of(this).account}-${cdk.Stack.of(this).region}`,
      versioned: true,
      encryption: s3.BucketEncryption.S3_MANAGED,
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      removalPolicy: cdk.RemovalPolicy.RETAIN,
      lifecycleRules: [{
        id: 'delete-old-versions',
        noncurrentVersionExpiration: cdk.Duration.days(30),
      }],
    });

    // Create policy for EC2 instances to read from bucket
    this.bucketPolicy = new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: [
        's3:GetObject',
        's3:ListBucket',
      ],
      resources: [
        this.bucket.bucketArn,
        `${this.bucket.bucketArn}/*`,
      ],
    });

    // Output the bucket name
    new cdk.CfnOutput(this, 'ArtifactsBucketName', {
      value: this.bucket.bucketName,
      description: 'S3 bucket for Obsidian artifacts',
    });
  }
}