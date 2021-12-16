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
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/require"
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

	workingDir := filepath.Join(".", "stages", t.Name())

	//os.Setenv("SKIP_setup", "true")
	//os.Setenv("SKIP_create_namespace", "true")
	//os.Setenv("SKIP_install", "true")
	//os.Setenv("SKIP_validate_initial_deployment", "true")
	//os.Setenv("SKIP_upgrade", "true")
	//os.Setenv("SKIP_validate_upgrade", "true")
	//os.Setenv("SKIP_delete", "true")
	//os.Setenv("SKIP_delete_namespace", "true")

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)
	examplePath, err := filepath.Abs(filepath.Join("..", "examples", "k8s-service-nginx"))
	require.NoError(t, err)

	// Create a test namespace to deploy resources into, to avoid colliding with other tests
	test_structure.RunTestStage(t, "setup", func() {
		kubectlOptions := k8s.NewKubectlOptions("", "", "")
		test_structure.SaveKubectlOptions(t, workingDir, kubectlOptions)

		uniqueID := random.UniqueId()
		test_structure.SaveString(t, workingDir, "uniqueID", uniqueID)
	})
	kubectlOptions := test_structure.LoadKubectlOptions(t, workingDir)
	uniqueID := test_structure.LoadString(t, workingDir, "uniqueID")
	testNamespace := fmt.Sprintf("k8s-service-nginx-%s", strings.ToLower(uniqueID))

	defer test_structure.RunTestStage(t, "delete_namespace", func() {
		k8s.DeleteNamespace(t, kubectlOptions, testNamespace)
	})

	test_structure.RunTestStage(t, "create_namespace", func() {
		k8s.CreateNamespace(t, kubectlOptions, testNamespace)
	})

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
			"ingress.pathType":    "Prefix",
			"ingress.servicePort": "http",
			"ingress.annotations.kubernetes\\.io/ingress\\.class":                  "nginx",
			"ingress.annotations.nginx\\.ingress\\.kubernetes\\.io/rewrite-target": "/",
			"ingress.additionalPaths[0].path":                                      "/app",
			"ingress.additionalPaths[0].pathType":                                  "Prefix",
			"ingress.additionalPaths[0].serviceName":                               "black-hole",
			"ingress.additionalPaths[0].servicePort":                               "80",
			"ingress.additionalPaths[1].path":                                      "/black-hole",
			"ingress.additionalPaths[1].pathType":                                  "Prefix",
			"ingress.additionalPaths[1].serviceName":                               "black-hole",
			"ingress.additionalPaths[1].servicePort":                               "80",
		},
	}

	defer test_structure.RunTestStage(t, "delete", func() {
		helm.Delete(t, options, releaseName, true)
	})

	test_structure.RunTestStage(t, "install", func() {
		helm.Install(t, options, helmChartPath, releaseName)
	})

	test_structure.RunTestStage(t, "validate_initial_deployment", func() {
		verifyPodsCreatedSuccessfully(t, kubectlOptions, "nginx", releaseName, NumPodsExpected)
		verifyAllPodsAvailable(t, kubectlOptions, "nginx", releaseName, nginxValidationFunction)
		verifyServiceAvailable(t, kubectlOptions, "nginx", releaseName, nginxValidationFunction)

		// We expect this to succeed, because the black hole service that overlaps with the nginx service is added as lower
		// priority.
		verifyIngressAvailable(t, kubectlOptions, releaseName, "/app", nginxValidationFunction)

		// On the other hand, we expect this to fail because the black hole service does not exist
		verifyIngressAvailable(t, kubectlOptions, releaseName, "/black-hole", serviceUnavailableValidationFunction)
	})

	test_structure.RunTestStage(t, "upgrade", func() {
		// Now redeploy with higher priority path and make sure it fails
		options.SetValues["ingress.additionalPathsHigherPriority[0].path"] = "/app"
		options.SetValues["ingress.additionalPathsHigherPriority[0].serviceName"] = "black-hole"
		options.SetValues["ingress.additionalPathsHigherPriority[0].servicePort"] = "80"
		helm.Upgrade(t, options, helmChartPath, releaseName)
	})

	test_structure.RunTestStage(t, "validate_upgrade", func() {
		// We expect the service to still come up cleanly
		verifyPodsCreatedSuccessfully(t, kubectlOptions, "nginx", releaseName, NumPodsExpected)
		verifyAllPodsAvailable(t, kubectlOptions, "nginx", releaseName, nginxValidationFunction)
		verifyServiceAvailable(t, kubectlOptions, "nginx", releaseName, nginxValidationFunction)

		// ... but now the nginx service via ingress should be unavailable because of the higher priority black hole path
		verifyIngressAvailable(t, kubectlOptions, releaseName, "/app", serviceUnavailableValidationFunction)
	})
}

// nginxValidationFunction checks that we get a 200 response with the nginx welcome page.
func nginxValidationFunction(statusCode int, body string) bool {
	return statusCode == 200 && strings.Contains(body, "Welcome to nginx")
}

// serviceUnavailableValidationFunction checks that we get a 503 response and the maintenance page
func serviceUnavailableValidationFunction(statusCode int, body string) bool {
	return statusCode == 503 && strings.Contains(body, "Service Temporarily Unavailable")
}

func verifyIngressAvailable(
	t *testing.T,
	kubectlOptions *k8s.KubectlOptions,
	ingressName string,
	path string,
	validationFunction func(int, string) bool,
) {
	// Get the ingress and wait until it is available
	k8s.WaitUntilIngressAvailable(
		t,
		kubectlOptions,
		ingressName,
		WaitTimerRetries,
		WaitTimerSleep,
	)

	// Now hit the service endpoint to verify it is accessible
	ingress := k8s.GetIngress(t, kubectlOptions, ingressName)
	ingressEndpoint := ingress.Status.LoadBalancer.Ingress[0].IP
	http_helper.HttpGetWithRetryWithCustomValidation(
		t,
		fmt.Sprintf("http://%s%s", ingressEndpoint, path),
		nil,
		WaitTimerRetries,
		WaitTimerSleep,
		validationFunction,
	)
}
