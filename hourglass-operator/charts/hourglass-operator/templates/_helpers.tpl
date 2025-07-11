{{/*
Expand the name of the chart.
*/}}
{{- define "hourglass-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "hourglass-operator.fullname" -}}
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
Create chart name and version as used by the chart label.
*/}}
{{- define "hourglass-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "hourglass-operator.labels" -}}
helm.sh/chart: {{ include "hourglass-operator.chart" . }}
{{ include "hourglass-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: hourglass
component: operator
{{- end }}

{{/*
Selector labels
*/}}
{{- define "hourglass-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hourglass-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "hourglass-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "hourglass-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the ClusterRole
*/}}
{{- define "hourglass-operator.clusterRoleName" -}}
{{- printf "%s-manager-role" (include "hourglass-operator.fullname" .) }}
{{- end }}

{{/*
Create the name of the ClusterRoleBinding
*/}}
{{- define "hourglass-operator.clusterRoleBindingName" -}}
{{- printf "%s-manager-rolebinding" (include "hourglass-operator.fullname" .) }}
{{- end }}

{{/*
Create the name of the service
*/}}
{{- define "hourglass-operator.serviceName" -}}
{{- printf "%s-service" (include "hourglass-operator.fullname" .) }}
{{- end }}

{{/*
Create the name of the webhook service
*/}}
{{- define "hourglass-operator.webhookServiceName" -}}
{{- printf "%s-webhook-service" (include "hourglass-operator.fullname" .) }}
{{- end }}

{{/*
Create the name of the webhook certificate
*/}}
{{- define "hourglass-operator.webhookCertName" -}}
{{- printf "%s-webhook-cert" (include "hourglass-operator.fullname" .) }}
{{- end }}

{{/*
Create the name of the webhook issuer
*/}}
{{- define "hourglass-operator.webhookIssuerName" -}}
{{- printf "%s-webhook-issuer" (include "hourglass-operator.fullname" .) }}
{{- end }}