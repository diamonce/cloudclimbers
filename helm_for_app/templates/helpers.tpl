{{- define "hns-blog.labels" -}}
app.kubernetes.io/name: {{ include "hns-blog.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "hns-blog.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hns-blog.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "hns-blog.name" -}}
{{ .Chart.Name }}
{{- end }}

{{- define "hns-blog.fullname" -}}
{{ include "hns-blog.name" . }}-{{ .Release.Name }}
{{- end }}
