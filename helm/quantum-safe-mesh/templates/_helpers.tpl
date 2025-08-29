{{/*
Expand the name of the chart.
*/}}
{{- define "quantum-safe-mesh.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "quantum-safe-mesh.fullname" -}}
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
{{- define "quantum-safe-mesh.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "quantum-safe-mesh.labels" -}}
helm.sh/chart: {{ include "quantum-safe-mesh.chart" . }}
{{ include "quantum-safe-mesh.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "quantum-safe-mesh.selectorLabels" -}}
app.kubernetes.io/name: {{ include "quantum-safe-mesh.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "quantum-safe-mesh.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "quantum-safe-mesh.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Get the full image name
*/}}
{{- define "quantum-safe-mesh.image" -}}
{{- $registry := .registry -}}
{{- if .Values.global.imageRegistry -}}
{{- $registry = .Values.global.imageRegistry -}}
{{- end -}}
{{- if $registry -}}
{{- printf "%s/%s:%s" $registry .repository .tag -}}
{{- else -}}
{{- printf "%s:%s" .repository .tag -}}
{{- end -}}
{{- end }}

{{/*
Create auth service URL
*/}}
{{- define "quantum-safe-mesh.authServiceUrl" -}}
{{- printf "http://%s.%s.svc.cluster.local:%d" .Values.authService.name .Values.namespace.name (.Values.authService.port | int) }}
{{- end }}

{{/*
Create backend service URL
*/}}
{{- define "quantum-safe-mesh.backendServiceUrl" -}}
{{- printf "http://%s.%s.svc.cluster.local:%d" .Values.backendService.name .Values.namespace.name (.Values.backendService.port | int) }}
{{- end }}