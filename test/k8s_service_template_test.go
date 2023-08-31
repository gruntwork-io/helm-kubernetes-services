//go:build all || tpl
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
	policyv1beta1 "k8s.io/api/policy/v1beta1"
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

func TestK8SServicePodSecurityContextAnnotationRenderCorrectly(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"podSecurityContext.fsGroup": "2000",
		},
	)
	renderedPodSpec := deployment.Spec.Template.Spec
	assert.NotNil(t, renderedPodSpec.SecurityContext)
	assert.Equal(t, *renderedPodSpec.SecurityContext.FSGroup, int64(2000))
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

func TestK8SServiceIngressPortNumberTypeConversionWithValues(t *testing.T) {
	t.Parallel()

	ingress := renderK8SServiceIngressWithValuesFile(t, filepath.Join("fixtures", "ingress_values_with_number_port.yaml"))
	pathRules := ingress.Spec.Rules[0].HTTP.Paths
	assert.Equal(t, len(pathRules), 2)

	// The first path should be the main service path
	firstPath := pathRules[0]
	assert.Equal(t, firstPath.Path, "/app")
	assert.Equal(t, strings.ToLower(firstPath.Backend.Service.Name), "ingress-linter")
	assert.Equal(t, firstPath.Backend.Service.Port.Number, int32(80))

	// The second path should be the black hole
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/black-hole")
	assert.Equal(t, secondPath.Backend.Service.Name, "black-hole")
	assert.Equal(t, secondPath.Backend.Service.Port.Number, int32(80))
}

func TestK8SServiceIngressPortStringTypeConversionWithValues(t *testing.T) {
	t.Parallel()

	ingress := renderK8SServiceIngressWithValuesFile(t, filepath.Join("fixtures", "ingress_values_with_name_port.yaml"))
	pathRules := ingress.Spec.Rules[0].HTTP.Paths
	assert.Equal(t, len(pathRules), 2)

	// The first path should be the main service path
	firstPath := pathRules[0]
	assert.Equal(t, firstPath.Path, "/app")
	assert.Equal(t, strings.ToLower(firstPath.Backend.Service.Name), "ingress-linter")
	assert.Equal(t, firstPath.Backend.Service.Port.Name, "app")

	// The second path should be the black hole
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/black-hole")
	assert.Equal(t, secondPath.Backend.Service.Name, "black-hole")
	assert.Equal(t, secondPath.Backend.Service.Port.Name, "black-hole")
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
	assert.Equal(t, strings.ToLower(firstPath.Backend.Service.Name), "ingress-linter")
	assert.Equal(t, firstPath.Backend.Service.Port.Name, "app")

	// The second path should be the black hole
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/black-hole")
	assert.Equal(t, secondPath.Backend.Service.Name, "black-hole")
	assert.Equal(t, secondPath.Backend.Service.Port.Number, int32(80))
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
	assert.Equal(t, strings.ToLower(firstPath.Backend.Service.Name), "ingress-linter")
	assert.Equal(t, firstPath.Backend.Service.Port.Name, "app")

	// The second path should be the sun
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/sun")
	assert.Equal(t, secondPath.Backend.Service.Name, "sun")
	assert.Equal(t, secondPath.Backend.Service.Port.Number, int32(3000))

	// The third path should be the black hole
	thirdPath := pathRules[2]
	assert.Equal(t, thirdPath.Path, "/black-hole")
	assert.Equal(t, thirdPath.Backend.Service.Name, "black-hole")
	assert.Equal(t, thirdPath.Backend.Service.Port.Number, int32(80))
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
	assert.Equal(t, strings.ToLower(firstPath.Backend.Service.Name), "ingress-linter")
	assert.Equal(t, firstPath.Backend.Service.Port.Name, "app")

	// The second path should be the black hole
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/black-hole")
	assert.Equal(t, strings.ToLower(secondPath.Backend.Service.Name), "ingress-linter")
	assert.Equal(t, secondPath.Backend.Service.Port.Number, int32(3000))
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
	assert.Equal(t, firstPath.Backend.Service.Name, "black-hole")
	assert.Equal(t, firstPath.Backend.Service.Port.Number, int32(80))

	// The second path should be the main service path
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/app")
	assert.Equal(t, strings.ToLower(secondPath.Backend.Service.Name), "ingress-linter")
	assert.Equal(t, secondPath.Backend.Service.Port.Name, "app")
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
	assert.Equal(t, firstPath.Backend.Service.Name, "sun")
	assert.Equal(t, firstPath.Backend.Service.Port.Number, int32(3000))

	// The second path should be the black hole
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/black-hole")
	assert.Equal(t, secondPath.Backend.Service.Name, "black-hole")
	assert.Equal(t, secondPath.Backend.Service.Port.Number, int32(80))

	// The last path should be the main service path
	thirdPath := pathRules[2]
	assert.Equal(t, thirdPath.Path, "/app")
	assert.Equal(t, strings.ToLower(thirdPath.Backend.Service.Name), "ingress-linter")
	assert.Equal(t, thirdPath.Backend.Service.Port.Name, "app")
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
	assert.Equal(t, strings.ToLower(firstPath.Backend.Service.Name), "ingress-linter")
	assert.Equal(t, firstPath.Backend.Service.Port.Number, int32(3000))

	// The second path should be the main service path
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/app")
	assert.Equal(t, strings.ToLower(secondPath.Backend.Service.Name), "ingress-linter")
	assert.Equal(t, secondPath.Backend.Service.Port.Name, "app")
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
	assert.Equal(t, strings.ToLower(firstPath.Backend.Service.Name), "ingress-linter")
	assert.Equal(t, firstPath.Backend.Service.Port.Name, "app")

	// The second path should be the black hole
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/black-hole")
	assert.Equal(t, strings.ToLower(secondPath.Backend.Service.Name), "ingress-linter")
	assert.Equal(t, secondPath.Backend.Service.Port.Number, int32(3000))
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
	assert.Equal(t, strings.ToLower(firstPath.Backend.Service.Name), "ingress-linter")
	assert.Equal(t, firstPath.Backend.Service.Port.Number, int32(3000))

	// The second path should be the main service path
	secondPath := pathRules[1]
	assert.Equal(t, secondPath.Path, "/app")
	assert.Equal(t, strings.ToLower(secondPath.Backend.Service.Name), "ingress-linter")
	assert.Equal(t, secondPath.Backend.Service.Port.Name, "app")
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

// Test that omitting containerArgs does not set args attribute on the Deployment container spec.
func TestK8SServiceDefaultHasNullArgSpec(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(t, map[string]string{})
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	assert.Nil(t, appContainer.Args)
}

// Test that setting containerCommand sets the command attribute on the Deployment container spec.
func TestK8SServiceWithContainerArgsHasArgsSpec(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"containerArgs[0]": "echo",
			"containerArgs[1]": "Hello world",
		},
	)
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	assert.Equal(t, appContainer.Args, []string{"echo", "Hello world"})
}

// Test that omitting hostAliases does not set hostAliases attribute on the Deployment container spec.
func TestK8SServiceDefaultHasNullHostAliasesSpec(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(t, map[string]string{})
	renderedPodSpec := deployment.Spec.Template.Spec
	assert.Nil(t, renderedPodSpec.HostAliases)
}

// Test that setting hostAliases sets the hostAliases attribute on the Deployment container spec.
func TestK8SServiceWithHostAliasesHasHostAliasesSpec(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"hostAliases[0].ip":           "127.0.0.1",
			"hostAliases[0].hostnames[0]": "foo.local",
			"hostAliases[0].hostnames[1]": "bar.local",
			"hostAliases[1].ip":           "10.1.2.3",
			"hostAliases[1].hostnames[0]": "foo.remote",
			"hostAliases[1].hostnames[1]": "bar.remote",
		},
	)
	renderedPodSpec := deployment.Spec.Template.Spec
	assert.Equal(t, len(renderedPodSpec.HostAliases), 2)
	// order should be preserved, since order is important for /etc/hosts
	assert.Equal(t, renderedPodSpec.HostAliases[0].IP, "127.0.0.1")
	assert.Equal(t, renderedPodSpec.HostAliases[0].Hostnames, []string{"foo.local", "bar.local"})
	assert.Equal(t, renderedPodSpec.HostAliases[1].IP, "10.1.2.3")
	assert.Equal(t, renderedPodSpec.HostAliases[1].Hostnames, []string{"foo.remote", "bar.remote"})
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

func TestK8SServiceInitContainersRendersCorrectly(t *testing.T) {
	t.Parallel()
	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"initContainers.flyway.image": "flyway/flyway",
		},
	)
	renderedContainers := deployment.Spec.Template.Spec.InitContainers
	require.Equal(t, len(renderedContainers), 1)
	assert.Equal(t, renderedContainers[0].Image, "flyway/flyway")
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
			"additionalDeploymentLabels.second-label": second_custom_deployment_label_value})

	assert.Equal(t, deployment.Labels["first-label"], first_custom_deployment_label_value)
	assert.Equal(t, deployment.Labels["second-label"], second_custom_deployment_label_value)
}

func TestK8SServicePodAddingAdditionalLabels(t *testing.T) {
	t.Parallel()
	first_custom_pod_label_value := "first-custom-value"
	second_custom_pod_label_value := "second-custom-value"
	deployment := renderK8SServiceDeploymentWithSetValues(t,
		map[string]string{"additionalPodLabels.first-label": first_custom_pod_label_value,
			"additionalPodLabels.second-label": second_custom_pod_label_value})

	assert.Equal(t, deployment.Spec.Template.Labels["first-label"], first_custom_pod_label_value)
	assert.Equal(t, deployment.Spec.Template.Labels["second-label"], second_custom_pod_label_value)
}

func TestK8SServiceDeploymentStrategyOnlySetIfEnabled(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"deploymentStrategy.enabled": "false",
		},
	)

	// Strategy shouldn't be set
	assert.Equal(t, "", string(deployment.Spec.Strategy.Type))
	assert.Nil(t, deployment.Spec.Strategy.RollingUpdate)
}

func TestK8SServiceDeploymentRollingUpdateStrategy(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"deploymentStrategy.enabled": "true",
			"deploymentStrategy.type":    "RollingUpdate",
		},
	)

	assert.EqualValues(t, "RollingUpdate", string(deployment.Spec.Strategy.Type))
	require.Nil(t, deployment.Spec.Strategy.RollingUpdate)
}

func TestK8SServiceDeploymentRollingUpdateStrategyWithCustomOptions(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"deploymentStrategy.enabled":                      "true",
			"deploymentStrategy.type":                         "RollingUpdate",
			"deploymentStrategy.rollingUpdate.maxSurge":       "30%",
			"deploymentStrategy.rollingUpdate.maxUnavailable": "20%",
		},
	)

	assert.EqualValues(t, "RollingUpdate", string(deployment.Spec.Strategy.Type))

	rollingUpdateOptions := deployment.Spec.Strategy.RollingUpdate
	require.NotNil(t, rollingUpdateOptions)
	assert.Equal(t, rollingUpdateOptions.MaxSurge.String(), "30%")
	assert.Equal(t, rollingUpdateOptions.MaxUnavailable.String(), "20%")
}

func TestK8SServiceDeploymentRecreateStrategy(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"deploymentStrategy.enabled": "true",
			"deploymentStrategy.type":    "Recreate",
		},
	)

	assert.Equal(t, "Recreate", string(deployment.Spec.Strategy.Type))
	assert.Nil(t, deployment.Spec.Strategy.RollingUpdate)

	// Test that custom rolling update options are ignore if the strategy is set to recreate
	deployment = renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"deploymentStrategy.enabled":                      "true",
			"deploymentStrategy.type":                         "Recreate",
			"deploymentStrategy.rollingUpdate.maxSurge":       "30%",
			"deploymentStrategy.rollingUpdate.maxUnavailable": "20%",
		},
	)

	assert.Equal(t, "Recreate", string(deployment.Spec.Strategy.Type))
	assert.Nil(t, deployment.Spec.Strategy.RollingUpdate)
}

func TestK8SServiceFullnameOverride(t *testing.T) {
	t.Parallel()

	overiddenName := "overidden-name"

	deployment := renderK8SServiceDeploymentWithSetValues(t,
		map[string]string{
			"fullnameOverride": overiddenName,
		},
	)

	assert.Equal(t, deployment.Name, overiddenName)
}

func TestK8SServiceEnvFrom(t *testing.T) {
	t.Parallel()

	t.Run("BothConfigMapsAndSecretsEnvFrom", func(t *testing.T) {
		deployment := renderK8SServiceDeploymentWithSetValues(t,
			map[string]string{
				"configMaps.test-configmap.as": "envFrom",
				"secrets.test-secret.as":       "envFrom",
			},
		)

		assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].EnvFrom)
		assert.Equal(t, len(deployment.Spec.Template.Spec.Containers[0].EnvFrom), 2)
		assert.Equal(t, deployment.Spec.Template.Spec.Containers[0].EnvFrom[0].ConfigMapRef.Name, "test-configmap")
		assert.Equal(t, deployment.Spec.Template.Spec.Containers[0].EnvFrom[1].SecretRef.Name, "test-secret")
	})

	t.Run("OnlyConfigMapsEnvFrom", func(t *testing.T) {
		deployment := renderK8SServiceDeploymentWithSetValues(t,
			map[string]string{
				"configMaps.test-configmap.as": "envFrom",
			},
		)

		assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].EnvFrom)
		assert.Equal(t, len(deployment.Spec.Template.Spec.Containers[0].EnvFrom), 1)
		assert.Equal(t, deployment.Spec.Template.Spec.Containers[0].EnvFrom[0].ConfigMapRef.Name, "test-configmap")
	})

	t.Run("OnlySecretsEnvFrom", func(t *testing.T) {
		deployment := renderK8SServiceDeploymentWithSetValues(t,
			map[string]string{
				"secrets.test-secret.as": "envFrom",
			},
		)

		assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].EnvFrom)
		assert.Equal(t, len(deployment.Spec.Template.Spec.Containers[0].EnvFrom), 1)
		assert.Equal(t, deployment.Spec.Template.Spec.Containers[0].EnvFrom[0].SecretRef.Name, "test-secret")
	})
}

func TestK8SServiceMinPodsAvailableZeroMeansNoPDB(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   map[string]string{"minPodsAvailable": "0"},
	}
	_, err = helm.RenderTemplateE(t, options, helmChartPath, "pdb", []string{"templates/pdb.yaml"})
	require.Error(t, err)
}

func TestK8SServiceMinPodsAvailableGreaterThanZeroMeansPDB(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   map[string]string{"minPodsAvailable": "1"},
	}
	out := helm.RenderTemplate(t, options, helmChartPath, "pdb", []string{"templates/pdb.yaml"})

	var pdb policyv1beta1.PodDisruptionBudget
	helm.UnmarshalK8SYaml(t, out, &pdb)
	assert.Equal(t, 1, pdb.Spec.MinAvailable.IntValue())
}

// Test that rendering extensions.v1beta1 Ingress works.
func TestK8SServiceRenderExtV1Beta1Ingress(t *testing.T) {
	t.Parallel()

	ingress := renderK8SServiceExtV1Beta1IngressWithSetValues(
		t,
		map[string]string{
			"kubeVersionOverride":                    "1.17.0",
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

// Test that sessionAffinity and sessionAffinityConfig render correctly when set
func TestK8SServiceSessionAffinityConfig(t *testing.T) {
	t.Parallel()

	service := renderK8SServiceWithSetValues(
		t,
		map[string]string{
			"service.sessionAffinity":                               "ClientIP",
			"service.sessionAffinityConfig.clientIP.timeoutSeconds": "10800",
		},
	)

	assert.Equal(t, corev1.ServiceAffinity("ClientIP"), service.Spec.SessionAffinity)
	assert.Equal(t, int32(10800), *service.Spec.SessionAffinityConfig.ClientIP.TimeoutSeconds)
}

// Test that externalTrafficPolicy is correctly set
func TestK8SServiceExternalTrafficPolicy(t *testing.T) {
	t.Parallel()

	service := renderK8SServiceWithSetValues(
		t,
		map[string]string{
			"service.externalTrafficPolicy": "Local",
		},
	)

	assert.Equal(t, corev1.ServiceExternalTrafficPolicyType("Local"), service.Spec.ExternalTrafficPolicy)
}

// Test that internalTrafficPolicy is correctly set
func TestK8SServiceInternalTrafficPolicy(t *testing.T) {
	t.Parallel()

	service := renderK8SServiceWithSetValues(
		t,
		map[string]string{
			"service.internalTrafficPolicy": "Local",
		},
	)

	assert.Equal(t, corev1.ServiceInternalTrafficPolicyType("Local"), *service.Spec.InternalTrafficPolicy)
}

// Test that sessionAffinity and sessionAffinityConfig are not rendered if not set
func TestK8SServiceSessionAffinityOnlySetIfDefined(t *testing.T) {
	t.Parallel()

	service := renderK8SServiceWithSetValues(
		t,
		map[string]string{},
	)

	// SessionAffinity and SessionAffinityConfig shouldn't be set
	assert.Equal(t, corev1.ServiceAffinity(""), service.Spec.SessionAffinity)
	assert.Nil(t, service.Spec.SessionAffinityConfig)
}

// Test that clusterIP is rendered correctly when it is set.
func TestK8SServiceClusterIP(t *testing.T) {
	t.Parallel()

	testCases := []string{
		// Unset
		"",
		// headless service:
		// https://kubernetes.io/docs/concepts/services-networking/service/#headless-services
		"None",
		// Some random IP
		"192.168.0.42",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			values := make(map[string]string)
			if tc != "" {
				values["service.clusterIP"] = tc
			}

			service := renderK8SServiceWithSetValues(t, values)
			assert.Equal(t, tc, service.Spec.ClusterIP)
		})
	}
}
