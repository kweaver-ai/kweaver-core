{{/*
Expand the name of the chart.
*/}}
{{- define "sandbox.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "sandbox.fullname" -}}
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
{{- define "sandbox.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Control Plane name
*/}}
{{- define "sandbox.controlPlaneName" -}}
{{- printf "%s-control-plane" (include "sandbox.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Web Console name
*/}}
{{- define "sandbox.webName" -}}
{{- printf "%s-web" (include "sandbox.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
ServiceAccount name
*/}}
{{- define "sandbox.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "sandbox.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
RBAC name
*/}}
{{- define "sandbox.rbacName" -}}
{{- if .Values.rbac.create }}
{{- default (include "sandbox.fullname" .) .Values.rbac.name }}
{{- else }}
{{- default "default" .Values.rbac.name }}
{{- end }}
{{- end }}

{{/*
Labels
*/}}
{{- define "sandbox.labels" -}}
helm.sh/chart: {{ include "sandbox.chart" . }}
{{ include "sandbox.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "sandbox.selectorLabels" -}}
app.kubernetes.io/name: {{ include "sandbox.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Get the ingress class name from depServices
*/}}
{{- define "sandbox.ingressClass" -}}
{{- if and .Values.depServices (index .Values.depServices "class-443") (index .Values.depServices "class-443").ingressClass -}}
{{- (index .Values.depServices "class-443").ingressClass -}}
{{- else -}}
{{- "nginx" -}}
{{- end -}}
{{- end }}
