{{/*
Expand the name of the chart.
*/}}
{{- define "dataflow.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "dataflow.fullname" -}}
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
ecron-management fullname
*/}}
{{- define "dataflow.ecronManagement.fullname" -}}
ecron-management
{{- end }}

{{/*
flow-automation fullname
*/}}
{{- define "dataflow.flowAutomation.fullname" -}}
flow-automation
{{- end }}

{{/*
flow-stream-data-pipeline fullname
*/}}
{{- define "dataflow.streamDataPipeline.fullname" -}}
flow-stream-data-pipeline
{{- end }}

{{/*
Common labels
*/}}
{{- define "dataflow.labels" -}}
helm.sh/chart: {{ include "dataflow.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
{{- end }}

{{/*
flow-automation labels
*/}}
{{- define "flow-automation.labels" -}}
helm.sh/chart: {{ include "dataflow.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Helper to merge component and global dependency services
Usage: {{ include "dataflow.utils.mergeDep" (dict "component" .Values.ecronAnalysis.depServices.rds "global" .Values.depServices.rds) }}
*/}}
{{- define "dataflow.utils.mergeDep" -}}
{{- $component := .component | default dict -}}
{{- $global := .global | default dict -}}
{{- merge $component $global | toYaml -}}
{{- end }}
