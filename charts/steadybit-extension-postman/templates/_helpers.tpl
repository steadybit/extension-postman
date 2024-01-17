{{/* vim: set filetype=mustache: */}}

{{/*
Expand the name of the chart.
*/}}
{{- define "postman.secret.name" -}}
{{- default "steadybit-extension-postman" .Values.postman.existingSecret -}}
{{- end -}}
