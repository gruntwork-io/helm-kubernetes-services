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
)

// Test that setting serviceAccount.create = true will cause the helm template to render the Service Account resource
func TestK8SServiceAccountCreateTrueCreatesServiceAccount(t *testing.T) {
	t.Parallel()
	randomSAName := strings.ToLower(random.UniqueId())

	serviceaccount := renderK8SServiceAccountWithSetValues(
		t,
		map[string]string{
			"serviceAccount.create": "true",
			"serviceAccount.name":   randomSAName,
		},
	)

	assert.Equal(t, serviceaccount.Name, randomSAName)
}

// Test that setting serviceAccount.create = false will cause the helm template to not render the Service Account
// resource
func TestK8SServiceAccountCreateFalse(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues: map[string]string{
			"serviceAccount.create": "false",
		},
	}
	_, err = helm.RenderTemplateE(t, options, helmChartPath, "serviceaccount", []string{"templates/serviceaccount.yaml"})
	require.Error(t, err)
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

// Test that the Annotations of a service account are correctly rendered
func TestK8SServiceAccountAnnotationRendering(t *testing.T) {
	t.Parallel()

	serviceAccountAnnotationKey := "testAnnotation"
	serviceAccountAnnotationValue := strings.ToLower(random.UniqueId())

	serviceaccount := renderK8SServiceAccountWithSetValues(
		t,
		map[string]string{
			"serviceAccount.create": "true",
			"serviceAccount.annotations." + serviceAccountAnnotationKey: serviceAccountAnnotationValue,
		},
	)

	renderedAnnotation := serviceaccount.Annotations
	assert.Equal(t, len(renderedAnnotation), 1)
	assert.Equal(t, renderedAnnotation[serviceAccountAnnotationKey], serviceAccountAnnotationValue)
}

// Test that default imagePullSecrets do not render any
func TestK8SServiceAccountNoImagePullSecrets(t *testing.T) {
	t.Parallel()

	serviceaccount := renderK8SServiceAccountWithSetValues(
		t,
		map[string]string{
			"serviceAccount.create": "true",
		},
	)

	renderedImagePullSecrets := serviceaccount.ImagePullSecrets
	require.Equal(t, len(renderedImagePullSecrets), 0)
}

// Test that multiple imagePullSecrets renders each one correctly
func TestK8SServiceAccountMultipleImagePullSecrets(t *testing.T) {
	t.Parallel()

	serviceaccount := renderK8SServiceAccountWithSetValues(
		t,
		map[string]string{
			"serviceAccount.create": "true",
			"imagePullSecrets[0]":   "docker-private-registry-key",
			"imagePullSecrets[1]":   "gcr-registry-key",
		},
	)

	renderedImagePullSecrets := serviceaccount.ImagePullSecrets
	require.Equal(t, len(renderedImagePullSecrets), 2)
	assert.Equal(t, renderedImagePullSecrets[0].Name, "docker-private-registry-key")
	assert.Equal(t, renderedImagePullSecrets[1].Name, "gcr-registry-key")
}
