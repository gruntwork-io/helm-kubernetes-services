//go:build all || tpl
// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that setting canary.enabled = false will cause the helm template to not render the canary Statefulset resource
func TestK8SServiceCanaryEnabledFalseDoesNotCreateCanaryStatefulset(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   map[string]string{"canary.enabled": "false", "workloadType": "statefulset"},
	}
	_, err = helm.RenderTemplateE(t, options, helmChartPath, "canary", []string{"templates/canarystatefulset.yaml"})
	require.Error(t, err)
}

// Test that configuring a canary statefulset will render to a manifest with a container that is clearly a canary
func TestK8SServiceCanaryEnabledCreatesCanaryStatefulset(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{
			filepath.Join("..", "charts", "k8s-service", "linter_values.yaml"),
			filepath.Join("fixtures", "canary_statefulset_values.yaml"),
		},
		SetValues: map[string]string{"workloadType": "statefulset"},
	}
	out := helm.RenderTemplate(t, options, helmChartPath, "canary", []string{"templates/canarystatefulset.yaml"})
	
	// We take the output and render it to a map to validate it has created a canary deployment or not
	rendered := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(out), &rendered)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(rendered))

	// Inspect the name of the rendered canary deployment
	nameField := rendered["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["name"]
	nameString := fmt.Sprintf("%v", nameField)

	// Ensure the name contains the string "-canary"
	assert.True(t, strings.Contains(string(nameString), "-canary"))
}
