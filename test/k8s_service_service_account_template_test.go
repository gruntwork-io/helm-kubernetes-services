// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
