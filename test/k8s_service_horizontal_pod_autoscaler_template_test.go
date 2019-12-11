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

// Test that setting horizontalPodAutoscaler.enabled = true will cause the helm template to render the Horizontal Pod
// Autoscaler resource
func TestK8SServiceHorizontalPodAutoscalerCreateTrueCreatesHorizontalPodAutoscaler(t *testing.T) {
	t.Parallel()
	minReplicas := "20"
	maxReplicas := "30"
	avgCpuUtil := "55"
	avgMemoryUtil := "65"

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues: map[string]string{
			"horizontalPodAutoscaler.enabled":              "true",
			"horizontalPodAutoscaler.minReplicas":          minReplicas,
			"horizontalPodAutoscaler.maxReplicas":          maxReplicas,
			"horizontalPodAutoscaler.avgCpuUtilization":    avgCpuUtil,
			"horizontalPodAutoscaler.avgMemoryUtilization": avgMemoryUtil,
		},
	}
	out := helm.RenderTemplate(t, options, helmChartPath, []string{"templates/horizontalpodautoscaler.yaml"})

	// We take the output and render it to a map to validate it has created a Horizontal Pod Autoscaler output or not
	rendered := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(out), &rendered)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(rendered))
	min, err := strconv.ParseFloat(minReplicas, 64)
	max, err := strconv.ParseFloat(maxReplicas, 64)
	avgCpu, err := strconv.ParseFloat(avgCpuUtil, 64)
	avgMem, err := strconv.ParseFloat(avgMemoryUtil, 64)
	assert.Equal(t, min, rendered["spec"].(map[string]interface{})["minReplicas"])
	assert.Equal(t, max, rendered["spec"].(map[string]interface{})["maxReplicas"])
	assert.Equal(t, avgCpu, rendered["spec"].(map[string]interface{})["metrics"].([]interface{})[0].(map[string]interface{})["resource"].(map[string]interface{})["target"].(map[string]interface{})["averageUtilization"])
	assert.Equal(t, avgMem, rendered["spec"].(map[string]interface{})["metrics"].([]interface{})[1].(map[string]interface{})["resource"].(map[string]interface{})["target"].(map[string]interface{})["averageUtilization"])
}

// Test that setting horizontalPodAutoscaler.enabled = false will cause the helm template to not render the Horizontal
// Pod Autoscaler resource
func TestK8SServiceHorizontalPodAutoscalerCreateFalse(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues: map[string]string{
			"horizontalPodAutoscaler.enabled": "false",
		},
	}
	out := helm.RenderTemplate(t, options, helmChartPath, []string{"templates/horizontalpodautoscaler.yaml"})

	// We take the output and render it to a map to validate it has created a Horizontal Pod Autoscaler output or not
	rendered := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(out), &rendered)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(rendered))
}
