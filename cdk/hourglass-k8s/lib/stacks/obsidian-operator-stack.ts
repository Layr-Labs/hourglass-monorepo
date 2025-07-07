import * as cdk from 'aws-cdk-lib';
import * as eks from 'aws-cdk-lib/aws-eks';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';

interface ObsidianOperatorStackProps extends cdk.StackProps {
  cluster: eks.Cluster;
}

export class ObsidianOperatorStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: ObsidianOperatorStackProps) {
    super(scope, id, props);

    const { cluster } = props;

    // Create namespace for Obsidian
    const namespace = cluster.addManifest('obsidian-namespace', {
      apiVersion: 'v1',
      kind: 'Namespace',
      metadata: {
        name: 'obsidian-system',
      },
    });

    // Create CRD for AVS
    const avsCrd = cluster.addManifest('avs-crd', {
      apiVersion: 'apiextensions.k8s.io/v1',
      kind: 'CustomResourceDefinition',
      metadata: {
        name: 'avs.hourglass.io',
      },
      spec: {
        group: 'hourglass.io',
        versions: [
          {
            name: 'v1alpha1',
            served: true,
            storage: true,
            subresources: {
              status: {},
            },
            additionalPrinterColumns: [
              {
                name: 'Operator',
                type: 'string',
                jsonPath: '.spec.operator',
              },
              {
                name: 'Ready',
                type: 'string',
                jsonPath: '.status.readyReplicas',
              },
              {
                name: 'Phase',
                type: 'string',
                jsonPath: '.status.phase',
              },
            ],
            schema: {
              openAPIV3Schema: {
                type: 'object',
                properties: {
                  spec: {
                    type: 'object',
                    required: ['operator', 'serviceImage', 'replicas', 'computeRequirements'],
                    properties: {
                      operator: {
                        type: 'string',
                      },
                      serviceImage: {
                        type: 'string',
                      },
                      replicas: {
                        type: 'integer',
                        minimum: 1,
                      },
                      computeRequirements: {
                        type: 'object',
                        required: ['cpu', 'memory'],
                        properties: {
                          cpu: {
                            type: 'string',
                          },
                          memory: {
                            type: 'string',
                          },
                          teeType: {
                            type: 'string',
                            enum: ['NONE', 'MOCK', 'SEV-SNP', 'TDX', 'SGX'],
                            default: 'NONE',
                          },
                          nodeSelector: {
                            type: 'object',
                            additionalProperties: {
                              type: 'string',
                            },
                          },
                        },
                      },
                      attestationPolicy: {
                        type: 'object',
                        properties: {
                          allowedMeasurements: {
                            type: 'array',
                            items: {
                              type: 'string',
                            },
                          },
                          maxAttestationAge: {
                            type: 'string',
                          },
                          requireSEV: {
                            type: 'boolean',
                            default: false,
                          },
                        },
                      },
                      servicePort: {
                        type: 'integer',
                        default: 8080,
                      },
                    },
                  },
                  status: {
                    type: 'object',
                    properties: {
                      phase: {
                        type: 'string',
                      },
                      readyReplicas: {
                        type: 'integer',
                      },
                      totalReplicas: {
                        type: 'integer',
                      },
                      attestations: {
                        type: 'array',
                        items: {
                          type: 'object',
                          properties: {
                            podName: {
                              type: 'string',
                            },
                            instanceId: {
                              type: 'string',
                            },
                            measurement: {
                              type: 'string',
                            },
                            valid: {
                              type: 'boolean',
                            },
                            lastChecked: {
                              type: 'string',
                            },
                          },
                        },
                      },
                      lastUpdated: {
                        type: 'string',
                      },
                    },
                  },
                },
              },
            },
          },
        ],
        scope: 'Namespaced',
        names: {
          plural: 'avs',
          singular: 'avs',
          kind: 'AVS',
          shortNames: ['avs'],
        },
      },
    });

    // Create ServiceAccount for operator
    const operatorSa = cluster.addManifest('operator-sa', {
      apiVersion: 'v1',
      kind: 'ServiceAccount',
      metadata: {
        name: 'avs-operator',
        namespace: 'obsidian-system',
      },
    });

    // Create ClusterRole for operator
    const operatorRole = cluster.addManifest('operator-role', {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRole',
      metadata: {
        name: 'avs-operator-role',
      },
      rules: [
        {
          apiGroups: ['hourglass.io'],
          resources: ['avs'],
          verbs: ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete'],
        },
        {
          apiGroups: ['hourglass.io'],
          resources: ['avs/status'],
          verbs: ['get', 'update', 'patch'],
        },
        {
          apiGroups: ['apps'],
          resources: ['deployments'],
          verbs: ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete'],
        },
        {
          apiGroups: [''],
          resources: ['services', 'pods'],
          verbs: ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete'],
        },
        {
          apiGroups: [''],
          resources: ['configmaps', 'secrets'],
          verbs: ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete'],
        },
      ],
    });

    // Create ClusterRoleBinding
    const operatorRoleBinding = cluster.addManifest('operator-rolebinding', {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRoleBinding',
      metadata: {
        name: 'avs-operator-rolebinding',
      },
      roleRef: {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'ClusterRole',
        name: 'avs-operator-role',
      },
      subjects: [
        {
          kind: 'ServiceAccount',
          name: 'avs-operator',
          namespace: 'obsidian-system',
        },
      ],
    });

    // Create Operator Deployment
    const operatorDeployment = cluster.addManifest('operator-deployment', {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: {
        name: 'avs-operator',
        namespace: 'obsidian-system',
      },
      spec: {
        replicas: 1,
        selector: {
          matchLabels: {
            app: 'avs-operator',
          },
        },
        template: {
          metadata: {
            labels: {
              app: 'avs-operator',
            },
          },
          spec: {
            serviceAccountName: 'avs-operator',
            containers: [
              {
                name: 'operator',
                image: 'obsidian-operator:latest',
                command: ['/manager'],
                args: ['--leader-elect'],
                env: [
                  {
                    name: 'WATCH_NAMESPACE',
                    value: '', // Watch all namespaces
                  },
                  {
                    name: 'ENABLE_TEE',
                    value: 'false', // Disable TEE for this deployment
                  },
                ],
                resources: {
                  requests: {
                    cpu: '100m',
                    memory: '128Mi',
                  },
                  limits: {
                    cpu: '500m',
                    memory: '512Mi',
                  },
                },
              },
            ],
          },
        },
      },
    });

    // Create example AVS without TEE
    const exampleAvs = cluster.addManifest('example-avs', {
      apiVersion: 'hourglass.io/v1alpha1',
      kind: 'AVS',
      metadata: {
        name: 'example-avs',
        namespace: 'default',
      },
      spec: {
        operator: 'operator-1',
        serviceImage: 'obsidian-service:latest',
        replicas: 3,
        computeRequirements: {
          cpu: '500m',
          memory: '1Gi',
          teeType: 'NONE', // No TEE requirement
          nodeSelector: {
            'workload-type': 'general',
          },
        },
        attestationPolicy: {
          allowedMeasurements: [
            '0000000000000000000000000000000000000000000000000000000000000000', // Mock measurement
          ],
          maxAttestationAge: '1h',
          requireSEV: false,
        },
        servicePort: 8080,
      },
    });

    // Create monitoring AVS for demo
    const monitoringAvs = cluster.addManifest('monitoring-avs', {
      apiVersion: 'hourglass.io/v1alpha1',
      kind: 'AVS',
      metadata: {
        name: 'monitoring-avs',
        namespace: 'default',
      },
      spec: {
        operator: 'operator-2',
        serviceImage: 'obsidian-service:latest',
        replicas: 2,
        computeRequirements: {
          cpu: '250m',
          memory: '512Mi',
          teeType: 'MOCK', // Use mock attestation
          nodeSelector: {
            'workload-type': 'general',
          },
        },
        attestationPolicy: {
          allowedMeasurements: [
            '0000000000000000000000000000000000000000000000000000000000000000',
          ],
          maxAttestationAge: '2h',
          requireSEV: false,
        },
        servicePort: 8080,
      },
    });

    // Add dependencies
    operatorSa.node.addDependency(namespace);
    operatorRole.node.addDependency(avsCrd);
    operatorRoleBinding.node.addDependency(operatorRole);
    operatorRoleBinding.node.addDependency(operatorSa);
    operatorDeployment.node.addDependency(operatorRoleBinding);
    exampleAvs.node.addDependency(avsCrd);
    monitoringAvs.node.addDependency(avsCrd);

    // Outputs
    new cdk.CfnOutput(this, 'ObsidianOperatorStatus', {
      value: 'kubectl get deployment -n obsidian-system avs-operator',
      description: 'Command to check operator status',
    });

    new cdk.CfnOutput(this, 'ListAVS', {
      value: 'kubectl get avs -A',
      description: 'Command to list all AVS resources',
    });
  }
}