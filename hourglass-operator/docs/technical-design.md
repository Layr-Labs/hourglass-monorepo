# HourglassOperator

The HourglassOperator is a custom Kubernetes Operator designed to facilitate running the Hourglass Executor and its Performer containers at scale on a kubnernetes cluster.

Currently, the Executor spawns Performer containers by talking directly with the Docker daemon. This works for
single-host deployments, but for the purpose of project BlastPad (EigenCompute MVP), we need the ability to run the Executor
as one pod and have the Performer containers spawned on other nodes in the cluster that are more secure (running Bottlerocket).

The HourglassOperator will be responsible for managing the lifecycle of the Performer containers, taking commands from the Executor
to know what containers to run for an AVS and when to upgrade to the next version. 

Performers will still run as containers, but as pods within the cluster, exposing their gRPC interface for the executor
to interact with. Performers are long-running processes, not one-off jobs. They always need to exist until the Executor decides 
to terminate them or upgrade the container version. Performers should have a service in front of the pod to serve as the entrypoint
for the Executor to connect to using a kubedns name that is configured as part of the operator config.

Deploying the Hourglass Operator should include the ability to deploy the executor and then take input from the executor to 
deploy Performer pods, this way a single "Executor Operator" can encapsulate the entire deployment and management of the Hourglass system.
