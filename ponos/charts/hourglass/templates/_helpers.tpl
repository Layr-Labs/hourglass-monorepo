{{- define "hourglass.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "aggregator.name" -}}
{{- default "aggregator" .Values.aggregator.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "aggregator.configSecretName" -}}
{{ include "aggregator.name" . }}-config-secret
{{- end }}

{{- define "aggregator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "aggregator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "aggregator.labels" -}}
helm.sh/chart: {{ include "hourglass.chart" . }}
{{ include "aggregator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}
{{- with .Values.aggregator.additionalLabels }}
{{- toYaml . | nindent 0 }}
{{- end }}

{{- define "aggregator.metadataLabels" -}}
{{ include "aggregator.selectorLabels" . }}
{{- with .Values.aggregator.metadataLabels }}
{{- toYaml . | nindent 0 }}
{{- end }}
{{- end }}



{{- define "executor.name" -}}
{{- default "executor" .Values.executor.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "executor.configSecretName" -}}
{{ include "executor.name" . }}-config-secret
{{- end }}

{{- define "executor.selectorLabels" -}}
app.kubernetes.io/name: {{ include "executor.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "executor.labels" -}}
helm.sh/chart: {{ include "hourglass.chart" . }}
{{ include "executor.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}
{{- with .Values.executor.additionalLabels }}
{{- toYaml . | nindent 0 }}
{{- end }}

{{- define "executor.metadataLabels" -}}
{{ include "executor.selectorLabels" . }}
{{- with .Values.executor.metadataLabels }}
{{- toYaml . | nindent 0 }}
{{- end }}
{{- end }}

{{- define "executor.serviceAccountName" -}}
{{- if .Values.executor.serviceAccount.create -}}
    {{- default (include "executor.name" .) .Values.executor.serviceAccount.name }}
{{- else -}}
    {{- default "default" .Values.executor.serviceAccount.name }}
{{- end -}}
{{- end }}

{{- define "executor.clusterRoleName" -}}
{{ include "executor.name" . }}-{{ .Release.Namespace }}
{{- end }}

{{- define "executor.clusterRoleBindingName" -}}
{{ include "executor.name" . }}-{{ .Release.Namespace }}
{{- end }}
