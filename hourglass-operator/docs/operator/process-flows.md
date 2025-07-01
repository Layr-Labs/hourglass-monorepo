# Process Flow Diagrams

This document contains detailed process flow diagrams showing how the **singleton** Hourglass Kubernetes Operator manages resource lifecycles across multiple user-deployed Executors.

## Overall System Flow (Singleton Architecture)

```mermaid
graph TB
    subgraph "User Actions"
        UA[User deploys Executor StatefulSet]
        UU[User updates Executor]
        UP[User creates/updates Performer]
        UD[User deletes resources]
    end
    
    subgraph "Kubernetes Control Plane"
        API[Kubernetes API Server]
        ETCD[(etcd)]
        SCHED[Kubernetes Scheduler]
    end
    
    subgraph "Singleton Operator"
        OP[Operator Manager]
        PC[Performer Controller Only]
        WQ[Work Queue]
    end
    
    subgraph "Namespace: avs-project-a"
        subgraph "User-Managed Resources A"
            EA[Executor StatefulSet A]
            EPA[Executor Pod A]
        end
        
        subgraph "Operator-Managed Resources A"
            PA1[Performer A1]
            PA2[Performer A2]
            PPA1[Performer Pod A1]
            PPA2[Performer Pod A2]
            PSA1[Performer Service A1]
            PSA2[Performer Service A2]
        end
    end
    
    subgraph "Namespace: avs-project-b"
        subgraph "User-Managed Resources B"
            EB[Executor StatefulSet B]
            EPB[Executor Pod B]
        end
        
        subgraph "Operator-Managed Resources B"
            PB1[Performer B1]
            PPB1[Performer Pod B1]
            PSB1[Performer Service B1]
        end
    end
    
    subgraph "External Services"
        AGG[Aggregator Service]
        ETH[Ethereum L1]
        BASE[Base L2]
    end
    
    UA --> API
    UU --> API
    UP --> API
    UD --> API
    
    API --> ETCD
    API --> OP
    
    OP --> WQ
    WQ --> PC
    PC --> API
    
    API --> SCHED
    SCHED --> EPA
    SCHED --> EPB
    SCHED --> PPA1
    SCHED --> PPA2
    SCHED --> PPB1
    
    EPA -->|Creates| PA1
    EPA -->|Creates| PA2
    EPB -->|Creates| PB1
    
    PC -->|Watches All| PA1
    PC -->|Watches All| PA2
    PC -->|Watches All| PB1
    
    PC -->|Manages| PPA1
    PC -->|Manages| PPA2
    PC -->|Manages| PPB1
    PC -->|Manages| PSA1
    PC -->|Manages| PSA2
    PC -->|Manages| PSB1
    
    EPA --> AGG
    EPB --> AGG
    EPA --> ETH
    EPB --> ETH
    EPA --> BASE
    EPB --> BASE
    
    EPA -.->|gRPC| PSA1
    EPA -.->|gRPC| PSA2
    EPB -.->|gRPC| PSB1
    
    PSA1 --> PPA1
    PSA2 --> PPA2
    PSB1 --> PPB1
```

## User-Managed Executor Lifecycle

```mermaid
stateDiagram-v2
    [*] --> UserDeploys
    
    UserDeploys --> StatefulSetCreated : User applies StatefulSet YAML
    StatefulSetCreated --> PodScheduling : K8s schedules pods
    PodScheduling --> PodStartup : Pods starting
    PodStartup --> ExecutorRunning : Executor process ready
    
    ExecutorRunning --> PerformerCreation : Executor creates Performer CRDs
    PerformerCreation --> OperatorWatch : Singleton operator watches CRDs
    OperatorWatch --> PerformerManagement : Operator manages performer resources
    PerformerManagement --> FullyOperational : System fully operational
    
    FullyOperational --> UserUpdates : User updates Executor config
    UserUpdates --> RollingUpdate : StatefulSet rolling update
    RollingUpdate --> FullyOperational : Update complete
    
    FullyOperational --> UserScales : User scales replicas
    UserScales --> FullyOperational : Scaling complete
    
    FullyOperational --> UserDeletes : User deletes StatefulSet
    UserDeletes --> PerformerCleanup : Executor cleans up Performers
    PerformerCleanup --> [*] : Resources cleaned up
    
    PodScheduling --> Failed : Scheduling failure
    PodStartup --> Failed : Startup failure
    Failed --> UserFixes : User fixes configuration
    UserFixes --> StatefulSetCreated : Redeploy
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

## Singleton Performer Controller Reconciliation Loop

```mermaid
flowchart TD
    Start([Performer Reconcile Request<br/>Any Namespace]) --> GetPerformer[Get Performer from Namespace]
    GetPerformer --> ResourceExists{Performer Exists?}
    
    ResourceExists -->|No| End([End - Performer Deleted])
    ResourceExists -->|Yes| CheckDeletion{DeletionTimestamp Set?}
    
    CheckDeletion -->|Yes| HandleDeletion[Handle Deletion in Namespace]
    HandleDeletion --> CleanupPod[Delete Performer Pod]
    CleanupPod --> CleanupService[Delete Performer Service]
    CleanupService --> RemoveFinalizer[Remove Finalizer]
    RemoveFinalizer --> End
    
    CheckDeletion -->|No| AddFinalizer[Add Finalizer if Missing]
    AddFinalizer --> ReconcileResources[Reconcile Performer Resources]
    
    subgraph "Namespace-Scoped Resource Reconciliation"
        ReconcileResources --> PodCreation[Create/Update Performer Pod]
        PodCreation --> ApplyScheduling[Apply Scheduling Constraints]
        ApplyScheduling --> ApplyHardware[Apply Hardware Requirements]
        ApplyHardware --> ServiceCreation[Create/Update Performer Service]
        ServiceCreation --> OwnerRef[Set Owner References]
        OwnerRef --> StatusUpdate[Update Performer Status]
    end
    
    StatusUpdate --> CheckRequeue{Need Requeue?}
    CheckRequeue -->|Yes| RequeueAfter[Requeue After 2 Minutes]
    CheckRequeue -->|No| End
    RequeueAfter --> End
    
    GetPerformer -->|Error| ErrorHandler[Handle API Error]
    ErrorHandler --> RetryDecision{Retry?}
    RetryDecision -->|Yes| RequeueImmediate[Immediate Requeue]
    RetryDecision -->|No| End
    RequeueImmediate --> End
```

## Advanced Scheduling Flow (Singleton Operator)

```mermaid
sequenceDiagram
    participant Executor as User Executor Pod
    participant API as Kubernetes API
    participant Controller as Singleton Performer Controller
    participant Scheduler as K8s Scheduler
    participant Node1 as Regular Node
    participant Node2 as GPU Node
    participant Node3 as TEE Node
    
    Note over Executor: User-deployed Executor needs GPU performer
    Executor->>API: Create Performer CRD with GPU requirements
    API->>Controller: Watch event triggered (any namespace)
    Controller->>API: Get Performer resource from namespace
    Controller->>Controller: Build pod specification with constraints
    
    Note over Controller: Apply nodeSelector, GPU affinity, tolerations, hardware requirements
    
    Controller->>API: Create Pod in performer's namespace
    API->>Scheduler: Schedule pod with constraints
    
    Scheduler->>Node1: Check node compatibility
    Node1-->>Scheduler: No GPU available
    
    Scheduler->>Node2: Check node compatibility  
    Node2-->>Scheduler: GPU available, constraints satisfied
    
    Scheduler->>API: Bind pod to Node2
    API->>Node2: Create pod in namespace
    Node2->>Node2: Pull image and start container
    
    Node2-->>API: Pod status update
    API->>Controller: Pod status change event
    Controller->>API: Update Performer status in namespace
    
    Controller->>API: Create Service in namespace
    API->>Controller: Service created
    Controller->>API: Update Performer with gRPC endpoint
    
    Note over Executor: Connect to performer via stable DNS
    Executor->>API: Query Performer status
    API-->>Executor: Performer ready, endpoint available
    Executor->>Node2: Connect via performer-{name}.{namespace}.svc.cluster.local
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

## Service Discovery Flow (Multi-Namespace)

```mermaid
sequenceDiagram
    participant ExecutorA as Executor A<br/>(Namespace A)
    participant ExecutorB as Executor B<br/>(Namespace B)
    participant DNS as Cluster DNS
    participant ServiceA as Performer Service A<br/>(Namespace A)
    participant ServiceB as Performer Service B<br/>(Namespace B)
    participant PodA as Performer Pod A<br/>(Namespace A)
    participant PodB as Performer Pod B<br/>(Namespace B)
    participant Controller as Singleton Performer Controller
    
    Note over Controller: Singleton controller manages all performers
    Controller->>ServiceA: Create service with stable name in Namespace A
    Controller->>PodA: Create performer pod in Namespace A
    Controller->>ServiceB: Create service with stable name in Namespace B
    Controller->>PodB: Create performer pod in Namespace B
    Controller->>Controller: Update Performer statuses with gRPC endpoints
    
    Note over ExecutorA: Executor A connects to its performers
    ExecutorA->>DNS: Resolve performer-a1.namespace-a.svc.cluster.local
    DNS-->>ExecutorA: Return service A IP
    
    Note over ExecutorB: Executor B connects to its performers
    ExecutorB->>DNS: Resolve performer-b1.namespace-b.svc.cluster.local
    DNS-->>ExecutorB: Return service B IP
    
    ExecutorA->>ServiceA: Connect via gRPC (port 9090)
    ServiceA->>PodA: Route traffic to performer pod A
    PodA-->>ServiceA: gRPC response
    ServiceA-->>ExecutorA: Forward response
    
    ExecutorB->>ServiceB: Connect via gRPC (port 9090)
    ServiceB->>PodB: Route traffic to performer pod B
    PodB-->>ServiceB: gRPC response
    ServiceB-->>ExecutorB: Forward response
    
    Note over ExecutorA,PodA: Namespace A communication
    Note over ExecutorB,PodB: Namespace B communication
    
    loop Continuous Communication (Namespace A)
        ExecutorA->>ServiceA: Send task requests
        ServiceA->>PodA: Forward to performer A
        PodA-->>ServiceA: Task responses
        ServiceA-->>ExecutorA: Forward responses
    end
    
    loop Continuous Communication (Namespace B)
        ExecutorB->>ServiceB: Send task requests
        ServiceB->>PodB: Forward to performer B
        PodB-->>ServiceB: Task responses
        ServiceB-->>ExecutorB: Forward responses
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

## Multi-Namespace Deployment Flow

```mermaid
sequenceDiagram
    participant Admin as Cluster Admin
    participant UserA as AVS Team A
    participant UserB as AVS Team B
    participant API as Kubernetes API
    participant Operator as Singleton Operator
    participant NSA as Namespace A
    participant NSB as Namespace B
    
    Note over Admin: Deploy singleton operator once
    Admin->>API: Deploy operator to hourglass-system namespace
    API->>Operator: Operator starts watching cluster-wide
    
    Note over UserA: Deploy AVS project A
    UserA->>API: Create namespace avs-project-a
    UserA->>NSA: Deploy Executor StatefulSet A
    UserA->>NSA: Deploy Executor config and secrets
    NSA->>API: Executor A creates Performer CRDs
    API->>Operator: Watch events from namespace A
    Operator->>NSA: Create Performer pods and services in namespace A
    
    Note over UserB: Deploy AVS project B (independent)
    UserB->>API: Create namespace avs-project-b
    UserB->>NSB: Deploy Executor StatefulSet B
    UserB->>NSB: Deploy Executor config and secrets
    NSB->>API: Executor B creates Performer CRDs
    API->>Operator: Watch events from namespace B
    Operator->>NSB: Create Performer pods and services in namespace B
    
    Note over Operator: Single operator manages both namespaces
    Operator->>Operator: Reconcile performers from both namespaces
    
    Note over NSA,NSB: Isolated operations
    NSA->>NSA: Executor A ↔ Performer A communication
    NSB->>NSB: Executor B ↔ Performer B communication
```

## Singleton Operator Benefits Flow

```mermaid
graph TB
    subgraph "Traditional Architecture"
        TA[Executor A + Operator A]
        TB[Executor B + Operator B]
        TC[Executor C + Operator C]
        
        TA --> RA[Resource Usage A]
        TB --> RB[Resource Usage B]
        TC --> RC[Resource Usage C]
        
        RA --> TotalOld[Total: 3x Operator Overhead]
        RB --> TotalOld
        RC --> TotalOld
    end
    
    subgraph "Singleton Architecture"
        SA[User Executor A]
        SB[User Executor B] 
        SC[User Executor C]
        SO[Single Operator]
        
        SA --> SO
        SB --> SO
        SC --> SO
        
        SO --> RSingle[Minimal Operator Overhead]
        SA --> RANew[User Executor Resources A]
        SB --> RBNew[User Executor Resources B]
        SC --> RCNew[User Executor Resources C]
        
        RSingle --> TotalNew[Total: 1x Operator + User Control]
        RANew --> TotalNew
        RBNew --> TotalNew
        RCNew --> TotalNew
    end
    
    TotalOld -->|Efficiency Gain| TotalNew
    
    subgraph "Benefits"
        B1[Reduced Resource Usage]
        B2[User Autonomy]
        B3[Centralized Management]
        B4[Namespace Isolation]
        B5[Easier Maintenance]
    end
    
    TotalNew --> B1
    TotalNew --> B2
    TotalNew --> B3
    TotalNew --> B4
    TotalNew --> B5
```

These diagrams provide a comprehensive view of how the **singleton** Hourglass Kubernetes Operator orchestrates complex workflows across multiple user-managed executors, specialized hardware scheduling, and multi-namespace service discovery patterns while maintaining operational efficiency and user autonomy.