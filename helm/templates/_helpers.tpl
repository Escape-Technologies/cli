{{- define "fullname" -}}
{{ .Release.Name }}-{{ .Chart.Name }}
{{- end -}}
