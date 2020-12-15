// +build all integration

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Test the base case of the k8s-service-custom-resources example.
// This test will:
//
// 1. Render a chart with multiple custom resources.
// 2. Run `kubectl apply` with the rendered chart.
// 3. Verify that the custom resources were deployed, by checking the k8s API.
func TestK8SServiceCustomResourcesExample(t *testing.T) {
	t.Parallel()

	// Setup paths for testing the example chart
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// Create a test namespace to deploy resources into, to avoid colliding with other tests
	kubectlOptions := k8s.NewKubectlOptions("", "", "")
	uniqueID := random.UniqueId()
	testNamespace := fmt.Sprintf("k8s-service-custom-resources-%s", strings.ToLower(uniqueID))
	k8s.CreateNamespace(t, kubectlOptions, testNamespace)
	defer k8s.DeleteNamespace(t, kubectlOptions, testNamespace)
	kubectlOptions.Namespace = testNamespace

	// Use the values file in the fixtures
	options := &helm.Options{
		ValuesFiles: []string{
			filepath.Join(helmChartPath, "linter_values.yaml"),
			filepath.Join("fixtures", "multiple_custom_resources_values.yaml"),
		},
	}

	// Render the chart
	out := helm.RenderTemplate(t, options, helmChartPath, "customresources", []string{"templates/customresources.yaml"})

	defer k8s.KubectlDeleteFromString(t, kubectlOptions, out)

	// Deploy a subset of the chart, just the ConfigMap and Secret
	k8s.KubectlApplyFromString(t, kubectlOptions, out)

	// Verify that ConfigMap and Secret got created, but do nothing with the output that is returned.
	// We only care that these functions do not error.
	k8s.GetSecret(t, kubectlOptions, "example-secret")
	getConfigMap(t, kubectlOptions, "example-config-map")
}

// getConfigMap should be implemented in Terratest
func getConfigMap(t *testing.T, options *k8s.KubectlOptions, name string) corev1.ConfigMap {
	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, options)
	require.NoError(t, err)

	configMap, err := clientset.CoreV1().ConfigMaps(options.Namespace).Get(name, metav1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, configMap)

	return *configMap
}
