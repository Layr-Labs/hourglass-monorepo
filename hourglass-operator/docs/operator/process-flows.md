# Process Flow Diagrams

This document contains detailed process flow diagrams showing how the Hourglass Kubernetes Operator manages resource lifecycles.

## Overall System Flow

```mermaid
graph TB
    subgraph "User Actions"
        UA[User applies CR]
        UU[User updates CR]
        UD[User deletes CR]
    end
    
    subgraph "Kubernetes Control Plane"
        API[Kubernetes API Server]
        ETCD[(etcd)]
        SCHED[Kubernetes Scheduler]
    end
    
    subgraph "Hourglass Operator"
        OP[Operator Manager]
        EC[Executor Controller]
        PC[Performer Controller]
        WQ[Work Queue]
    end
    
    subgraph "Worker Nodes"
        subgraph "Node 1"
            EP1[Executor Pod 1]
            EP2[Executor Pod 2]
        end
        
        subgraph "Node 2 (GPU)"
            PP1[GPU Performer Pod]
        end
        
        subgraph "Node 3 (TEE)"
            PP2[TEE Performer Pod]
        end
    end
    
    subgraph "External Services"
        AGG[Aggregator Service]
        ETH[Ethereum L1]
        BASE[Base L2]
    end
    
    UA --> API
    UU --> API
    UD --> API
    
    API --> ETCD
    API --> OP
    
    OP --> WQ
    WQ --> EC
    WQ --> PC
    
    EC --> API
    PC --> API
    
    API --> SCHED
    SCHED --> EP1
    SCHED --> EP2
    SCHED --> PP1
    SCHED --> PP2
    
    EP1 --> AGG
    EP2 --> AGG
    EP1 --> ETH
    EP2 --> ETH
    EP1 --> BASE
    EP2 --> BASE
    
    EP1 -.->|gRPC| PP1
    EP2 -.->|gRPC| PP1
    EP1 -.->|gRPC| PP2
    EP2 -.->|gRPC| PP2
```

## HourglassExecutor Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Pending
    
    Pending --> Creating : Controller starts reconciliation
    Creating --> ConfigMapCreation : Create/Update ConfigMap
    ConfigMapCreation --> DeploymentCreation : ConfigMap ready
    DeploymentCreation --> PodScheduling : Deployment created
    PodScheduling --> PodCreation : Pods scheduled
    PodCreation --> Running : All pods ready
    
    Running --> Updating : Spec changes detected
    Updating --> ConfigMapUpdate : Update configuration
    ConfigMapUpdate --> RollingUpdate : ConfigMap updated
    RollingUpdate --> Running : Rolling update complete
    
    Running --> Scaling : Replica count changes
    Scaling --> Running : Scaling complete
    
    Running --> Terminating : Resource deleted
    Terminating --> Cleanup : Finalizer processing
    Cleanup --> [*] : Resources cleaned up
    
    Creating --> Failed : Creation error
    Updating --> Failed : Update error
    Scaling --> Failed : Scaling error
    Failed --> Creating : Retry
```

## Performer Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Pending
    
    Pending --> Creating : Controller starts reconciliation
    Creating --> PodCreation : Create pod with constraints
    PodCreation --> Scheduling : Pod submitted to scheduler
    Scheduling --> NodeSelection : Evaluate scheduling constraints
    NodeSelection --> HardwareValidation : Check hardware requirements
    HardwareValidation --> PodBinding : Bind to suitable node
    PodBinding --> ImagePull : Pull container image
    ImagePull --> ContainerStart : Start containers
    ContainerStart --> ServiceCreation : Create service
    ServiceCreation --> Running : Service ready
    
    Running --> Upgrading : Version/image changes
    Upgrading --> PodRecreation : Recreate pod
    PodRecreation --> Running : Upgrade complete
    
    Running --> Terminating : Resource deleted
    Terminating --> ServiceCleanup : Delete service
    ServiceCleanup --> PodCleanup : Delete pod
    PodCleanup --> [*] : Cleanup complete
    
    NodeSelection --> Failed : No suitable nodes
    HardwareValidation --> Failed : Hardware unavailable
    ImagePull --> Failed : Image pull error
    ContainerStart --> Failed : Container startup error
    Failed --> Creating : Retry with backoff
```

## Controller Reconciliation Loop

```mermaid
flowchart TD
    Start([Reconcile Request]) --> GetResource[Get Resource from API]
    GetResource --> ResourceExists{Resource Exists?}
    
    ResourceExists -->|No| End([End - Resource Deleted])
    ResourceExists -->|Yes| CheckDeletion{DeletionTimestamp Set?}
    
    CheckDeletion -->|Yes| HandleDeletion[Handle Deletion]
    HandleDeletion --> RemoveFinalizer[Remove Finalizer]
    RemoveFinalizer --> End
    
    CheckDeletion -->|No| AddFinalizer[Add Finalizer if Missing]
    AddFinalizer --> ReconcileResources[Reconcile Owned Resources]
    
    subgraph "Resource Reconciliation"
        ReconcileResources --> ConfigMap[ConfigMap/Pod Creation]
        ConfigMap --> Deployment[Deployment/Service Creation]
        Deployment --> OwnerRef[Set Owner References]
        OwnerRef --> StatusUpdate[Update Resource Status]
    end
    
    StatusUpdate --> CheckRequeue{Need Requeue?}
    CheckRequeue -->|Yes| RequeueAfter[Requeue After Interval]
    CheckRequeue -->|No| End
    RequeueAfter --> End
    
    GetResource -->|Error| ErrorHandler[Handle API Error]
    ErrorHandler --> RetryDecision{Retry?}
    RetryDecision -->|Yes| RequeueImmediate[Immediate Requeue]
    RetryDecision -->|No| End
    RequeueImmediate --> End
```

## Advanced Scheduling Flow

```mermaid
sequenceDiagram
    participant User
    participant API as Kubernetes API
    participant Controller as Performer Controller
    participant Scheduler as K8s Scheduler
    participant Node1 as Regular Node
    participant Node2 as GPU Node
    participant Node3 as TEE Node
    
    User->>API: Apply Performer with GPU requirements
    API->>Controller: Watch event triggered
    Controller->>API: Get Performer resource
    Controller->>Controller: Build pod specification
    
    Note over Controller: Apply nodeSelector, affinity, tolerations, hardware requirements
    
    Controller->>API: Create Pod with scheduling constraints
    API->>Scheduler: Schedule pod
    
    Scheduler->>Node1: Check node compatibility
    Node1-->>Scheduler: No GPU available
    
    Scheduler->>Node2: Check node compatibility
    Node2-->>Scheduler: GPU available, constraints satisfied
    
    Scheduler->>API: Bind pod to Node2
    API->>Node2: Create pod
    Node2->>Node2: Pull image and start container
    
    Node2-->>API: Pod status update
    API->>Controller: Pod status change event
    Controller->>API: Update Performer status
    
    Controller->>API: Create Service
    API->>Controller: Service created
    Controller->>API: Update Performer with gRPC endpoint
```

## Hardware Requirements Processing

```mermaid
flowchart TD
    Start([Performer with Hardware Requirements]) --> ParseHW[Parse Hardware Requirements]
    
    ParseHW --> CheckGPU{GPU Required?}
    CheckGPU -->|Yes| SetGPUResources[Set GPU Resource Requests/Limits]
    CheckGPU -->|No| CheckTEE
    
    SetGPUResources --> GPULabels[Add GPU Node Selectors]
    GPULabels --> GPUTolerations[Add GPU Tolerations]
    GPUTolerations --> CheckTEE
    
    CheckTEE{TEE Required?} -->|Yes| SetTEELabels[Add TEE Node Selectors]
    CheckTEE -->|No| CheckCustom
    
    SetTEELabels --> TEETolerations[Add TEE Tolerations]
    TEETolerations --> RuntimeClass[Set Runtime Class]
    RuntimeClass --> CheckCustom
    
    CheckCustom{Custom Labels?} -->|Yes| AddCustomLabels[Add Custom Node Selectors]
    CheckCustom -->|No| BuildAffinity
    
    AddCustomLabels --> BuildAffinity[Build Node Affinity Rules]
    BuildAffinity --> CreatePod[Create Pod with All Constraints]
    
    CreatePod --> SchedulerEval[Scheduler Evaluates Constraints]
    SchedulerEval --> NodeMatch{Matching Node Found?}
    
    NodeMatch -->|Yes| PodScheduled[Pod Scheduled Successfully]
    NodeMatch -->|No| PodPending[Pod Remains Pending]
    
    PodScheduled --> End([Performer Running])
    PodPending --> Retry[Retry Scheduling]
    Retry --> SchedulerEval
```

## Service Discovery Flow

```mermaid
sequenceDiagram
    participant Executor as Executor Pod
    participant DNS as Cluster DNS
    participant Service as Performer Service
    participant Pod as Performer Pod
    participant Controller as Performer Controller
    
    Note over Controller: Performer controller creates service
    Controller->>Service: Create service with stable name
    Controller->>Pod: Create performer pod
    Controller->>Controller: Update Performer status with gRPC endpoint
    
    Note over Executor: Executor needs to connect to performer
    Executor->>DNS: Resolve performer-{name}.{namespace}.svc.cluster.local
    DNS-->>Executor: Return service IP
    
    Executor->>Service: Connect via gRPC (port 9090)
    Service->>Pod: Route traffic to performer pod
    Pod-->>Service: gRPC response
    Service-->>Executor: Forward response
    
    Note over Executor,Pod: Persistent gRPC connection established
    
    loop Continuous Communication
        Executor->>Service: Send task requests
        Service->>Pod: Forward to performer
        Pod-->>Service: Task responses
        Service-->>Executor: Forward responses
    end
```

## Error Handling and Recovery

```mermaid
flowchart TD
    Error[Error Detected] --> ClassifyError{Error Type?}
    
    ClassifyError -->|API Error| APIRetry[Exponential Backoff Retry]
    ClassifyError -->|Resource Conflict| ConflictRetry[Immediate Retry]
    ClassifyError -->|Validation Error| LogError[Log Error and Stop]
    ClassifyError -->|Node Issues| RescheduleAttempt[Attempt Rescheduling]
    
    APIRetry --> CheckRetryCount{Max Retries Reached?}
    CheckRetryCount -->|No| WaitBackoff[Wait Backoff Period]
    CheckRetryCount -->|Yes| LogError
    WaitBackoff --> RetryOperation[Retry Operation]
    RetryOperation --> Success{Operation Successful?}
    Success -->|Yes| UpdateStatus[Update Status to Healthy]
    Success -->|No| Error
    
    ConflictRetry --> RetryImmediate[Retry Immediately]
    RetryImmediate --> Success
    
    RescheduleAttempt --> DeletePod[Delete Current Pod]
    DeletePod --> CreateNewPod[Create New Pod]
    CreateNewPod --> Success
    
    LogError --> UpdateStatusFailed[Update Status to Failed]
    UpdateStatus --> End([Recovery Complete])
    UpdateStatusFailed --> End
```

## Multi-Chain Executor Configuration

```mermaid
flowchart LR
    subgraph "HourglassExecutor"
        Config[Executor Config]
        Chain1[Ethereum Chain Config]
        Chain2[Base Chain Config]
        Chain3[Arbitrum Chain Config]
    end
    
    subgraph "Generated ConfigMap"
        YAML[executor-config.yaml]
    end
    
    subgraph "Executor Pod"
        Process[Executor Process]
        ETHClient[Ethereum Client]
        BaseClient[Base Client]
        ArbClient[Arbitrum Client]
    end
    
    subgraph "External Networks"
        ETH[Ethereum Mainnet]
        BASE[Base Network]
        ARB[Arbitrum One]
    end
    
    Config --> YAML
    Chain1 --> YAML
    Chain2 --> YAML
    Chain3 --> YAML
    
    YAML -->|Mount| Process
    Process --> ETHClient
    Process --> BaseClient
    Process --> ArbClient
    
    ETHClient --> ETH
    BaseClient --> BASE
    ArbClient --> ARB
    
    ETH -.->|Events| ETHClient
    BASE -.->|Events| BaseClient
    ARB -.->|Events| ArbClient
```

These diagrams provide a comprehensive view of how the Hourglass Kubernetes Operator orchestrates complex workflows involving multiple chains, specialized hardware, and service discovery patterns.