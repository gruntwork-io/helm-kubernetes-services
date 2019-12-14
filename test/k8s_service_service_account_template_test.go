// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that setting serviceAccount.create = true will cause the helm template to render the Service Account resource
func TestK8SServiceAccountCreateTrueCreatesServiceAccount(t *testing.T) {
	t.Parallel()
	randomSAName := strings.ToLower(random.UniqueId())

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   map[string]string{"serviceAccount.name": randomSAName, "serviceAccount.create": "true"},
	}
	out := helm.RenderTemplate(t, options, helmChartPath, []string{"templates/serviceaccount.yaml"})

	// We take the output and render it to a map to validate it has created a service account output or not
	rendered := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(out), &rendered)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(rendered))
	assert.Equal(t, randomSAName, rendered["metadata"].(map[string]interface{})["name"])
}

// Test that setting serviceAccount.create = false will cause the helm template to not render the Service Account
// resource
func TestK8SServiceAccountCreateFalse(t *testing.T) {
	t.Parallel()
	randomSAName := strings.ToLower(random.UniqueId())

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   map[string]string{"serviceAccount.name": randomSAName, "serviceAccount.create": "false"},
	}
	out := helm.RenderTemplate(t, options, helmChartPath, []string{"templates/serviceaccount.yaml"})

	// We take the output and render it to a map to validate it has created a service account output or not
	rendered := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(out), &rendered)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(rendered))
}

func TestK8SServiceServiceAccountInjection(t *testing.T) {
	t.Parallel()
	randomSAName := strings.ToLower(random.UniqueId())
	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"serviceAccount.name": randomSAName,
		},
	)
	renderedServiceAccountName := deployment.Spec.Template.Spec.ServiceAccountName
	assert.Equal(t, renderedServiceAccountName, randomSAName)
}

func TestK8SServiceServiceAccountNoNameIsEmpty(t *testing.T) {
	t.Parallel()
	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{},
	)
	renderedServiceAccountName := deployment.Spec.Template.Spec.ServiceAccountName
	assert.Equal(t, renderedServiceAccountName, "")
}

func TestK8SServiceServiceAccountAutomountTokenTrueInjection(t *testing.T) {
	t.Parallel()
	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"serviceAccount.automountServiceAccountToken": "true",
		},
	)
	renderedServiceAccountTokenAutomountSetting := deployment.Spec.Template.Spec.AutomountServiceAccountToken
	require.NotNil(t, renderedServiceAccountTokenAutomountSetting)
	assert.True(t, *renderedServiceAccountTokenAutomountSetting)
}

func TestK8SServiceServiceAccountAutomountTokenFalseInjection(t *testing.T) {
	t.Parallel()
	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"serviceAccount.automountServiceAccountToken": "false",
		},
	)
	renderedServiceAccountTokenAutomountSetting := deployment.Spec.Template.Spec.AutomountServiceAccountToken
	require.NotNil(t, renderedServiceAccountTokenAutomountSetting)
	assert.False(t, *renderedServiceAccountTokenAutomountSetting)
}

func TestK8SServiceServiceAccountOmitAutomountToken(t *testing.T) {
	t.Parallel()
	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{},
	)
	renderedServiceAccountTokenAutomountSetting := deployment.Spec.Template.Spec.AutomountServiceAccountToken
	assert.Nil(t, renderedServiceAccountTokenAutomountSetting)
}
