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
	"github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Test that:
//
// 1. We can deploy the example
// 2. The deployment succeeds without errors
// 3. We can open a port forward to one of the Pods and access nginx
// 4. We can access nginx via the service endpoint
// 5. We can access nginx via the ingress endpoint
// 6. If we set a lower priority path, the application path takes precendence over the nginx service
// 7. If we set a higher priority path, that takes precedence over the nginx service
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
		SetValues: map[string]string{
			"ingress.enabled":     "true",
			"ingress.path":        "/app",
			"ingress.servicePort": "http",
			"ingress.annotations.kubernetes\\.io/ingress\\.class":                  "nginx",
			"ingress.annotations.nginx\\.ingress\\.kubernetes\\.io/rewrite-target": "/",
			"ingress.additionalPaths[0].path":                                      "/app",
			"ingress.additionalPaths[0].serviceName":                               "black-hole",
			"ingress.additionalPaths[0].servicePort":                               "80",
			"ingress.additionalPaths[1].path":                                      "/black-hole",
			"ingress.additionalPaths[1].serviceName":                               "black-hole",
			"ingress.additionalPaths[1].servicePort":                               "80",
		},
	}
	defer helm.Delete(t, options, releaseName, true)
	helm.Install(t, options, helmChartPath, releaseName)

	verifyPodsCreatedSuccessfully(t, kubectlOptions, "nginx", releaseName)
	verifyAllPodsAvailable(t, kubectlOptions, "nginx", releaseName, nginxValidationFunction)
	verifyServiceAvailable(t, kubectlOptions, "nginx", releaseName, nginxValidationFunction)

	// We expect this to succeed, because the black hole service that overlaps with the nginx service is added as lower
	// priority.
	verifyIngressAvailable(t, kubectlOptions, "nginx", releaseName, "/app", nginxValidationFunction, false)

	// On the other hand, we expect this to fail because the black hole service does not exist
	verifyIngressAvailable(t, kubectlOptions, "nginx", releaseName, "/black-hole", serviceAvailableValidationFunction, true)

	// Now redeploy with higher priority path and make sure it fails
	options.SetValues["ingress.additionalPathsHigherPriority[0].path"] = "/app"
	options.SetValues["ingress.additionalPathsHigherPriority[0].serviceName"] = "black-hole"
	options.SetValues["ingress.additionalPathsHigherPriority[0].servicePort"] = "80"
	helm.Upgrade(t, options, helmChartPath, releaseName)

	// We expect the service to still come up cleanly
	verifyPodsCreatedSuccessfully(t, kubectlOptions, "nginx", releaseName)
	verifyAllPodsAvailable(t, kubectlOptions, "nginx", releaseName, nginxValidationFunction)
	verifyServiceAvailable(t, kubectlOptions, "nginx", releaseName, nginxValidationFunction)

	// ... but now the nginx service via ingress should be unavailable because of the higher priority black hole path
	verifyIngressAvailable(t, kubectlOptions, "nginx", releaseName, "/app", serviceAvailableValidationFunction, true)
}

// nginxValidationFunction checks that we get a 200 response with the nginx welcome page.
func nginxValidationFunction(statusCode int, body string) bool {
	return serviceAvailableValidationFunction(statusCode, body) && strings.Contains(body, "Welcome to nginx")
}

// serviceAvailableValidationFunction checks that we get a 200 response
func serviceAvailableValidationFunction(statusCode int, body string) bool {
	return statusCode == 200
}

func verifyIngressAvailable(
	t *testing.T,
	kubectlOptions *k8s.KubectlOptions,
	appName string,
	releaseName string,
	path string,
	validationFunction func(int, string) bool,
	expectTimeout bool,
) {
	var retries int
	if expectTimeout {
		retries = ExpectTimeoutWaitTimerRetries
	} else {
		retries = WaitTimerRetries
	}

	// Get the service and wait until it is available
	filters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s", appName, releaseName),
	}
	ingresses := k8s.ListIngresses(t, kubectlOptions, filters)
	require.Equal(t, len(ingresses), 1)
	ingressName := ingresses[0].Name
	k8s.WaitUntilIngressAvailable(
		t,
		kubectlOptions,
		ingressName,
		retries,
		WaitTimerSleep,
	)

	// Now hit the service endpoint to verify it is accessible
	ingress := k8s.GetIngress(t, kubectlOptions, ingressName)
	ingressEndpoint := ingress.Status.LoadBalancer.Ingress[0].IP
	err := http_helper.HttpGetWithRetryWithCustomValidationE(
		t,
		fmt.Sprintf("http://%s%s", ingressEndpoint, path),
		retries,
		WaitTimerSleep,
		validationFunction,
	)
	if expectTimeout {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}
}
