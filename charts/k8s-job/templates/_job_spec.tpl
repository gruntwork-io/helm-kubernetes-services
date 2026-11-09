{{- /*
Common job spec. This template requires the
context:
- Values
- Release
- Chart
You can construct this context using dict:
(dict "Values" .Values "Release" .Release "Chart" .Chart "isCanary" true)
*/ -}}
{{- define "k8s-job.jobSpec" -}}
{{- /*
We must decide whether or not there are volumes to inject. The logic to decide whether or not to inject is based on
whether or not there are configMaps OR secrets that are specified as volume mounts (`as: volume` attributes). We do this
by using a map to track whether or not we have seen a volume type. We have to use a map because we can't update a
variable in helm chart templates.

Similarly, we need to decide whether or not there are environment variables to add

We need this because certain sections are omitted if there are no volumes or environment variables to add.
*/ -}}

{{/* Go Templates do not support variable updating, so we simulate it using dictionaries */}}
{{- $hasInjectionTypes := dict "hasVolume" false "hasEnvVars" false "exposePorts" false -}}
{{- if .Values.envVars -}}
  {{- $_ := set $hasInjectionTypes "hasEnvVars" true -}}
{{- end -}}
{{- if .Values.additionalContainerEnv -}}
  {{- $_ := set $hasInjectionTypes "hasEnvVars" true -}}
{{- end -}}
{{- $allContainerPorts := values .Values.containerPorts -}}
{{- range $allContainerPorts -}}
  {{/* We are exposing ports if there is at least one key in containerPorts that is not disabled (disabled = false or
       omitted)
  */}}
  {{- if or (not (hasKey . "disabled")) (not .disabled) -}}
    {{- $_ := set $hasInjectionTypes "exposePorts" true -}}
  {{- end -}}
{{- end -}}
{{- $allSecrets := values .Values.secrets -}}
{{- range $allSecrets -}}
  {{- if eq (index . "as") "volume" -}}
    {{- $_ := set $hasInjectionTypes "hasVolume" true -}}
  {{- else if eq (index . "as") "environment" -}}
    {{- $_ := set $hasInjectionTypes "hasEnvVars" true -}}
  {{- else if eq (index . "as") "envFrom" }}
    {{- $_ := set $hasInjectionTypes "hasEnvFrom" true -}}
  {{- else if eq (index . "as") "none" -}}
    {{- /* noop */ -}}
  {{- else -}}
    {{- fail printf "secrets config has unknown type: %s" (index . "as") -}}
  {{- end -}}
{{- end -}}
{{- $allConfigMaps := values .Values.configMaps -}}
{{- range $allConfigMaps -}}
  {{- if eq (index . "as") "volume" -}}
    {{- $_ := set $hasInjectionTypes "hasVolume" true -}}
  {{- else if eq (index . "as") "environment" -}}
    {{- $_ := set $hasInjectionTypes "hasEnvVars" true -}}
  {{- else if eq (index . "as") "envFrom" }}
    {{- $_ := set $hasInjectionTypes "hasEnvFrom" true -}}
  {{- else if eq (index . "as") "none" -}}
    {{- /* noop */ -}}
  {{- else -}}
    {{- fail printf "configMaps config has unknown type: %s" (index . "as") -}}
  {{- end -}}
{{- end -}}
{{- if gt (len .Values.persistentVolumes) 0 -}}
  {{- $_ := set $hasInjectionTypes "hasVolume" true -}}
{{- end -}}
{{- if gt (len .Values.scratchPaths) 0 -}}
  {{- $_ := set $hasInjectionTypes "hasVolume" true -}}
{{- end -}}
{{- if gt (len .Values.emptyDirs) 0 -}}
  {{- $_ := set $hasInjectionTypes "hasVolume" true -}}
{{- end -}}
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ include "k8s-job.fullname" . }}
  labels:
    # These labels are required by helm. You can read more about required labels in the chart best practices guide:
    # https://docs.helm.sh/chart_best_practices/#standard-labels
    helm.sh/chart: {{ include "k8s-job.chart" . }}
    app.kubernetes.io/name: {{ include "k8s-job.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    {{- range $key, $value := .Values.additionalJobLabels }}
    {{ $key }}: {{ $value }}
    {{- end}}
{{- with .Values.jobAnnotations }}
  annotations:
{{ toYaml . | indent 4 }}
{{- end }}
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "k8s-job.name" . }}
      app.kubernetes.io/instance: {{ .Release.Name }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ include "k8s-job.name" . }}
        app.kubernetes.io/instance: {{ .Release.Name }}
        gruntwork.io/deployment-type: main
        {{- end }}
        {{- range $key, $value := .Values.additionalPodLabels }}
        {{ $key }}: {{ $value }}
        {{- end }}

      {{- with .Values.podAnnotations }}
      annotations:
{{ toYaml . | indent 8 }}
      {{- end }}
    spec:
      {{- if .Values.podSecurityContext }}
      securityContext:
{{ toYaml .Values.podSecurityContext | indent 8 }}
      {{- end}}
      
      restartPolicy: {{ toYaml .Values.restartPolicy | indent 12 }}
      containers:
        - name: {{ .Values.applicationName }}
          {{- $repo := required ".Values.containerImage.repository is required" .Values.containerImage.repository }}
          {{- $tag := required ".Values.containerImage.tag is required" .Values.containerImage.tag }}
          image: "{{ $repo }}:{{ $tag }}"
          imagePullPolicy: {{ .Values.containerImage.pullPolicy | default "IfNotPresent" }}
          {{- if .Values.containerCommand }}
          command:
{{ toYaml .Values.containerCommand | indent 12 }}
          {{- if .Values.containerArgs }}
          args:
{{ toYaml .Values.containerArgs | indent 12 }}
          {{- end }}
          securityContext:
{{ toYaml .Values.securityContext | indent 12 }}
          {{- end}}
          resources:
{{ toYaml .Values.containerResources | indent 12 }}

          {{- /* START ENV VAR LOGIC */ -}}
          {{- if index $hasInjectionTypes "hasEnvVars" }}
          env:
          {{- end }}
          {{- range $key, $value := .Values.envVars }}
            - name: {{ $key }}
              value: {{ quote $value }}
          {{- end }}
          {{- if .Values.additionalContainerEnv }}
{{ toYaml .Values.additionalContainerEnv | indent 12 }}
          {{- end }}
          {{- range $name, $value := .Values.configMaps }}
            {{- if eq $value.as "environment" }}
            {{- range $configKey, $keyEnvVarConfig := $value.items }}
            - name: {{ required "envVarName is required on configMaps items when using environment" $keyEnvVarConfig.envVarName | quote }}
              valueFrom:
                configMapKeyRef:
                  name: {{ $name }}
                  key: {{ $configKey }}
            {{- end }}
            {{- end }}
          {{- end }}
          {{- range $name, $value := .Values.secrets }}
            {{- if eq $value.as "environment" }}
            {{- range $secretKey, $keyEnvVarConfig := $value.items }}
            - name: {{ required "envVarName is required on secrets items when using environment" $keyEnvVarConfig.envVarName | quote }}
              valueFrom:
                secretKeyRef:
                  name: {{ $name }}
                  key: {{ $secretKey }}
            {{- end }}
            {{- end }}
          {{- end }}
          {{- if index $hasInjectionTypes "hasEnvFrom" }}
          envFrom:
          {{- range $name, $value := .Values.configMaps }}
            {{- if eq $value.as "envFrom" }}
            - configMapRef:
                name: {{ $name }}
            {{- end }}
          {{- end }}
          {{- range $name, $value := .Values.secrets }}
            {{- if eq $value.as "envFrom" }}
            - secretRef:
                name: {{ $name }}
            {{- end }}
          {{- end }}
          {{- end }}
          {{- /* END ENV VAR LOGIC */ -}}

    {{- /* START IMAGE PULL SECRETS LOGIC */ -}}
    {{- if gt (len .Values.imagePullSecrets) 0 }}
      imagePullSecrets:
        {{- range $secretName := .Values.imagePullSecrets }}
        - name: {{ $secretName }}
        {{- end }}
    {{- end }}
    {{- /* END IMAGE PULL SECRETS LOGIC */ -}}


{{- end -}}
