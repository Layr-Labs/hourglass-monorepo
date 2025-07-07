import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as eks from 'aws-cdk-lib/aws-eks';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';

export class K8sClusterStack extends cdk.Stack {
  public readonly cluster: eks.Cluster;
  public readonly vpc: ec2.Vpc;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create VPC
    this.vpc = new ec2.Vpc(this, 'HourglassVpc', {
      maxAzs: 2,
      natGateways: 1,
      subnetConfiguration: [
        {
          name: 'Public',
          subnetType: ec2.SubnetType.PUBLIC,
          cidrMask: 24,
        },
        {
          name: 'Private',
          subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
          cidrMask: 24,
        },
      ],
    });

    // Create EKS cluster
    this.cluster = new eks.Cluster(this, 'HourglassCluster', {
      vpc: this.vpc,
      version: eks.KubernetesVersion.V1_28,
      defaultCapacity: 0, // We'll add our own node groups
      clusterName: 'hourglass-eks',
    });

    // Add managed node group for general workloads
    const generalNodeGroup = this.cluster.addNodegroupCapacity('GeneralNodeGroup', {
      instanceTypes: [
        ec2.InstanceType.of(ec2.InstanceClass.M5, ec2.InstanceSize.LARGE),
        ec2.InstanceType.of(ec2.InstanceClass.M5A, ec2.InstanceSize.LARGE),
      ],
      minSize: 2,
      maxSize: 10,
      desiredSize: 3,
      diskSize: 100,
      capacityType: eks.CapacityType.SPOT,
      subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
      labels: {
        'workload-type': 'general',
      },
    });

    // Add node group for executor workloads (needs Docker-in-Docker)
    const executorNodeGroup = this.cluster.addNodegroupCapacity('ExecutorNodeGroup', {
      instanceTypes: [
        ec2.InstanceType.of(ec2.InstanceClass.M5, ec2.InstanceSize.XLARGE),
      ],
      minSize: 1,
      maxSize: 5,
      desiredSize: 2,
      diskSize: 200,
      capacityType: eks.CapacityType.ON_DEMAND,
      subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
      labels: {
        'workload-type': 'executor',
        'supports-dind': 'true',
      },
      taints: [
        {
          key: 'executor-only',
          value: 'true',
          effect: eks.TaintEffect.NO_SCHEDULE,
        },
      ],
    });

    // Install AWS Load Balancer Controller
    const awsLbControllerServiceAccount = this.cluster.addServiceAccount('aws-load-balancer-controller', {
      name: 'aws-load-balancer-controller',
      namespace: 'kube-system',
    });

    awsLbControllerServiceAccount.addToPrincipalPolicy(
      new iam.PolicyStatement({
        effect: iam.Effect.ALLOW,
        actions: [
          'elasticloadbalancing:*',
          'ec2:CreateTags',
          'ec2:DeleteTags',
          'ec2:DescribeAccountAttributes',
          'ec2:DescribeAddresses',
          'ec2:DescribeAvailabilityZones',
          'ec2:DescribeInstances',
          'ec2:DescribeInstanceStatus',
          'ec2:DescribeInternetGateways',
          'ec2:DescribeNetworkInterfaces',
          'ec2:DescribeSecurityGroups',
          'ec2:DescribeSubnets',
          'ec2:DescribeTags',
          'ec2:DescribeVpcs',
          'ec2:ModifyInstanceAttribute',
          'ec2:ModifyNetworkInterfaceAttribute',
          'ec2:AssignPrivateIpAddresses',
          'ec2:UnassignPrivateIpAddresses',
          'ec2:AuthorizeSecurityGroupIngress',
          'ec2:RevokeSecurityGroupIngress',
          'cognito-idp:DescribeUserPoolClient',
          'waf-regional:GetWebACLForResource',
          'waf-regional:GetWebACL',
          'waf-regional:AssociateWebACL',
          'waf-regional:DisassociateWebACL',
          'tag:GetResources',
          'tag:TagResources',
          'waf:GetWebACL',
          'shield:DescribeProtection',
          'shield:GetSubscriptionState',
          'shield:DeleteProtection',
          'shield:CreateProtection',
          'shield:DescribeSubscription',
          'shield:ListProtections',
        ],
        resources: ['*'],
      })
    );

    // Install AWS Load Balancer Controller using Helm
    const awsLbControllerChart = this.cluster.addHelmChart('AWSLoadBalancerController', {
      chart: 'aws-load-balancer-controller',
      repository: 'https://aws.github.io/eks-charts',
      namespace: 'kube-system',
      values: {
        clusterName: this.cluster.clusterName,
        serviceAccount: {
          create: false,
          name: 'aws-load-balancer-controller',
        },
        region: cdk.Stack.of(this).region,
        vpcId: this.vpc.vpcId,
      },
    });

    // Install EBS CSI Driver
    const ebsCsiDriverServiceAccount = this.cluster.addServiceAccount('ebs-csi-driver', {
      name: 'ebs-csi-controller-sa',
      namespace: 'kube-system',
    });

    ebsCsiDriverServiceAccount.role.addManagedPolicy(
      iam.ManagedPolicy.fromAwsManagedPolicyName('service-role/AmazonEBSCSIDriverPolicy')
    );

    const ebsCsiDriverChart = this.cluster.addHelmChart('EBSCSIDriver', {
      chart: 'aws-ebs-csi-driver',
      repository: 'https://kubernetes-sigs.github.io/aws-ebs-csi-driver',
      namespace: 'kube-system',
      values: {
        controller: {
          serviceAccount: {
            create: false,
            name: 'ebs-csi-controller-sa',
          },
        },
        node: {
          tolerateAllTaints: true,
        },
      },
    });

    // Add default storage class
    const storageClass = this.cluster.addManifest('gp3-storage-class', {
      apiVersion: 'storage.k8s.io/v1',
      kind: 'StorageClass',
      metadata: {
        name: 'gp3',
        annotations: {
          'storageclass.kubernetes.io/is-default-class': 'true',
        },
      },
      provisioner: 'ebs.csi.aws.com',
      parameters: {
        type: 'gp3',
        encrypted: 'true',
      },
      volumeBindingMode: 'WaitForFirstConsumer',
      allowVolumeExpansion: true,
    });

    // Output cluster information
    new cdk.CfnOutput(this, 'ClusterName', {
      value: this.cluster.clusterName,
      description: 'Name of the EKS cluster',
    });

    new cdk.CfnOutput(this, 'ClusterEndpoint', {
      value: this.cluster.clusterEndpoint,
      description: 'Endpoint for the EKS cluster',
    });

    new cdk.CfnOutput(this, 'KubectlCommand', {
      value: `aws eks update-kubeconfig --name ${this.cluster.clusterName} --region ${cdk.Stack.of(this).region}`,
      description: 'Command to configure kubectl',
    });
  }
}