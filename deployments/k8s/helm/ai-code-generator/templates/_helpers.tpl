{{/*
Expand the name of the chart.
*/}}
{{- define "ai-code-generator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ai-code-generator.fullname" -}}
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
{{- define "ai-code-generator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ai-code-generator.labels" -}}
helm.sh/chart: {{ include "ai-code-generator.chart" . }}
{{ include "ai-code-generator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ai-code-generator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ai-code-generator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
PostgreSQL host
*/}}
{{- define "ai-code-generator.postgresql.host" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql" (include "ai-code-generator.fullname" .) }}
{{- else }}
{{- required "postgresql.host is required when postgresql.enabled is false" .Values.postgresql.host }}
{{- end }}
{{- end }}

{{/*
PostgreSQL port
*/}}
{{- define "ai-code-generator.postgresql.port" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "5432" }}
{{- else }}
{{- .Values.postgresql.port | default "5432" }}
{{- end }}
{{- end }}

{{/*
PostgreSQL database
*/}}
{{- define "ai-code-generator.postgresql.database" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.database }}
{{- else }}
{{- .Values.postgresql.database | default "ai_ui_generator" }}
{{- end }}
{{- end }}

{{/*
PostgreSQL username
*/}}
{{- define "ai-code-generator.postgresql.username" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.username | default "postgres" }}
{{- else }}
{{- .Values.postgresql.username | default "postgres" }}
{{- end }}
{{- end }}

{{/*
PostgreSQL secret name
*/}}
{{- define "ai-code-generator.postgresql.secretName" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql" (include "ai-code-generator.fullname" .) }}
{{- else }}
{{- .Values.postgresql.existingSecret | default "ai-code-generator-postgresql" }}
{{- end }}
{{- end }}

{{/*
Redis host
*/}}
{{- define "ai-code-generator.redis.host" -}}
{{- if .Values.redis.enabled }}
{{- printf "%s-redis" (include "ai-code-generator.fullname" .) }}
{{- else }}
{{- required "redis.host is required when redis.enabled is false" .Values.redis.host }}
{{- end }}
{{- end }}

{{/*
Redis port
*/}}
{{- define "ai-code-generator.redis.port" -}}
{{- if .Values.redis.enabled }}
{{- printf "6379" }}
{{- else }}
{{- .Values.redis.port | default "6379" }}
{{- end }}
{{- end }}

{{/*
Redis secret name
*/}}
{{- define "ai-code-generator.redis.secretName" -}}
{{- if .Values.redis.enabled }}
{{- printf "%s-redis" (include "ai-code-generator.fullname" .) }}
{{- else }}
{{- .Values.redis.existingSecret | default "ai-code-generator-redis" }}
{{- end }}
{{- end }}