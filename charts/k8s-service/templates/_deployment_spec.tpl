{{- /*
Common deployment spec that is shared between the canary and main Deployment controllers. This template requires the
context:
- Values
- Release
- Chart
- isCanary (a boolean indicating if we are rendering the canary deployment or not)
You can construct this context using dict:
(dict "Values" .Values "Release" .Release "Chart" .Chart "isCanary" true)
*/ -}}
{{- define "k8s-service.deploymentSpec" -}}
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
  {{- else if eq (index . "as") "csi" -}}
    {{- $_ := set $hasInjectionTypes "hasEnvVars" true -}}
    {{- $_ := set $hasInjectionTypes "hasVolume" true -}}
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
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "k8s-service.fullname" . }}{{ if .isCanary }}-canary{{ end }}
  labels:
    # These labels are required by helm. You can read more about required labels in the chart best practices guide:
    # https://docs.helm.sh/chart_best_practices/#standard-labels
    helm.sh/chart: {{ include "k8s-service.chart" . }}
    app.kubernetes.io/name: {{ include "k8s-service.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    {{- range $key, $value := .Values.additionalDeploymentLabels }}
    {{ $key }}: "{{ $value }}"
    {{- end}}
{{- with .Values.deploymentAnnotations }}
  annotations:
{{ toYaml . | indent 4 }}
{{- end }}
spec:
{{- if .isCanary }}
  replicas: {{ .Values.canary.replicaCount | default 1 }}
{{ else }}
{{- if not .Values.horizontalPodAutoscaler.enabled }} 
  replicas: {{ .Values.replicaCount }}
{{- end }}
{{- end }}
{{- if .Values.deploymentStrategy.enabled }}
  strategy:
    type: {{ .Values.deploymentStrategy.type }}
{{- if and (eq .Values.deploymentStrategy.type "RollingUpdate") .Values.deploymentStrategy.rollingUpdate }}
    rollingUpdate:
{{ toYaml .Values.deploymentStrategy.rollingUpdate | indent 6 }}
{{- end }}
{{- end }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "k8s-service.name" . }}
      app.kubernetes.io/instance: {{ .Release.Name }}
      {{- if .isCanary }}
      gruntwork.io/deployment-type: canary
      {{- else }}
      gruntwork.io/deployment-type: main
      {{- end }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ include "k8s-service.name" . }}
        app.kubernetes.io/instance: {{ .Release.Name }}
        {{- if .isCanary }}
        gruntwork.io/deployment-type: canary
        {{- else }}
        gruntwork.io/deployment-type: main
        {{- end }}
        {{- range $key, $value := .Values.additionalPodLabels }}
        {{ $key }}: "{{ $value }}"
        {{- end }}

      {{- with .Values.podAnnotations }}
      annotations:
{{ toYaml . | indent 8 }}
      {{- end }}
    spec:
      {{- if gt (len .Values.serviceAccount.name) 0 }}
      serviceAccountName: "{{ .Values.serviceAccount.name }}"
      {{- end }}
      {{- if hasKey .Values.serviceAccount "automountServiceAccountToken" }}
      automountServiceAccountToken : {{ .Values.serviceAccount.automountServiceAccountToken }}
      {{- end }}
      {{- if .Values.podSecurityContext }}
      securityContext:
{{ toYaml .Values.podSecurityContext | indent 8 }}
      {{- end}}
      {{- if .Values.hostAliases }}
      hostAliases:
{{ toYaml .Values.hostAliases | indent 8 }}
      {{- end }}

      {{- if .Values.dnsPolicy }}
      dnsPolicy: {{ .Values.dnsPolicy }}
      {{- end }}

      containers:
        {{- if .isCanary }}
        - name: {{ .Values.applicationName }}-canary
          {{- $repo := required ".Values.canary.containerImage.repository is required" .Values.canary.containerImage.repository }}
          {{- $tag := required ".Values.canary.containerImage.tag is required" .Values.canary.containerImage.tag }}
          image: "{{ $repo }}:{{ $tag }}"
          imagePullPolicy: {{ .Values.canary.containerImage.pullPolicy | default "IfNotPresent" }}
        {{- else }}
        - name: {{ .Values.applicationName }}
          {{- $repo := required ".Values.containerImage.repository is required" .Values.containerImage.repository }}
          {{- $tag := required ".Values.containerImage.tag is required" .Values.containerImage.tag }}
          image: "{{ $repo }}:{{ $tag }}"
          imagePullPolicy: {{ .Values.containerImage.pullPolicy | default "IfNotPresent" }}
        {{- end }}
          {{- if .Values.containerCommand }}
          command:
{{ toYaml .Values.containerCommand | indent 12 }}
          {{- end }}
          {{- if .Values.containerArgs }}
          args:
{{ toYaml .Values.containerArgs | indent 12 }}
          {{- end }}
          {{- if index $hasInjectionTypes "exposePorts" }}
          ports:
            {{- /*
              NOTE: we check for a disabled flag here so that users of the helm
              chart can override the default containerPorts. Specifically, defining a new
              containerPorts in values.yaml will be merged with the default provided by the
              chart. For example, if the user provides:

                    containerPorts:
                      app:
                        port: 8080
                        protocol: TCP

              Then this is merged with the default and becomes:

                    containerPorts:
                      app:
                        port: 8080
                        protocol: TCP
                      http:
                        port: 80
                        protocol: TCP
                      https:
                        port: 443
                        protocol: TCP

              and so it becomes append as opposed to replace. To handle this,
              we allow users to explicitly disable predefined ports. So if the user wants to
              replace the ports with their own, they would provide the following values file:

                    containerPorts:
                      app:
                        port: 8080
                        protocol: TCP
                      http:
                        disabled: true
                      https:
                        disabled: true
            */ -}}
            {{- range $key, $portSpec := .Values.containerPorts }}
            {{- if not $portSpec.disabled }}
            - name: {{ $key }}
              containerPort: {{ int $portSpec.port }}
              protocol: {{ $portSpec.protocol }}
            {{- end }}
            {{- end }}
          {{- end }}

          {{- if .Values.startupProbe }}
          startupProbe:
{{ toYaml .Values.startupProbe | indent 12 }}
          {{- end }}

          {{- if .Values.livenessProbe }}
          livenessProbe:
{{ toYaml .Values.livenessProbe | indent 12 }}
          {{- end }}

          {{- if .Values.readinessProbe }}
          readinessProbe:
{{ toYaml .Values.readinessProbe | indent 12 }}
          {{- end }}
          {{- if .Values.securityContext }}
          securityContext:
{{ toYaml .Values.securityContext | indent 12 }}
          {{- end}}
          resources:
{{ toYaml .Values.containerResources | indent 12 }}

          {{- if or .Values.lifecycleHooks.enabled (gt (int .Values.shutdownDelay) 0) }}
          lifecycle:
            {{- if and .Values.lifecycleHooks.enabled .Values.lifecycleHooks.postStart }}
            postStart:
{{ toYaml .Values.lifecycleHooks.postStart | indent 14 }}
            {{- end }}

            {{- if and .Values.lifecycleHooks.enabled .Values.lifecycleHooks.preStop }}
            preStop:
{{ toYaml .Values.lifecycleHooks.preStop | indent 14 }}
            {{- else if gt (int .Values.shutdownDelay) 0 }}
            # Include a preStop hook with a shutdown delay for eventual consistency reasons.
            # See https://blog.gruntwork.io/delaying-shutdown-to-wait-for-pod-deletion-propagation-445f779a8304
            preStop:
              exec:
                command:
                  - sleep
                  - "{{ int .Values.shutdownDelay }}"
            {{- end }}

          {{- end }}

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
            {{- if or (eq $value.as "environment") (eq $value.as "csi") }}
            {{- range $secretKey, $keyEnvVarConfig := $value.items }}
            - name: {{ required "envVarName is required on secrets items when using environment or csi" $keyEnvVarConfig.envVarName | quote }}
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


          {{- /* START VOLUME MOUNT LOGIC */ -}}
          {{- if index $hasInjectionTypes "hasVolume" }}
          volumeMounts:
          {{- end }}
          {{- range $name, $value := .Values.configMaps }}
            {{- if eq $value.as "volume" }}
            - name: {{ $name }}-volume
              mountPath: {{ quote $value.mountPath }}
              {{- if $value.subPath }}
              subPath: {{ quote $value.subPath }}
              {{- end }}
            {{- end }}
          {{- end }}
          {{- range $name, $value := .Values.secrets }}
            {{- if or (eq $value.as "volume") (eq $value.as "csi") }}
            - name: {{ $name }}-volume
              mountPath: {{ quote $value.mountPath }}
              {{- if $value.subPath }}
              subPath: {{ quote $value.subPath }}
              {{- end }}
              readOnly: {{ $value.readOnly }}
            {{- end }}
          {{- end }}
          {{- range $name, $value := .Values.persistentVolumes }}
            - name: {{ $name }}
              mountPath: {{ quote $value.mountPath }}
          {{- end }}
          {{- range $name, $value := .Values.scratchPaths }}
            - name: {{ $name }}
              mountPath: {{ quote $value }}
          {{- end }}
          {{- range $name, $value := .Values.emptyDirs }}
            - name: {{ $name }}
              mountPath: {{ quote $value }}
          {{- end }}
          {{- /* END VOLUME MOUNT LOGIC */ -}}

        {{- range $key, $value := .Values.sideCarContainers }}
        - name: {{ $key }}
{{ toYaml $value | indent 10 }}
        {{- end }}


    {{- if gt (len .Values.initContainers) 0 }}
      initContainers:
        {{- range $key, $value := .Values.initContainers }}
        - name: {{ $key }}
{{ toYaml $value | indent 10 }}
        {{- end }}
    {{- end }}

    {{- /* START IMAGE PULL SECRETS LOGIC */ -}}
    {{- if gt (len .Values.imagePullSecrets) 0 }}
      imagePullSecrets:
        {{- range $secretName := .Values.imagePullSecrets }}
        - name: {{ $secretName }}
        {{- end }}
    {{- end }}
    {{- /* END IMAGE PULL SECRETS LOGIC */ -}}

    {{- /* START TERMINATION GRACE PERIOD LOGIC */ -}}
    {{- if .Values.terminationGracePeriodSeconds }}
      terminationGracePeriodSeconds: {{ .Values.terminationGracePeriodSeconds }}
    {{- end}}
    {{- /* END TERMINATION GRACE PERIOD LOGIC */ -}}

    {{- /* START VOLUME LOGIC */ -}}
    {{- if index $hasInjectionTypes "hasVolume" }}
      volumes:
    {{- end }}
    {{- range $name, $value := .Values.configMaps }}
      {{- if eq $value.as "volume" }}
        - name: {{ $name }}-volume
          configMap:
            name: {{ $name }}
            {{- if $value.items }}
            items:
              {{- range $configKey, $keyMountConfig := $value.items }}
              - key: {{ $configKey }}
                path: {{ required "filePath is required for configMap items" $keyMountConfig.filePath | quote }}
                {{- if $keyMountConfig.fileMode }}
                mode: {{ include "k8s-service.fileModeOctalToDecimal" $keyMountConfig.fileMode }}
                {{- end }}
              {{- end }}
            {{- end }}
      {{- end }}
    {{- end }}
    {{- range $name, $value := .Values.secrets }}
      {{- if eq $value.as "volume" }}
        - name: {{ $name }}-volume
          secret:
            secretName: {{ $name }}
            {{- if $value.items }}
            items:
              {{- range $secretKey, $keyMountConfig := $value.items }}
              - key: {{ $secretKey }}
                path: {{ required "filePath is required for secrets items" $keyMountConfig.filePath | quote }}
                {{- if $keyMountConfig.fileMode }}
                mode: {{ include "k8s-service.fileModeOctalToDecimal" $keyMountConfig.fileMode }}
                {{- end }}
              {{- end }}
            {{- end }}
      {{- end }}
      {{- if eq $value.as "csi" }}
        - name: {{ $name }}-volume
          csi: 
            readOnly: {{ $value.readOnly }}
            driver:  {{ $value.csi.driver }}
            volumeAttributes:
              secretProviderClass: {{ $value.csi.secretProviderClass }}

      {{- end }}    
    {{- end }}
    {{- range $name, $value := .Values.persistentVolumes }}
        - name: {{ $name }}
          persistentVolumeClaim:
            claimName: {{ $value.claimName }}
    {{- end }}
    {{- range $name, $value := .Values.scratchPaths }}
        - name: {{ $name }}
          emptyDir:
            medium: "Memory"
    {{- end }}
    {{- range $name, $value := .Values.emptyDirs }}
        - name: {{ $name }}
          emptyDir: {}
    {{- end }}
    {{- /* END VOLUME LOGIC */ -}}

    {{- with .Values.nodeSelector }}
      nodeSelector:
{{ toYaml . | indent 8 }}
    {{- end }}

    {{- with .Values.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
    {{- end }}

    {{- with .Values.topologySpreadConstraints }}
      topologySpreadConstraints:
{{ toYaml . | indent 8 }}
    {{- end }}

    {{- with .Values.priorityClassName }}
      priorityClassName:
{{ toYaml . | indent 8 }}
    {{- end }}

    {{- with .Values.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
    {{- end }}
{{- end -}}
