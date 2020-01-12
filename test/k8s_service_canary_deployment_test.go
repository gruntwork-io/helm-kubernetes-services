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
	"github.com/stretchr/testify/assert"


	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

)

// Test that:
//
// 1. Setting the canary.enabled input variable results in canary pods being created
// 2. The deployment succeeds without errors
// 3. The canary pods are correctly labeled with gruntwork.io/deployment-type=canary
// 4. Enabling canary deployment does not interfere with main deployment. Main pods come up cleanly as well
// 5. As configured, the canary and main deployments are running separate image tags
func TestK8SServiceCanaryDeployment(t *testing.T) {
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
		SetValues: map[string]string{
			"containerImage.repository": "nginx",
			"containerImage.tag": "1.14.2",
			"containerImage.pullPolicy": "IfNotPresent",
			"applicationName": "canary-test",
			"replicaCount": "3",
			"canary.enabled": "true",
			"canary.replicaCount": "3",
			"canary.containerImage.repository": "nginx",
			"canary.containerImage.tag": "1.16.0",
			"livenessProbe.httpGet.path": "/",
			"livenessProbe.httpGet.port": "http",
			"readinessProbe.httpGet.path": "/",
			"readinessProbe.httpGet.port": "http",
			"service.type": "NodePort",
		},
	}

	defer helm.Delete(t, options, releaseName, true)
	helm.Install(t, options, helmChartPath, releaseName)

	// Uses label filters including gruntwork.io/deployment-type=canary to ensure the correct pods were created
	verifyCanaryAndMainPodsCreatedSuccessfully(t, kubectlOptions, "canary-test", releaseName)
	verifyAllPodsAvailable(t, kubectlOptions, "canary-test", releaseName, nginxValidationFunction)


	// Ensure that the canary deployment is running a separate tag from the main deployment, as configured
	canaryFilters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s,gruntwork.io/deployment-type=canary", "canary-test", releaseName),
	}

	canaryPods := k8s.ListPods(t, kubectlOptions, canaryFilters)
	canaryTag := canaryPods[0].Spec.Containers[0].Image

	mainFilters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s,gruntwork.io/deployment-type=main", "canary-test", releaseName),
	}

	mainPods := k8s.ListPods(t, kubectlOptions, mainFilters)
	mainTag := mainPods[0].Spec.Containers[0].Image

	fmt.Printf("Canary tag: %v and main tag: %v", canaryTag, mainTag)

	assert.NotEqual(t, canaryTag, mainTag)
}