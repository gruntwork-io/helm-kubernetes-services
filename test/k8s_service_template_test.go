// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

// Test that setting ingress.enabled = false will cause the helm template to not render the Ingress resource
func TestK8SServiceIngressEnabledFalseDoesNotCreateIngress(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   map[string]string{"ingress.enabled": "false"},
	}
	_, err = helm.RenderTemplateE(t, options, helmChartPath, "ingress", []string{"templates/ingress.yaml"})
	require.Error(t, err)
}

// Test that setting service.enabled = false will cause the helm template to not render the Service resource
func TestK8SServiceServiceEnabledFalseDoesNotCreateService(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   map[string]string{"service.enabled": "false"},
	}
	_, err = helm.RenderTemplateE(t, options, helmChartPath, "service", []string{"templates/service.yaml"})
	require.Error(t, err)
}

// Test each of the required values. Here, we take advantage of the fact that linter_values.yaml is supposed to define
// all the required values, so we check the template rendering by nulling out each field.
func TestK8SServiceRequiredValuesAreRequired(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	eachRequired := []string{
		"containerImage.repository",
		"containerImage.tag",
		"applicationName",
	}
	for _, requiredVal := range eachRequired {
		// Capture the range value and force it into this scope. Otherwise, it is defined outside this block so it can
		// change when the subtests parallelize and switch contexts.
		requiredVal := requiredVal
		t.Run(requiredVal, func(t *testing.T) {
			t.Parallel()

			// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
			// We then use SetValues to null out the value.
			options := &helm.Options{
				ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
				SetValues:   map[string]string{requiredVal: "null"},
			}
			_, err := helm.RenderTemplateE(t, options, helmChartPath, strings.ToLower(t.Name()), []string{})
			assert.Error(t, err)
		})
	}
}

// Test each of the optional values defined in linter_values.yaml. Here, we take advantage of the fact that
// linter_values.yaml is supposed to define all the required values, so we check the template rendering by nulling out
// each field.
func TestK8SServiceOptionalValuesAreOptional(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	eachOptional := []string{
		"containerImage.pullPolicy",
	}
	for _, optionalVal := range eachOptional {
		// Capture the range value and force it into this scope. Otherwise, it is defined outside this block so it can
		// change when the subtests parallelize and switch contexts.
		optionalVal := optionalVal
		t.Run(optionalVal, func(t *testing.T) {
			t.Parallel()

			// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
			// We then use SetValues to null out the value.
			options := &helm.Options{
				ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
				SetValues:   map[string]string{optionalVal: "null"},
			}
			// Make sure it renders without error
			helm.RenderTemplate(t, options, helmChartPath, "all", []string{})
		})
	}
}

// Test that deploymentAnnotations render correctly to annotate the Deployment resource
func TestK8SServiceDeploymentAnnotationsRenderCorrectly(t *testing.T) {
	t.Parallel()

	uniqueID := random.UniqueId()
	deployment := renderK8SServiceDeploymentWithSetValues(t, map[string]string{"deploymentAnnotations.unique-id": uniqueID})

	assert.Equal(t, len(deployment.Annotations), 1)
	assert.Equal(t, deployment.Annotations["unique-id"], uniqueID)
}

func TestK8SServiceSecurityContextAnnotationRenderCorrectly(t *testing.T) {
	t.Parallel()
	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"securityContext.privileged": "true",
			"securityContext.runAsUser":  "1000",
		},
	)
	renderedContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedContainers), 1)
	testContainer := renderedContainers[0]
	assert.NotNil(t, testContainer.SecurityContext)
	assert.True(t, *testContainer.SecurityContext.Privileged)
	assert.Equal(t, *testContainer.SecurityContext.RunAsUser, int64(1000))
}

// Test that podAnnotations render correctly to annotate the Pod Template Spec on the Deployment resource
func TestK8SServicePodAnnotationsRenderCorrectly(t *testing.T) {
	t.Parallel()

	uniqueID := random.UniqueId()
	deployment := renderK8SServiceDeploymentWithSetValues(t, map[string]string{"podAnnotations.unique-id": uniqueID})

	renderedPodAnnotations := deployment.Spec.Template.Annotations
	assert.Equal(t, len(renderedPodAnnotations), 1)
	assert.Equal(t, renderedPodAnnotations["unique-id"], uniqueID)
}

// Test that containerPorts render correctly to convert the map to a list
func TestK8SServiceContainerPortsSetPortsCorrectly(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			// disable the default ports
			"containerPorts.http.disabled":  "true",
			"containerPorts.https.disabled": "true",
			// ... and specify a new port
			"containerPorts.app.port":     "9876",
			"containerPorts.app.protocol": "TCP",
		},
	)

	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]

	require.Equal(t, len(appContainer.Ports), 1)
	setPort := appContainer.Ports[0]

	assert.Equal(t, setPort.Name, "app")
	assert.Equal(t, setPort.ContainerPort, int32(9876))
	assert.Equal(t, setPort.Protocol, corev1.Protocol("TCP"))
}

// Test that default imagePullSecrets do not render any
func TestK8SServiceNoImagePullSecrets(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{},
	)

	renderedImagePullSecrets := deployment.Spec.Template.Spec.ImagePullSecrets
	require.Equal(t, len(renderedImagePullSecrets), 0)
}

// Test that multiple imagePullSecrets renders each one correctly
func TestK8SServiceMultipleImagePullSecrets(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"imagePullSecrets[0]": "docker-private-registry-key",
			"imagePullSecrets[1]": "gcr-registry-key",
		},
	)

	renderedImagePullSecrets := deployment.Spec.Template.Spec.ImagePullSecrets
	require.Equal(t, len(renderedImagePullSecrets), 2)
	assert.Equal(t, renderedImagePullSecrets[0].Name, "docker-private-registry-key")
	assert.Equal(t, renderedImagePullSecrets[1].Name, "gcr-registry-key")
}

// Test that setting shutdownDelay to 0 will disable the preStop hook
func TestK8SServiceShutdownDelayZeroDisablesPreStopHook(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(t, map[string]string{"shutdownDelay": "0"})

	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	assert.Nil(t, appContainer.Lifecycle)
}

// Test that setting shutdownDelay to something greater than 0 will include a preStop hook
func TestK8SServiceNonZeroShutdownDelayIncludesPreStopHook(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(t, map[string]string{"shutdownDelay": "5"})

	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	require.NotNil(t, appContainer.Lifecycle)
	require.NotNil(t, appContainer.Lifecycle.PreStop)
	require.NotNil(t, appContainer.Lifecycle.PreStop.Exec)
	require.Equal(t, appContainer.Lifecycle.PreStop.Exec.Command, []string{"sleep", "5"})
}

// Test that setting additionalPaths on ingress add paths after service path
func TestK8SServiceIngressAdditionalPathsAfterMainServicePath(t *testing.T) {
	t.Parallel()

	ingress := renderK8SServiceIngressWithSetValues(
		t,
		map[string]string{
			"ingress.enabled":                        "true",
			"ingress.path":                           "/app",
			"ingress.servicePort":                    "app",
			"ingress.additionalPaths[0].path":        "/black-hole",
			"ingress.additionalPaths[0].serviceName": "black-hole",
			"ingress.additionalPaths[0].servicePort": "80",
		},
	)
	pathRules := ingress.Spec.Rules[0].HTTP.Paths
	assert.Equal(t, len(pathRules), 2)

	// The first path should be the main service path
	firstPath := pathRules[0]
	assert.Equal(t, firstPath.Path, "/app")
	assert.Equal(t, strings.ToLower(firstPath.Backend.ServiceName), "ingress-linter")
	assert.Equal(t, firstPath.Backend.ServicePort.StrVal, "app")

	// The second path should be the black hole
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/black-hole")
	assert.Equal(t, secondPath.Backend.ServiceName, "black-hole")
	assert.Equal(t, secondPath.Backend.ServicePort.IntVal, int32(80))
}

// Test that setting additionalPaths with multiple entries on ingress add paths after service path in order
func TestK8SServiceIngressAdditionalPathsMultipleAfterMainServicePath(t *testing.T) {
	t.Parallel()

	ingress := renderK8SServiceIngressWithSetValues(
		t,
		map[string]string{
			"ingress.enabled":                        "true",
			"ingress.path":                           "/app",
			"ingress.servicePort":                    "app",
			"ingress.additionalPaths[0].path":        "/sun",
			"ingress.additionalPaths[0].serviceName": "sun",
			"ingress.additionalPaths[0].servicePort": "3000",
			"ingress.additionalPaths[1].path":        "/black-hole",
			"ingress.additionalPaths[1].serviceName": "black-hole",
			"ingress.additionalPaths[1].servicePort": "80",
		},
	)
	pathRules := ingress.Spec.Rules[0].HTTP.Paths
	assert.Equal(t, len(pathRules), 3)

	// The first path should be the main service path
	firstPath := pathRules[0]
	assert.Equal(t, firstPath.Path, "/app")
	assert.Equal(t, strings.ToLower(firstPath.Backend.ServiceName), "ingress-linter")
	assert.Equal(t, firstPath.Backend.ServicePort.StrVal, "app")

	// The second path should be the sun
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/sun")
	assert.Equal(t, secondPath.Backend.ServiceName, "sun")
	assert.Equal(t, secondPath.Backend.ServicePort.IntVal, int32(3000))

	// The third path should be the black hole
	thirdPath := pathRules[2]
	assert.Equal(t, thirdPath.Path, "/black-hole")
	assert.Equal(t, thirdPath.Backend.ServiceName, "black-hole")
	assert.Equal(t, thirdPath.Backend.ServicePort.IntVal, int32(80))
}

// Test that omitting a serviceName on additionalPaths reuses the application service name
func TestK8SServiceIngressAdditionalPathsNoServiceName(t *testing.T) {
	t.Parallel()

	ingress := renderK8SServiceIngressWithSetValues(
		t,
		map[string]string{
			"ingress.enabled":                        "true",
			"ingress.path":                           "/app",
			"ingress.servicePort":                    "app",
			"ingress.additionalPaths[0].path":        "/black-hole",
			"ingress.additionalPaths[0].servicePort": "3000",
		},
	)
	pathRules := ingress.Spec.Rules[0].HTTP.Paths
	assert.Equal(t, len(pathRules), 2)

	// The first path should be the main service path
	firstPath := pathRules[0]
	assert.Equal(t, firstPath.Path, "/app")
	assert.Equal(t, strings.ToLower(firstPath.Backend.ServiceName), "ingress-linter")
	assert.Equal(t, firstPath.Backend.ServicePort.StrVal, "app")

	// The second path should be the black hole
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/black-hole")
	assert.Equal(t, strings.ToLower(secondPath.Backend.ServiceName), "ingress-linter")
	assert.Equal(t, secondPath.Backend.ServicePort.IntVal, int32(3000))
}

// Test that setting additionalPathsHigherPriority on ingress add paths before service path
func TestK8SServiceIngressAdditionalPathsHigherPriorityBeforeMainServicePath(t *testing.T) {
	t.Parallel()

	ingress := renderK8SServiceIngressWithSetValues(
		t,
		map[string]string{
			"ingress.enabled":     "true",
			"ingress.path":        "/app",
			"ingress.servicePort": "app",
			"ingress.additionalPathsHigherPriority[0].path":        "/black-hole",
			"ingress.additionalPathsHigherPriority[0].serviceName": "black-hole",
			"ingress.additionalPathsHigherPriority[0].servicePort": "80",
		},
	)
	pathRules := ingress.Spec.Rules[0].HTTP.Paths
	assert.Equal(t, len(pathRules), 2)

	// The first path should be the black hole
	firstPath := pathRules[0]
	assert.Equal(t, firstPath.Path, "/black-hole")
	assert.Equal(t, firstPath.Backend.ServiceName, "black-hole")
	assert.Equal(t, firstPath.Backend.ServicePort.IntVal, int32(80))

	// The second path should be the main service path
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/app")
	assert.Equal(t, strings.ToLower(secondPath.Backend.ServiceName), "ingress-linter")
	assert.Equal(t, secondPath.Backend.ServicePort.StrVal, "app")
}

// Test that setting additionalPathsHigherPriority with multiple entries on ingress add paths berfore service path in
// order
func TestK8SServiceIngressAdditionalPathsHigherPriorityMultipleBeforeMainServicePath(t *testing.T) {
	t.Parallel()

	ingress := renderK8SServiceIngressWithSetValues(
		t,
		map[string]string{
			"ingress.enabled":     "true",
			"ingress.path":        "/app",
			"ingress.servicePort": "app",
			"ingress.additionalPathsHigherPriority[0].path":        "/sun",
			"ingress.additionalPathsHigherPriority[0].serviceName": "sun",
			"ingress.additionalPathsHigherPriority[0].servicePort": "3000",
			"ingress.additionalPathsHigherPriority[1].path":        "/black-hole",
			"ingress.additionalPathsHigherPriority[1].serviceName": "black-hole",
			"ingress.additionalPathsHigherPriority[1].servicePort": "80",
		},
	)
	pathRules := ingress.Spec.Rules[0].HTTP.Paths
	assert.Equal(t, len(pathRules), 3)

	// The first path should be the sun
	firstPath := pathRules[0]
	assert.Equal(t, firstPath.Path, "/sun")
	assert.Equal(t, firstPath.Backend.ServiceName, "sun")
	assert.Equal(t, firstPath.Backend.ServicePort.IntVal, int32(3000))

	// The second path should be the black hole
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/black-hole")
	assert.Equal(t, secondPath.Backend.ServiceName, "black-hole")
	assert.Equal(t, secondPath.Backend.ServicePort.IntVal, int32(80))

	// The last path should be the main service path
	thirdPath := pathRules[2]
	assert.Equal(t, thirdPath.Path, "/app")
	assert.Equal(t, strings.ToLower(thirdPath.Backend.ServiceName), "ingress-linter")
	assert.Equal(t, thirdPath.Backend.ServicePort.StrVal, "app")
}

// Test that omitting a serviceName on additionalPathsHigherPriority reuses the application service name
func TestK8SServiceIngressAdditionalPathsHigherPriorityNoServiceName(t *testing.T) {
	t.Parallel()

	ingress := renderK8SServiceIngressWithSetValues(
		t,
		map[string]string{
			"ingress.enabled":     "true",
			"ingress.path":        "/app",
			"ingress.servicePort": "app",
			"ingress.additionalPathsHigherPriority[0].path":        "/black-hole",
			"ingress.additionalPathsHigherPriority[0].servicePort": "3000",
		},
	)
	pathRules := ingress.Spec.Rules[0].HTTP.Paths
	assert.Equal(t, len(pathRules), 2)

	// The first path should be the black hole
	firstPath := pathRules[0]
	assert.Equal(t, firstPath.Path, "/black-hole")
	assert.Equal(t, strings.ToLower(firstPath.Backend.ServiceName), "ingress-linter")
	assert.Equal(t, firstPath.Backend.ServicePort.IntVal, int32(3000))

	// The second path should be the main service path
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/app")
	assert.Equal(t, strings.ToLower(secondPath.Backend.ServiceName), "ingress-linter")
	assert.Equal(t, secondPath.Backend.ServicePort.StrVal, "app")
}

// Test that omitting a serviceName on additionalPaths reuses the application service name even when hosts is set
func TestK8SServiceIngressWithHostsAdditionalPathsNoServiceName(t *testing.T) {
	t.Parallel()

	ingress := renderK8SServiceIngressWithSetValues(
		t,
		map[string]string{
			"ingress.enabled":                        "true",
			"ingress.path":                           "/app",
			"ingress.servicePort":                    "app",
			"ingress.hosts[0]":                       "chart-example.local",
			"ingress.additionalPaths[0].path":        "/black-hole",
			"ingress.additionalPaths[0].servicePort": "3000",
		},
	)
	pathRules := ingress.Spec.Rules[0].HTTP.Paths
	assert.Equal(t, len(pathRules), 2)

	// The first path should be the main service path
	firstPath := pathRules[0]
	assert.Equal(t, firstPath.Path, "/app")
	assert.Equal(t, strings.ToLower(firstPath.Backend.ServiceName), "ingress-linter")
	assert.Equal(t, firstPath.Backend.ServicePort.StrVal, "app")

	// The second path should be the black hole
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/black-hole")
	assert.Equal(t, strings.ToLower(secondPath.Backend.ServiceName), "ingress-linter")
	assert.Equal(t, secondPath.Backend.ServicePort.IntVal, int32(3000))
}

// Test that omitting a serviceName on additionalPathsHigherPriority reuses the application service name even when hosts
// is set
func TestK8SServiceIngressWithHostsAdditionalPathsHigherPriorityNoServiceName(t *testing.T) {
	t.Parallel()

	ingress := renderK8SServiceIngressWithSetValues(
		t,
		map[string]string{
			"ingress.enabled":     "true",
			"ingress.path":        "/app",
			"ingress.servicePort": "app",
			"ingress.hosts[0]":    "chart-example.local",
			"ingress.additionalPathsHigherPriority[0].path":        "/black-hole",
			"ingress.additionalPathsHigherPriority[0].servicePort": "3000",
		},
	)
	pathRules := ingress.Spec.Rules[0].HTTP.Paths
	assert.Equal(t, len(pathRules), 2)

	// The first path should be the black hole
	firstPath := pathRules[0]
	assert.Equal(t, firstPath.Path, "/black-hole")
	assert.Equal(t, strings.ToLower(firstPath.Backend.ServiceName), "ingress-linter")
	assert.Equal(t, firstPath.Backend.ServicePort.IntVal, int32(3000))

	// The second path should be the main service path
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/app")
	assert.Equal(t, strings.ToLower(secondPath.Backend.ServiceName), "ingress-linter")
	assert.Equal(t, secondPath.Backend.ServicePort.StrVal, "app")
}

// Test rendering Managed Certificate
func TestK8SServiceManagedCertDomainNameAndName(t *testing.T) {
	t.Parallel()

	cert := renderK8SServiceManagedCertificateWithSetValues(
		t,
		map[string]string{
			"google.managedCertificate.enabled":    "true",
			"google.managedCertificate.domainName": "api.acme.io",
			"google.managedCertificate.name":       "acme-cert",
		},
	)

	domains := cert.Spec.Domains
	certName := cert.ObjectMeta.Name
	assert.Equal(t, len(domains), 1)
	assert.Equal(t, domains[0], "api.acme.io")
	assert.Equal(t, certName, "acme-cert")
}

// Test that setting ingress.enabled = false will cause the helm template to not render the ManagedCertificate resource
func TestK8SServiceManagedCertificateDefaultsDoesNotCreateManagedCertificate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   map[string]string{},
	}
	_, err = helm.RenderTemplateE(t, options, helmChartPath, "gmc", []string{"templates/gmc.yaml"})
	require.Error(t, err)
}

// Test that omitting containerCommand does not set command attribute on the Deployment container spec.
func TestK8SServiceDefaultHasNullCommandSpec(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(t, map[string]string{})
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	assert.Nil(t, appContainer.Command)
}

// Test that setting containerCommand sets the command attribute on the Deployment container spec.
func TestK8SServiceWithContainerCommandHasCommandSpec(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"containerCommand[0]": "echo",
			"containerCommand[1]": "Hello world",
		},
	)
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	assert.Equal(t, appContainer.Command, []string{"echo", "Hello world"})
}

// Test that omitting aws.irsa.role_arn does not render the IRSA vars
func TestK8SServiceWithoutIRSA(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{},
	)
	renderedPodSpec := deployment.Spec.Template.Spec
	assert.Equal(t, len(renderedPodSpec.Volumes), 0)
	renderedPodContainers := renderedPodSpec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	assert.Equal(t, len(appContainer.Env), 0)
}

// Test that setting aws.irsa.role_arn renders the IRSA vars
func TestK8SServiceWithIRSA(t *testing.T) {
	t.Parallel()

	testRoleArn := "arn:aws:iam::123456789012:role/test-role"
	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"aws.irsa.role_arn": testRoleArn,
		},
	)
	renderedPodSpec := deployment.Spec.Template.Spec

	// Verify projected volume
	require.Equal(t, len(renderedPodSpec.Volumes), 1)
	volume := renderedPodSpec.Volumes[0]
	assert.Equal(t, volume.Name, "aws-iam-token")
	require.NotNil(t, volume.VolumeSource.Projected)
	projectedVolume := volume.VolumeSource.Projected
	require.Equal(t, len(projectedVolume.Sources), 1)
	projectedVolumeSource := projectedVolume.Sources[0]
	require.NotNil(t, projectedVolumeSource.ServiceAccountToken)
	assert.Equal(t, projectedVolumeSource.ServiceAccountToken.Audience, "sts.amazonaws.com")

	// Verify injected env vars
	renderedPodContainers := renderedPodSpec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	assert.Equal(t, len(appContainer.Env), 2)
	roleArnEnv := appContainer.Env[0]
	assert.Equal(t, roleArnEnv.Name, "AWS_ROLE_ARN")
	assert.Equal(t, roleArnEnv.Value, testRoleArn)
	tokenEnv := appContainer.Env[1]
	assert.Equal(t, tokenEnv.Name, "AWS_WEB_IDENTITY_TOKEN_FILE")
	assert.Equal(t, tokenEnv.Value, "/var/run/secrets/eks.amazonaws.com/serviceaccount/token")
}

// Test that providing tls configuration to Ingress renders correctly
func TestK8SServiceIngressMultiCert(t *testing.T) {
	t.Parallel()

	ingress := renderK8SServiceIngressWithSetValues(
		t,
		map[string]string{
			"ingress.enabled":           "true",
			"ingress.path":              "/app",
			"ingress.servicePort":       "app",
			"ingress.tls[0].secretName": "chart0-example-tls",
			"ingress.tls[0].hosts[0]":   "chart0-example-tls-host",
			"ingress.tls[1].secretName": "chart1-example-tls",
			"ingress.tls[1].hosts[0]":   "chart1-example-tls-host",
			"ingress.tls[1].hosts[1]":   "chart1-example-tls-host2",
		},
	)
	tls := ingress.Spec.TLS
	assert.Equal(t, len(tls), 2)

	// The first tls should be chart0
	firstTls := tls[0]
	assert.Equal(t, firstTls.SecretName, "chart0-example-tls")
	firstTlsHosts := firstTls.Hosts
	assert.Equal(t, len(firstTlsHosts), 1)
	assert.Equal(t, firstTlsHosts[0], "chart0-example-tls-host")

	// The second tls should be chart1 with multiple hosts
	secondTls := tls[1]
	assert.Equal(t, secondTls.SecretName, "chart1-example-tls")
	secondTlsHosts := secondTls.Hosts
	assert.Equal(t, len(secondTlsHosts), 2)
	assert.Equal(t, secondTlsHosts[0], "chart1-example-tls-host")
	assert.Equal(t, secondTlsHosts[1], "chart1-example-tls-host2")
}

func TestK8SServiceSideCarContainersRendersCorrectly(t *testing.T) {
	t.Parallel()
	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"sideCarContainers.datadog.image": "datadog/agent:latest",
		},
	)
	renderedContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedContainers), 2)
	sideCarContainer := renderedContainers[1]
	assert.Equal(t, sideCarContainer.Image, "datadog/agent:latest")
}

func TestK8SServiceDisableDefaultPort(t *testing.T) {
	t.Parallel()
	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"containerPorts.http.disabled": "true",
		},
	)
	renderedContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedContainers), 1)
	mainContainer := renderedContainers[0]
	assert.Equal(t, len(mainContainer.Ports), 0)
}

func TestK8SServiceCanaryDeploymentContainersLabeledCorrectly(t *testing.T) {
	t.Parallel()
	deployment := renderK8SServiceCanaryDeploymentWithSetValues(
		t,
		map[string]string{
			"canary.enabled":                   "true",
			"canary.replicaCount":              "1",
			"canary.containerImage.repository": "nginx",
			"canary.containerImage.tag":        "1.16.0",
		},
	)
	// Ensure a canary deployment has the canary deployment-type label
	assert.Equal(t, deployment.Spec.Selector.MatchLabels["gruntwork.io/deployment-type"], "canary")
}

func TestK8SServiceMainDeploymentContainersLabeledCorrectly(t *testing.T) {
	t.Parallel()
	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"containerImage.repository": "nginx",
			"containerImage.tag":        "1.16.0",
		},
	)
	// Ensure a "main" type deployment is properly labeled as such
	assert.Equal(t, deployment.Spec.Selector.MatchLabels["gruntwork.io/deployment-type"], "main")
}

func TestK8SServiceDeploymentAddingAdditionalLabels(t *testing.T) {
	t.Parallel()
	first_custom_deployment_label_value := "first-custom-value"
	second_custom_deployment_label_value := "second-custom-value"
	deployment := renderK8SServiceDeploymentWithSetValues(t,
		map[string]string{"additionalDeploymentLabels.first-label": first_custom_deployment_label_value,
			"additionalDeploymentLabels.second-label":second_custom_deployment_label_value})

	assert.Equal(t, deployment.Labels["first-label"], first_custom_deployment_label_value)
	assert.Equal(t, deployment.Labels["second-label"], second_custom_deployment_label_value)
}

func TestK8SServicePodAddingAdditionalLabels(t *testing.T) {
	t.Parallel()
	first_custom_pod_label_value := "first-custom-value"
	second_custom_pod_label_value := "second-custom-value"
	deployment := renderK8SServiceDeploymentWithSetValues(t,
		map[string]string{"additionalPodLabels.first-label":  first_custom_pod_label_value,
			"additionalPodLabels.second-label": second_custom_pod_label_value})

	assert.Equal(t, deployment.Spec.Template.Labels["first-label"], first_custom_pod_label_value)
	assert.Equal(t, deployment.Spec.Template.Labels["second-label"], second_custom_pod_label_value)
}

func TestK8SServiceDeploymentAddingPersistentVolumes(t *testing.T) {
	t.Parallel()

	volName := "pv-1"
	volClaim := "claim-1"
	volMountPath := "/mnt/path/1"

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"persistentVolumes.pv-1.claimName": volClaim,
			"persistentVolumes.pv-1.mountPath": volMountPath,
		},
	)

	// Verify that there is only one container and that the environments section is populated.
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)

	// Verify that a mount has been created for the PV
	mounts := renderedPodContainers[0].VolumeMounts
	assert.Equal(t, len(mounts), 1)
	mount := mounts[0]
	assert.Equal(t, volName, mount.Name)
	assert.Equal(t, volMountPath, mount.MountPath)

	// Verify that a volume has been declared for the PV
	volumes := deployment.Spec.Template.Spec.Volumes
	assert.Equal(t, len(volumes), 1)
	volume := volumes[0]
	assert.Equal(t, volName, volume.Name)
	assert.Equal(t, volClaim, volume.PersistentVolumeClaim.ClaimName)
}

func TestK8SServiceDeploymentStrategy(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"rollingUpdate.maxSurge":       "30%",
			"rollingUpdate.maxUnavailable": "20%",
		},
	)

	// Confirm that deploymentStrategy is set to RollingUpdate
	assert.NotNil(t, deployment.Spec.Strategy.RollingUpdate)

	// Confirm that maxSurge and maxUnavailable are equal to values set above
	rollingUpdateOptions := deployment.Spec.Strategy.RollingUpdate

	assert.Equal(t, rollingUpdateOptions.MaxSurge.String(), "30%")
	assert.Equal(t, rollingUpdateOptions.MaxUnavailable.String(), "20%")
}

func TestK8SServiceDeploymentRecreateStrategy(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"deploymentStrategy":           "Recreate",
			"rollingUpdate.maxSurge":       "30%",
			"rollingUpdate.maxUnavailable": "20%",
		},
	)

	// Confirm that RollingUpdate options are not injected if deployment strategy set to Recreate
	assert.Nil(t, deployment.Spec.Strategy.RollingUpdate)
}