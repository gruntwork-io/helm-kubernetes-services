{{/* Allow KubeVersion to be overridden. This is mostly used for testing purposes. */}}
{{- define "gruntwork.kubeVersion" -}}
  {{- default .Capabilities.KubeVersion.Version .Values.kubeVersionOverride -}}
{{- end -}}

{{/* Get Ingress API Version */}}
{{- define "gruntwork.ingress.apiVersion" -}}
  {{- if and (.Capabilities.APIVersions.Has "networking.k8s.io/v1") (semverCompare ">= 1.19-0" (include "gruntwork.kubeVersion" .)) -}}
    {{- print "networking.k8s.io/v1" -}}
  {{- else if .Capabilities.APIVersions.Has "networking.k8s.io/v1beta1" -}}
    {{- print "networking.k8s.io/v1beta1" -}}
  {{- else -}}
    {{- print "extensions/v1beta1" -}}
  {{- end -}}
{{- end -}}

{{/* Ingress API version aware ingress backend */}}
{{- define "gruntwork.ingress.backend" -}}
{{/* NOTE: The leading whitespace is significant, as it is the specific yaml indentation for injection into the ingress resource. */}}
              {{- if eq .ingressAPIVersion "networking.k8s.io/v1" }}
              service:
                name: {{ if .serviceName }}{{ .serviceName }}{{ else }}{{ .fullName }}{{ end }}
                port:
                  {{- if int .servicePort }}
                  number: {{ .servicePort }}
                  {{- else }}
                  name: {{ .servicePort }}
                  {{- end }}
              {{- else }}
              serviceName: {{ if .serviceName }}{{ .serviceName }}{{ else }}{{ .fullName }}{{ end }}
              servicePort: {{ .servicePort }}
              {{- end }}
{{- end -}}

{{/* Get PodDisruptionBudget API Version */}}
{{- define "gruntwork.pdb.apiVersion" -}}
  {{- if and (.Capabilities.APIVersions.Has "policy/v1") (semverCompare ">= 1.21-0" (include "gruntwork.kubeVersion" .)) -}}
    {{- print "policy/v1" -}}
  {{- else -}}
    {{- print "policy/v1beta1" -}}
  {{- end -}}
{{- end -}}
