//go:build all || integration
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
// 1. Setting the canary.enabled input variable results in canary pods being created
// 2. The statefulset deployment succeeds without errors
// 3. The canary pods are correctly labeled with gruntwork.io/deployment-type=canary
// 4. Enabling canary deployment does not interfere with main deployment. Main pods come up cleanly as well
// 5. As configured, the canary and main deployments are running separate image tags
func TestK8SServiceCanaryStatefulset(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// Create a test namespace to deploy resources into, to avoid colliding with other tests
	kubectlOptions := k8s.NewKubectlOptions("", "", "")
	uniqueID := random.UniqueId()
	testNamespace := fmt.Sprintf("k8s-service-canary-%s", strings.ToLower(uniqueID))
	k8s.CreateNamespace(t, kubectlOptions, testNamespace)
	defer k8s.DeleteNamespace(t, kubectlOptions, testNamespace)
	kubectlOptions.Namespace = testNamespace

	// Use the values file in the example and deploy the chart in the test namespace
	// Set a random release name
	releaseName := fmt.Sprintf("k8s-service-canary-%s", strings.ToLower(uniqueID))
	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		ValuesFiles: []string{
			filepath.Join("..", "charts", "k8s-service", "linter_values.yaml"),
			filepath.Join("fixtures", "canary_and_main_statefulset_values.yaml"),
		},
		SetValues: map[string]string{"workloadType": "statefulset"},
	}

	defer helm.Delete(t, options, releaseName, true)
	helm.Install(t, options, helmChartPath, releaseName)

	// Uses label filters including gruntwork.io/deployment-type=canary to ensure the correct pods were created
	verifyCanaryAndMainPodsCreatedSuccessfully(t, kubectlOptions, "canary-test", releaseName)
	verifyAllPodsAvailable(t, kubectlOptions, "canary-test", releaseName, nginxValidationFunction)
	verifyDifferentContainerTagsForCanaryPods(t, kubectlOptions, releaseName)
	verifyServiceRoutesToMainAndCanaryPods(t, kubectlOptions, "canary-test", releaseName)
}
