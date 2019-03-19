// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
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
	out := helm.RenderTemplate(t, options, helmChartPath, []string{"templates/ingress.yaml"})

	// We take the output and render it to a map to validate it is an empty yaml
	rendered := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(out), &rendered)
	assert.NoError(t, err)
	assert.Equal(t, len(rendered), 0)
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
	out := helm.RenderTemplate(t, options, helmChartPath, []string{"templates/service.yaml"})

	// We take the output and render it to a map to validate it is an empty yaml
	rendered := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(out), &rendered)
	assert.NoError(t, err)
	assert.Equal(t, len(rendered), 0)
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
			_, err := helm.RenderTemplateE(t, options, helmChartPath, []string{})
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
			helm.RenderTemplate(t, options, helmChartPath, []string{})
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
