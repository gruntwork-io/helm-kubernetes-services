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
)

// Test that:
//
// 1. We can deploy the example
// 2. The deployment succeeds without errors
// 3. We can open a port forward to one of the Pods and access nginx
// 4. We can access ngix via the service endpoint
func TestK8SServiceNginxExample(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)
	examplePath, err := filepath.Abs(filepath.Join("..", "examples", "k8s-service-nginx"))
	require.NoError(t, err)

	// Create a test namespace to deploy resources into, to avoid colliding with other tests
	kubectlOptions := k8s.NewKubectlOptions("", "")
	uniqueID := random.UniqueId()
	testNamespace := fmt.Sprintf("k8s-service-nginx-%s", strings.ToLower(uniqueID))
	k8s.CreateNamespace(t, kubectlOptions, testNamespace)
	defer k8s.DeleteNamespace(t, kubectlOptions, testNamespace)
	kubectlOptions.Namespace = testNamespace

	// Use the values file in the example and deploy the chart in the test namespace
	// Set a random release name
	releaseName := fmt.Sprintf("k8s-service-nginx-%s", strings.ToLower(uniqueID))
	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		ValuesFiles:    []string{filepath.Join(examplePath, "values.yaml")},
	}
	defer helm.Delete(t, options, releaseName, true)
	helm.Install(t, options, helmChartPath, releaseName)

	verifyPodsCreatedSuccessfully(t, kubectlOptions, "nginx", releaseName)
	verifyAllPodsAvailable(t, kubectlOptions, "nginx", releaseName, nginxValidationFunction)
	verifyServiceAvailable(t, kubectlOptions, "nginx", releaseName, nginxValidationFunction)
}

// nginxValidationFunction checks that we get a 200 response with the nginx welcome page.
func nginxValidationFunction(statusCode int, body string) bool {
	return statusCode == 200 && strings.Contains(body, "Welcome to nginx")
}
