//go:build all || tpl
// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that setting verticalPodAutoscaler.enabled = true will cause the helm template to render the Vertical Pod
// Autoscaler resource with main pod configuration
func TestK8SServiceVerticalPodAutoscalerCreateTrueCreatesVerticalPodAutoscalerWithMainPodConfiguration(t *testing.T) {
	t.Parallel()
	updateMode := "Initial"
	minReplicas := "20"
	controlledValues := "RequestsOnly"

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues: map[string]string{
			"verticalPodAutoscaler.enabled":                                      "true",
			"verticalPodAutoscaler.updateMode":                                   updateMode,
			"verticalPodAutoscaler.minReplicas":                                  minReplicas,
			"verticalPodAutoscaler.mainContainerResourcePolicy.controlledValues": controlledValues,
		},
	}
	out := helm.RenderTemplate(t, options, helmChartPath, "hpa", []string{"templates/verticalpodautoscaler.yaml"})

	// We take the output and render it to a map to validate it has created a Vertical Pod Autoscaler output or not
	rendered := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(out), &rendered)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(rendered))
	min, err := strconv.ParseFloat(minReplicas, 64)
	assert.Equal(t, updateMode, rendered["spec"].(map[string]interface{})["updatePolicy"].(map[string]interface{})["updateMode"])
	assert.Equal(t, min, rendered["spec"].(map[string]interface{})["updatePolicy"].(map[string]interface{})["minReplicas"])
	assert.Equal(t, controlledValues, rendered["spec"].(map[string]interface{})["resourcePolicy"].(map[string]interface{})["containerPolicies"].([]interface{})[0].(map[string]interface{})["controlledValues"])
}

// Test that setting verticalPodAutoscaler.enabled = false will cause the helm template to not render the Vertical
// Pod Autoscaler resource
func TestK8SServiceVerticalPodAutoscalerCreateFalse(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues: map[string]string{
			"verticalPodAutoscaler.enabled": "false",
		},
	}
	_, err = helm.RenderTemplateE(t, options, helmChartPath, "hpa", []string{"templates/verticalpodautoscaler.yaml"})
	require.Error(t, err)
}

// Test that setting verticalPodAutoscaler.enabled = true will cause the helm template to render the Vertical Pod
// Autoscaler resource with maxAllowed
func TestK8SServiceVerticalPodAutoscalerCreateTrueCreatesVerticalPodAutoscalerWithMinAllowed(t *testing.T) {
	t.Parallel()
	maxCPU := "1000m"
	maxMemory := "1000Mi"

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues: map[string]string{
			"verticalPodAutoscaler.enabled":                                       "true",
			"verticalPodAutoscaler.mainContainerResourcePolicy.maxAllowed.cpu":    maxCPU,
			"verticalPodAutoscaler.mainContainerResourcePolicy.maxAllowed.memory": maxMemory,
		},
	}
	out := helm.RenderTemplate(t, options, helmChartPath, "hpa", []string{"templates/verticalpodautoscaler.yaml"})

	// We take the output and render it to a map to validate it has created a Vertical Pod Autoscaler output or not
	rendered := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(out), &rendered)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(rendered))
	assert.Equal(t, maxCPU, rendered["spec"].(map[string]interface{})["resourcePolicy"].(map[string]interface{})["containerPolicies"].([]interface{})[0].(map[string]interface{})["maxAllowed"].(map[string]interface{})["cpu"])
	assert.Equal(t, maxMemory, rendered["spec"].(map[string]interface{})["resourcePolicy"].(map[string]interface{})["containerPolicies"].([]interface{})[0].(map[string]interface{})["maxAllowed"].(map[string]interface{})["memory"])
}

// Test that setting verticalPodAutoscaler.enabled = true will cause the helm template to render the Vertical Pod
// Autoscaler resource with minAllowed
func TestK8SServiceVerticalPodAutoscalerCreateTrueCreatesVerticalPodAutoscalerWithMaxAllowed(t *testing.T) {
	t.Parallel()
	minCPU := "1000m"
	minMemory := "1000Mi"

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues: map[string]string{
			"verticalPodAutoscaler.enabled":                                       "true",
			"verticalPodAutoscaler.mainContainerResourcePolicy.minAllowed.cpu":    minCPU,
			"verticalPodAutoscaler.mainContainerResourcePolicy.minAllowed.memory": minMemory,
		},
	}
	out := helm.RenderTemplate(t, options, helmChartPath, "hpa", []string{"templates/verticalpodautoscaler.yaml"})

	// We take the output and render it to a map to validate it has created a Vertical Pod Autoscaler output or not
	rendered := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(out), &rendered)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(rendered))
	assert.Equal(t, minCPU, rendered["spec"].(map[string]interface{})["resourcePolicy"].(map[string]interface{})["containerPolicies"].([]interface{})[0].(map[string]interface{})["minAllowed"].(map[string]interface{})["cpu"])
	assert.Equal(t, minMemory, rendered["spec"].(map[string]interface{})["resourcePolicy"].(map[string]interface{})["containerPolicies"].([]interface{})[0].(map[string]interface{})["minAllowed"].(map[string]interface{})["memory"])
}

// // Test that setting verticalPodAutoscaler.enabled = true will cause the helm template to render the Vertical Pod
// // Autoscaler resource updateMode = "Off"
func TestK8SServiceVerticalPodAutoscalerCreateTrueCreatesVerticalPodAutoscalerWithUpdateModeOff(t *testing.T) {
	t.Parallel()
	updateMode := "Off"

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues: map[string]string{
			"verticalPodAutoscaler.enabled": "true",
		},
	}
	out := helm.RenderTemplate(t, options, helmChartPath, "hpa", []string{"templates/verticalpodautoscaler.yaml"})

	// We take the output and render it to a map to validate it has created a Vertical Pod Autoscaler output or not
	rendered := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(out), &rendered)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(rendered))
	assert.Equal(t, updateMode, rendered["spec"].(map[string]interface{})["updatePolicy"].(map[string]interface{})["updateMode"])
}
