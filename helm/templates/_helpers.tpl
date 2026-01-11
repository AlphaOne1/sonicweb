{{- /*
SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
SPDX-License-Identifier: MPL-2.0

Expand the name of the chart.
*/ -}}

{{- define "SonicWeb.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "SonicWeb.fullname" -}}
{{- if .Values.fullnameOverride }}
{{-     .Values.fullnameOverride | trunc 63 | trimSuffix "-" | lower }}
{{- else }}
{{-     $name := default .Chart.Name .Values.nameOverride }}
{{-     $relName := .Release.Name }}
{{-     if contains $name $relName }}
{{-         $relName | trunc 63 | trimSuffix "-" | lower }}
{{-     else }}
{{-         printf "%s-%s" $relName $name | trunc 63 | trimSuffix "-" | lower }}
{{-     end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "SonicWeb.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "SonicWeb.labels" -}}
helm.sh/chart: {{ include "SonicWeb.chart" . }}
{{ include "SonicWeb.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "SonicWeb.selectorLabels" -}}
app.kubernetes.io/name: {{ include "SonicWeb.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "SonicWeb.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "SonicWeb.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
