{{/*
Expand the name of the chart.
*/}}
{{- define "hourglass-executor.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "hourglass-executor.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create the executor name (can be overridden by executor.name)
*/}}
{{- define "hourglass-executor.executorName" -}}
{{- if .Values.executor.name }}
{{- .Values.executor.name }}
{{- else }}
{{- include "hourglass-executor.fullname" . }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "hourglass-executor.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "hourglass-executor.labels" -}}
helm.sh/chart: {{ include "hourglass-executor.chart" . }}
{{ include "hourglass-executor.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: hourglass
component: executor
{{- end }}

{{/*
Selector labels
*/}}
{{- define "hourglass-executor.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hourglass-executor.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app: {{ include "hourglass-executor.executorName" . }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "hourglass-executor.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "hourglass-executor.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the secrets
*/}}
{{- define "hourglass-executor.secretName" -}}
{{- printf "%s-keys" (include "hourglass-executor.executorName" .) }}
{{- end }}

{{/*
Create the name of the configmap
*/}}
{{- define "hourglass-executor.configMapName" -}}
{{- printf "%s-config" (include "hourglass-executor.executorName" .) }}
{{- end }}

{{/*
Create the name of the TLS secret
*/}}
{{- define "hourglass-executor.tlsSecretName" -}}
{{- printf "%s-tls" (include "hourglass-executor.executorName" .) }}
{{- end }}

{{/*
Create the name of the registry secret
*/}}
{{- define "hourglass-executor.registrySecretName" -}}
{{- if .Values.secrets.imagePullSecrets.name }}
{{- .Values.secrets.imagePullSecrets.name }}
{{- else }}
{{- printf "%s-registry-secret" (include "hourglass-executor.executorName" .) }}
{{- end }}
{{- end }}

{{/*
Generate the executor configuration YAML
*/}}
{{- define "hourglass-executor.config" -}}
# Hourglass Executor Configuration
aggregator_endpoint: {{ .Values.aggregator.endpoint | quote }}
aggregator_tls_enabled: {{ .Values.aggregator.tls.enabled }}
aggregator_timeout: {{ .Values.aggregator.timeout | quote }}
deployment_mode: {{ .Values.executor.env.deploymentMode | quote }}
log_level: {{ .Values.executor.env.logLevel | quote }}
log_format: {{ .Values.executor.env.logFormat | quote }}

performer_config:
  service_pattern: {{ .Values.performer.servicePattern | quote }}
  default_port: {{ .Values.performer.defaultPort }}
  connection_timeout: {{ .Values.performer.connectionTimeout | quote }}
  startup_timeout: {{ .Values.performer.startupTimeout | quote }}
  retry_attempts: {{ .Values.performer.retryAttempts }}
  max_performers: {{ .Values.performer.maxPerformers }}
  health_check_interval: {{ .Values.performer.healthCheckInterval | quote }}
  health_check_timeout: {{ .Values.performer.healthCheckTimeout | quote }}
  resource_monitoring_enabled: {{ .Values.performer.resourceMonitoringEnabled }}
  resource_monitoring_interval: {{ .Values.performer.resourceMonitoringInterval | quote }}

kubernetes:
  namespace: {{ .Release.Namespace | quote }}
  operator_namespace: {{ .Values.global.operatorNamespace | quote }}
  performer_crd_group: "hourglass.eigenlayer.io"
  performer_crd_version: "v1alpha1"
  performer_crd_kind: "Performer"
  service_account: {{ include "hourglass-executor.serviceAccountName" . | quote }}
  cleanup_on_shutdown: {{ .Values.cleanup.onShutdown }}
  cleanup_timeout: {{ .Values.cleanup.timeout | quote }}

chains:
{{- if .Values.chains.ethereum.enabled }}
  - name: "ethereum"
    chain_id: {{ .Values.chains.ethereum.chainId }}
    rpc_url: {{ .Values.chains.ethereum.rpcUrl | quote }}
    {{- if .Values.chains.ethereum.wsUrl }}
    ws_url: {{ .Values.chains.ethereum.wsUrl | quote }}
    {{- end }}
    task_mailbox_address: {{ .Values.chains.ethereum.taskMailboxAddress | quote }}
    block_confirmations: {{ .Values.chains.ethereum.blockConfirmations }}
    gas_limit: {{ .Values.chains.ethereum.gasLimit }}
    gas_price_multiplier: {{ .Values.chains.ethereum.gasPriceMultiplier }}
    event_filter:
      from_block: {{ .Values.chains.ethereum.fromBlock | quote }}
    retry_config:
      max_attempts: {{ .Values.chains.ethereum.retryAttempts }}
      backoff_duration: {{ .Values.chains.ethereum.retryBackoff | quote }}
{{- end }}
{{- if .Values.chains.base.enabled }}
  - name: "base"
    chain_id: {{ .Values.chains.base.chainId }}
    rpc_url: {{ .Values.chains.base.rpcUrl | quote }}
    {{- if .Values.chains.base.wsUrl }}
    ws_url: {{ .Values.chains.base.wsUrl | quote }}
    {{- end }}
    task_mailbox_address: {{ .Values.chains.base.taskMailboxAddress | quote }}
    block_confirmations: {{ .Values.chains.base.blockConfirmations }}
    gas_limit: {{ .Values.chains.base.gasLimit }}
    gas_price_multiplier: {{ .Values.chains.base.gasPriceMultiplier }}
    event_filter:
      from_block: {{ .Values.chains.base.fromBlock | quote }}
    retry_config:
      max_attempts: {{ .Values.chains.base.retryAttempts }}
      backoff_duration: {{ .Values.chains.base.retryBackoff | quote }}
{{- end }}

avs_config:
  supported_avs:
{{- range .Values.avs.supportedAvs }}
    - address: {{ .address | quote }}
      name: {{ .name | quote }}
      performer_image: {{ .performer.image | quote }}
      performer_version: {{ .performer.version | quote }}
      default_resources:
        requests:
          {{- if .resources.requests.cpu }}
          cpu: {{ .resources.requests.cpu | quote }}
          {{- end }}
          {{- if .resources.requests.memory }}
          memory: {{ .resources.requests.memory | quote }}
          {{- end }}
          {{- if .hardware.gpu.required }}
          nvidia.com/gpu: {{ .hardware.gpu.count | quote }}
          {{- end }}
        limits:
          {{- if .resources.limits.cpu }}
          cpu: {{ .resources.limits.cpu | quote }}
          {{- end }}
          {{- if .resources.limits.memory }}
          memory: {{ .resources.limits.memory | quote }}
          {{- end }}
          {{- if .hardware.gpu.required }}
          nvidia.com/gpu: {{ .hardware.gpu.count | quote }}
          {{- end }}
      hardware_requirements:
        gpu_required: {{ .hardware.gpu.required }}
        {{- if .hardware.gpu.type }}
        gpu_type: {{ .hardware.gpu.type | quote }}
        {{- end }}
        {{- if .hardware.gpu.count }}
        gpu_count: {{ .hardware.gpu.count }}
        {{- end }}
        tee_required: {{ .hardware.tee.required }}
        {{- if .hardware.tee.type }}
        tee_type: {{ .hardware.tee.type | quote }}
        {{- end }}
      scheduling:
        {{- if .scheduling.nodeSelector }}
        node_selector:
          {{- toYaml .scheduling.nodeSelector | nindent 10 }}
        {{- end }}
        {{- if .scheduling.tolerations }}
        tolerations:
          {{- toYaml .scheduling.tolerations | nindent 10 }}
        {{- end }}
        {{- if .scheduling.priorityClassName }}
        priority_class: {{ .scheduling.priorityClassName | quote }}
        {{- end }}
        {{- if .scheduling.runtimeClass }}
        runtime_class: {{ .scheduling.runtimeClass | quote }}
        {{- end }}
{{- end }}

operator_keys:
  ecdsa_private_key_path: "/etc/secrets/ecdsa-private-key"
  {{- if .Values.secrets.operatorKeys.blsPrivateKey }}
  bls_private_key_path: "/etc/secrets/bls-private-key"
  {{- end }}

metrics:
  enabled: {{ .Values.metrics.enabled }}
  port: {{ .Values.metrics.port }}
  path: {{ .Values.metrics.path | quote }}
  custom_metrics:
    - name: "performer_count"
      help: "Number of active performers"
      type: "gauge"
    - name: "task_processing_duration"
      help: "Time taken to process tasks"
      type: "histogram"
    - name: "performer_connection_errors"
      help: "Number of performer connection errors"
      type: "counter"

health:
  enabled: {{ .Values.healthChecks.enabled }}
  port: {{ .Values.executor.env.healthCheckPort }}
  endpoints:
    liveness: "/health"
    readiness: "/ready"
    startup: "/startup"
  checks:
    - name: "aggregator_connection"
      enabled: true
      timeout: "10s"
    - name: "chain_connections"
      enabled: true
      timeout: "15s"
    - name: "performer_status"
      enabled: true
      timeout: "5s"
    - name: "kubernetes_api"
      enabled: true
      timeout: "5s"
{{- end }}

{{/*
Validate required values
*/}}
{{- define "hourglass-executor.validateValues" -}}
{{- if not .Values.aggregator.endpoint }}
{{- fail "aggregator.endpoint is required" }}
{{- end }}
{{- if not .Values.chains.ethereum.rpcUrl }}
{{- fail "chains.ethereum.rpcUrl is required" }}
{{- end }}
{{- if not .Values.chains.ethereum.taskMailboxAddress }}
{{- fail "chains.ethereum.taskMailboxAddress is required" }}
{{- end }}
{{- if not .Values.secrets.operatorKeys.ecdsaPrivateKey }}
{{- fail "secrets.operatorKeys.ecdsaPrivateKey is required" }}
{{- end }}
{{- range .Values.avs.supportedAvs }}
{{- if not .address }}
{{- fail "avs.supportedAvs[].address is required" }}
{{- end }}
{{- if not .performer.image }}
{{- fail "avs.supportedAvs[].performer.image is required" }}
{{- end }}
{{- end }}
{{- end }}