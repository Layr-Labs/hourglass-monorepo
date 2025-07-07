import * as cdk from 'aws-cdk-lib';
import * as eks from 'aws-cdk-lib/aws-eks';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as ssm from 'aws-cdk-lib/aws-ssm';
import { Construct } from 'constructs';

interface HourglassServicesStackProps extends cdk.StackProps {
  cluster: eks.Cluster;
  anvilL1Endpoint: string;
  anvilL2Endpoint: string;
  contractAddresses: { [key: string]: string };
}

export class HourglassServicesStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: HourglassServicesStackProps) {
    super(scope, id, props);

    const { cluster, anvilL1Endpoint, anvilL2Endpoint, contractAddresses } = props;

    // Read configuration
    const devnetConfig = require('../../config/devnet.json');

    // Create namespace
    const namespace = cluster.addManifest('hourglass-namespace', {
      apiVersion: 'v1',
      kind: 'Namespace',
      metadata: {
        name: 'hourglass',
      },
    });

    // Create ConfigMap for aggregator
    const aggregatorConfig = cluster.addManifest('aggregator-config', {
      apiVersion: 'v1',
      kind: 'ConfigMap',
      metadata: {
        name: 'aggregator-config',
        namespace: 'hourglass',
      },
      data: {
        'aggregator.yaml': `
debug: false
simulationConfig:
  simulatePeering:
    enabled: true
    operatorPeers:
      - networkAddress: "executor-service.hourglass.svc.cluster.local:9090"
        operatorAddress: "${devnetConfig.operators.executor.address}"
        operatorSetId: 1

serverConfig:
  port: 9000
  aggregatorUrl: "aggregator-service.hourglass.svc.cluster.local:9000"

operator:
  address: "${devnetConfig.operators.aggregator.address}"
  operatorPrivateKey: "${devnetConfig.operators.aggregator.privateKey}"
  signingKeys:
    bls:
      keystoreFile: "/keys/aggregator/key_bn254.json"
      password: ""

l1ChainId: ${devnetConfig.l1.chainId}

chains:
  - name: "ethereum"
    network: "mainnet"
    chainId: ${devnetConfig.l1.chainId}
    rpcUrl: "${anvilL1Endpoint}"
    pollIntervalSeconds: 10
  - name: "base"
    network: "mainnet"
    chainId: ${devnetConfig.l2.chainId}
    rpcUrl: "${anvilL2Endpoint}"
    pollIntervalSeconds: 2

avss:
  - address: "${devnetConfig.avs.address}"
    responseTimeout: ${devnetConfig.avs.responseTimeout}
    chainIds:
    - ${devnetConfig.l1.chainId}
    - ${devnetConfig.l2.chainId}
    signingCurve: "${devnetConfig.avs.signingCurve}"
    avsRegistrarAddress: "${contractAddresses.avsTaskRegistrar}"
`,
      },
    });

    // Create ConfigMap for executor
    const executorConfig = cluster.addManifest('executor-config', {
      apiVersion: 'v1',
      kind: 'ConfigMap',
      metadata: {
        name: 'executor-config',
        namespace: 'hourglass',
      },
      data: {
        'executor.yaml': `
grpcPort: 9090
performerNetworkName: default
operator:
  address: "${devnetConfig.operators.executor.address}"
  operatorPrivateKey: "${devnetConfig.operators.executor.privateKey}"
  signingKeys:
    bls:
      keystoreFile: "/keys/executor/key_bn254.json"
      password: ""
l1Chain:
  rpcUrl: "${anvilL1Endpoint}"
  chainId: ${devnetConfig.l1.chainId}

avsPerformers:
  - image:
      repository: "hello-performer"
      tag: "latest"
    processType: "server"
    avsAddress: "${devnetConfig.avs.address}"
    workerCount: 1
    signingCurve: "${devnetConfig.avs.signingCurve}"
    avsRegistrarAddress: "${contractAddresses.avsTaskRegistrar}"
simulation:
  simulatePeering:
    enabled: true
    aggregatorPeers:
      - networkAddress: "aggregator-service.hourglass.svc.cluster.local:9000"
        operatorAddress: "${devnetConfig.operators.aggregator.address}"
        operatorSetId: 0
`,
      },
    });

    // Create Aggregator Deployment
    const aggregatorDeployment = cluster.addManifest('aggregator-deployment', {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: {
        name: 'aggregator',
        namespace: 'hourglass',
      },
      spec: {
        replicas: 1,
        selector: {
          matchLabels: {
            app: 'aggregator',
          },
        },
        template: {
          metadata: {
            labels: {
              app: 'aggregator',
            },
          },
          spec: {
            containers: [
              {
                name: 'aggregator',
                image: 'public.ecr.aws/z6g0f8n7/eigenlayer-hourglass:v0.1.0',
                command: ['aggregator', 'run', '--config', '/config/aggregator.yaml'],
                ports: [
                  {
                    containerPort: 9000,
                    name: 'grpc',
                  },
                  {
                    containerPort: 8081,
                    name: 'metrics',
                  },
                ],
                volumeMounts: [
                  {
                    name: 'config',
                    mountPath: '/config',
                  },
                  {
                    name: 'keys',
                    mountPath: '/keys',
                  },
                ],
                resources: {
                  requests: {
                    cpu: '500m',
                    memory: '1Gi',
                  },
                  limits: {
                    cpu: '2',
                    memory: '4Gi',
                  },
                },
              },
            ],
            volumes: [
              {
                name: 'config',
                configMap: {
                  name: 'aggregator-config',
                },
              },
              {
                name: 'keys',
                emptyDir: {},
              },
            ],
          },
        },
      },
    });

    // Create Aggregator Service
    const aggregatorService = cluster.addManifest('aggregator-service', {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: 'aggregator-service',
        namespace: 'hourglass',
      },
      spec: {
        selector: {
          app: 'aggregator',
        },
        ports: [
          {
            name: 'grpc',
            port: 9000,
            targetPort: 9000,
          },
          {
            name: 'metrics',
            port: 8081,
            targetPort: 8081,
          },
        ],
        type: 'ClusterIP',
      },
    });

    // Create Executor Deployment
    const executorDeployment = cluster.addManifest('executor-deployment', {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: {
        name: 'executor',
        namespace: 'hourglass',
      },
      spec: {
        replicas: 1,
        selector: {
          matchLabels: {
            app: 'executor',
          },
        },
        template: {
          metadata: {
            labels: {
              app: 'executor',
            },
          },
          spec: {
            serviceAccountName: 'executor-sa',
            nodeSelector: {
              'workload-type': 'executor',
            },
            tolerations: [
              {
                key: 'executor-only',
                operator: 'Equal',
                value: 'true',
                effect: 'NoSchedule',
              },
            ],
            containers: [
              {
                name: 'executor',
                image: 'public.ecr.aws/z6g0f8n7/eigenlayer-hourglass:v0.1.0',
                command: ['executor', 'run', '--config', '/config/executor.yaml'],
                ports: [
                  {
                    containerPort: 9090,
                    name: 'grpc',
                  },
                ],
                volumeMounts: [
                  {
                    name: 'config',
                    mountPath: '/config',
                  },
                  {
                    name: 'keys',
                    mountPath: '/keys',
                  },
                  {
                    name: 'docker-sock',
                    mountPath: '/var/run/docker.sock',
                  },
                ],
                resources: {
                  requests: {
                    cpu: '1',
                    memory: '2Gi',
                  },
                  limits: {
                    cpu: '4',
                    memory: '8Gi',
                  },
                },
                securityContext: {
                  privileged: true, // Required for Docker-in-Docker
                },
              },
            ],
            volumes: [
              {
                name: 'config',
                configMap: {
                  name: 'executor-config',
                },
              },
              {
                name: 'keys',
                emptyDir: {},
              },
              {
                name: 'docker-sock',
                hostPath: {
                  path: '/var/run/docker.sock',
                  type: 'Socket',
                },
              },
            ],
          },
        },
      },
    });

    // Create Executor Service
    const executorService = cluster.addManifest('executor-service', {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: 'executor-service',
        namespace: 'hourglass',
      },
      spec: {
        selector: {
          app: 'executor',
        },
        ports: [
          {
            name: 'grpc',
            port: 9090,
            targetPort: 9090,
          },
        ],
        type: 'ClusterIP',
      },
    });

    // Create service account for executor (needs Docker access)
    const executorSa = cluster.addManifest('executor-sa', {
      apiVersion: 'v1',
      kind: 'ServiceAccount',
      metadata: {
        name: 'executor-sa',
        namespace: 'hourglass',
      },
    });

    // Create LoadBalancer for external access to aggregator
    const aggregatorLb = cluster.addManifest('aggregator-lb', {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: 'aggregator-lb',
        namespace: 'hourglass',
        annotations: {
          'service.beta.kubernetes.io/aws-load-balancer-type': 'nlb',
        },
      },
      spec: {
        selector: {
          app: 'aggregator',
        },
        ports: [
          {
            name: 'grpc',
            port: 9000,
            targetPort: 9000,
          },
        ],
        type: 'LoadBalancer',
      },
    });

    // Add dependencies
    aggregatorConfig.node.addDependency(namespace);
    executorConfig.node.addDependency(namespace);
    aggregatorDeployment.node.addDependency(aggregatorConfig);
    executorDeployment.node.addDependency(executorConfig);
    aggregatorService.node.addDependency(aggregatorDeployment);
    executorService.node.addDependency(executorDeployment);
    executorSa.node.addDependency(namespace);
    executorDeployment.node.addDependency(executorSa);
    aggregatorLb.node.addDependency(aggregatorService);

    // Store service endpoints in Parameter Store
    new ssm.StringParameter(this, 'AggregatorEndpoint', {
      parameterName: '/hourglass/services/aggregator-endpoint',
      stringValue: 'aggregator-lb.hourglass.svc.cluster.local:9000',
    });

    new ssm.StringParameter(this, 'ExecutorEndpoint', {
      parameterName: '/hourglass/services/executor-endpoint',
      stringValue: 'executor-service.hourglass.svc.cluster.local:9090',
    });

    // Outputs
    new cdk.CfnOutput(this, 'AggregatorServiceEndpoint', {
      value: 'Use kubectl get svc -n hourglass aggregator-lb to get the external endpoint',
      description: 'Aggregator gRPC endpoint',
    });
  }
}