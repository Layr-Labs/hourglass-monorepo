# ⚠️ Warning: This is Alpha, non audited code ⚠️
Hourglass is in active development and is not yet audited. Use at your own risk.

## What is Hourglass?

Hourglass is a framework for building a task-based EigenLayer AVS, providing AVS developers a batteries-included experience to get started quickly. It includes a set of tools and libraries that simplify the process of building, deploying, and managing AVS projects.

![](docs/images/hourglass-architecture_v.01.0.svg)

Hourglass as a framework has onchain and offchain components that work together to enable a task-based AVS.

### Onchain Components

#### TaskMailbox

The TaskMailbox is a singleton eigenlayer hourglass contract on L1 or L2 that is responsible for:

* Allowing users/apps to create tasks.
* Managing the lifecycle of tasks.
* Verifying the results of tasks and making it available for users/apps to query.
* Allowing AVSs to manage their TaskMailbox configurations.

#### TaskAVSRegistrar

The TaskAVSRegistrar is an instanced (per-AVS) eigenlayer middleware contract on L1 that is responsible for:

* Handling operator registration for specific operator sets of your AVS.
* Providing the offchain components with BLS public keys and socket endpoints for the Aggregator and Executor operators.

It works by default, but can be extended to include additional onchain logic for your AVS.

#### AVSTaskHook

The AVSTaskHook is an instanced (per-AVS) eigenlayer hourglass contract on L1 or L2 that is responsible for:

* Validating the task lifecycle.
* Creating fee markets for your AVS.

It's empty by default and works out of the box, but can be extended to include additional onchain validation logic for your AVS.

#### CertificateVerifier

The CertificateVerifier is an instanced (per-AVS) eigenlayer middleware contract on L1 or L2 that is responsible for:

* Verifying the validity of operator certificates.
* Verifying stake threshold requirements for operator sets.

### Offchain Components

#### Aggregator

The Aggregator is responsible for:

* Listening to events from the Mailbox contract on chain for new tasks
* Discovering Executors by querying the AVSRegistrar contract (via the EigenLayer Allocation manager), retrieving their metadata containing a BLS public key and a "socket" (url) endpoint that references the Executor's gRPC server.
* Distributing tasks to Executors by sending a gRPC request to the Executor's socket endpoint, including the task payload and a signature of the payload signed by the Aggregator. This is so the Executor can validate the message is coming from the expected Aggregator.
* Aggregates results from Executors until a signing threshold has been met
* Publish the result back to the Mailbox contract

#### Executor

The Executor is responsible for:
* Launching and managing Performer containers that execute the tasks
* Listening to gRPC requests from the Aggregator for new tasks
* Forwarding the task to the correct Performer
* Signing the result of the task with its BLS private key and sending it back to the Aggregator


#### Performer

The Performer is the component the AVS is responsible for building. At a high level, it is a simple gRPC server that listens for tasks, runs them and returns the results to the Executor.

The Hourglass framework provides all of the boilerplate and server code for your Performer; you simply need to fill in the logic to handle tasks for your AVS!
