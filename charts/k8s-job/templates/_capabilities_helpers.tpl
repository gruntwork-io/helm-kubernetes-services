{{/* Allow KubeVersion to be overridden. This is mostly used for testing purposes. */}}
{{- define "gruntwork.kubeVersion" -}}
  {{- default .Capabilities.KubeVersion.Version .Values.kubeVersionOverride -}}
{{- end -}}
