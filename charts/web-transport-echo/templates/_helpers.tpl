{{- define "web-transport-echo-server.name" -}}
{{- if .Values.nameOverride }}
{{- .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end }}

{{- define "web-transport-echo-server.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s" $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end }}

{{- define "web-transport-echo-server.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" -}}
{{- end }}

{{- define "web-transport-echo-server.labels" -}}
helm.sh/chart: {{ include "web-transport-echo-server.chart" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/name: {{ include "web-transport-echo-server.name" . }}
{{- end }}

{{- define "web-transport-echo-server.selectorLabels" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/name: {{ include "web-transport-echo-server.name" . }}
{{- end }}

{{- define "web-transport-echo-server.podLabels" -}}
{{- include "web-transport-echo-server.selectorLabels" . | nindent 0 }}
{{- with .Values.podLabels }}
{{- toYaml . | nindent 0 }}
{{- end }}
{{- end }}

{{- define "web-transport-echo-server.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "web-transport-echo-server.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end }}
