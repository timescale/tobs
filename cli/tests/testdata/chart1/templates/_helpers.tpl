{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "tobs.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "tobs.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "tobs.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "tobs.labels" -}}
helm.sh/chart: {{ include "tobs.chart" . }}
{{ include "tobs.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "tobs.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tobs.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "tobs.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "tobs.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
CLI release name and namespace
*/}}
{{- define "tobs.cliOptions" -}}
{{- if ne .Release.Name "tobs" }} -n {{ .Release.Name }}
{{- end -}}
{{- if ne .Release.Namespace "default" }} --namespace {{ .Release.Namespace }}
{{- end -}}
{{- end -}}

{{/*
Extract the username from db uri
*/}}
{{- define "tobs.dburi.user" -}}
  {{- $values := urlParse .Values.timescaledbExternal.db_uri }}
  {{- $userInfo := get $values "userinfo" }}
  {{- $userDetails :=  split ":" $userInfo }}
  {{- $user := $userDetails._0 }}
  {{- printf $user -}}
{{- end -}}

{{/*
Extract the password from db uri
*/}}
{{- define "tobs.dburi.password" -}}
  {{- $values := urlParse .Values.timescaledbExternal.db_uri }}
  {{- $userInfo := get $values "userinfo" }}
  {{- $userDetails :=  split ":" $userInfo }}
  {{- $pwd := $userDetails._1 }}
  {{- printf $pwd -}}
{{- end -}}

{{/*
Extract the host from db uri
*/}}
{{- define "tobs.dburi.host" -}}
  {{- $values := urlParse .Values.timescaledbExternal.db_uri }}
  {{- $hostURL := get $values "host" }}
  {{- printf $hostURL -}}
{{- end -}}

{{/*
Extract the dbname from db uri
*/}}
{{- define "tobs.dburi.dbname" -}}
  {{- $values := urlParse .Values.timescaledbExternal.db_uri }}
  {{- $dbDetails := get $values "path" }}
  {{- $dbName := trimPrefix "/" $dbDetails }}
  {{- printf $dbName -}}
{{- end -}}

{{/*
Extract the sslmode from db uri
*/}}
{{- define "tobs.dburi.sslmode" -}}
  {{- $values := urlParse .Values.timescaledbExternal.db_uri }}
  {{- $queryInfo := get $values "query" }}
  {{- $sslInfo := regexFind "ssl[mM]ode=[^&]+" $queryInfo}}
  {{- $sslDetails := split "=" $sslInfo }}
  {{- $sslMode := $sslDetails._1 }}
  {{- printf $sslMode -}}
{{- end -}}

{{/*
Extract the port from db uri
*/}}
{{- define "tobs.dburi.port" -}}
  {{- $values := urlParse .Values.timescaledbExternal.db_uri }}
  {{- $hostURL := get $values "host" }}
  {{- $hostDetails := split ":" $hostURL}}
  {{- $port := $hostDetails._1 | quote }}
  {{- printf $port -}}
{{- end -}}

{{/*
Extract the port from db uri
*/}}
{{- define "tobs.dburi.hostwithoutport" -}}
  {{- $values := urlParse .Values.timescaledbExternal.db_uri }}
  {{- $hostURL := get $values "host" }}
  {{- $hostDetails := split ":" $hostURL}}
  {{- $host := $hostDetails._0 | quote }}
  {{- printf $host -}}
{{- end -}}

{{/*
Allow the release namespace to be overridden
*/}}
{{- define "tobs.namespace" -}}
  {{- if .Values.namespaceOverride -}}
    {{- .Values.namespaceOverride -}}
  {{- else -}}
    {{- .Release.Namespace -}}
  {{- end -}}
{{- end -}}