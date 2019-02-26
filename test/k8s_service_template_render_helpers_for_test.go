// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
)

func renderK8SServiceDeploymentWithSetValues(t *testing.T, setValues map[string]string) appsv1.Deployment {
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   setValues,
	}
	// Render just the deployment resource
	out := helm.RenderTemplate(t, options, helmChartPath, []string{"templates/deployment.yaml"})

	// Parse the deployment and verify the preStop hook is set
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(t, out, &deployment)
	return deployment
}
