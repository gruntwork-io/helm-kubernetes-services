// +build all integration

// NOTE: We use build flags to differentiate between template test and integration test so that you can conveniently
// run just the template test. See the test README for more information.

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
// 1. The statefulset deployment succeeds without errors
// 2. Main pods come up cleanly
func TestK8SServiceStatefulset(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// Create a test namespace to deploy resources into, to avoid colliding with other tests
	kubectlOptions := k8s.NewKubectlOptions("", "", "")
	uniqueID := random.UniqueId()
	testNamespace := fmt.Sprintf("k8s-service-%s", strings.ToLower(uniqueID))
	k8s.CreateNamespace(t, kubectlOptions, testNamespace)
	defer k8s.DeleteNamespace(t, kubectlOptions, testNamespace)
	kubectlOptions.Namespace = testNamespace

	// Use the values file in the example and deploy the chart in the test namespace
	// Set a random release name
	releaseName := fmt.Sprintf("k8s-service-%s", strings.ToLower(uniqueID))
	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		ValuesFiles: []string{
			filepath.Join("..", "charts", "k8s-service", "linter_values.yaml"),
			filepath.Join("fixtures", "main_statefulset_values.yaml"),
		},
	}

	defer helm.Delete(t, options, releaseName, true)
	helm.Install(t, options, helmChartPath, releaseName)

	// Uses label filters to ensure the correct pods were created
	verifyMainPodsCreatedSuccessfully(t, kubectlOptions, "sts-test", releaseName)
	verifyAllPodsAvailable(t, kubectlOptions, "sts-test", releaseName, nginxValidationFunction)
	verifyServiceRoutesToMainStsPods(t, kubectlOptions, "sts-test", releaseName)
}
