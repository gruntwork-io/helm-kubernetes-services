// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

// Test that setting customResources.enabled = false will cause the helm template to not render any custom resources
func TestK8SServiceCustomResourcesEnabledFalseDoesNotCreateCustomResources(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   map[string]string{"customResources.enabled": "false"},
	}
	_, err = helm.RenderTemplateE(t, options, helmChartPath, "customresources", []string{"templates/customresources.yaml"})
	require.Error(t, err)
}

// Test that configuring a ConfigMap and a Secret will render correctly to something
func TestK8SServiceCustomResourcesEnabledCreatesCustomResources(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{
			filepath.Join("..", "charts", "k8s-service", "linter_values.yaml"),
			filepath.Join("fixtures", "custom_resources_values.yaml"),
		},
	}
	out := helm.RenderTemplate(t, options, helmChartPath, "customresources", []string{"templates/customresources.yaml"})

	// We render the output to a map to validate it
	renderedConfigMap := corev1.ConfigMap{}

	require.NoError(t, yaml.Unmarshal([]byte(out), &renderedConfigMap))
}
