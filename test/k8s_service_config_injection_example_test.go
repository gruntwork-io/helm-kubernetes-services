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

// Test the base case of the k8s-service-config-injection example, where the server port is set using hard coded
// environment variables. This test will check that:
//
// 1. The docker container can be built
// 2. The base values.yaml file can be used to deploy the docker container
// 3. The deployed container responds to web requests with the default server text.
func TestK8SServiceConfigInjectionBaseExample(t *testing.T) {
	t.Parallel()

	// Setup paths for testing the example chart
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)
	examplePath, err := filepath.Abs(filepath.Join("..", "examples", "k8s-service-config-injection"))
	require.NoError(t, err)

	// Create a test namespace to deploy resources into, to avoid colliding with other tests
	kubectlOptions := k8s.NewKubectlOptions("", "", "")
	uniqueID := random.UniqueId()
	testNamespace := fmt.Sprintf("k8s-service-config-injection-%s", strings.ToLower(uniqueID))
	k8s.CreateNamespace(t, kubectlOptions, testNamespace)
	defer k8s.DeleteNamespace(t, kubectlOptions, testNamespace)
	kubectlOptions.Namespace = testNamespace

	// Build the docker image
	createSampleAppDockerImage(t, uniqueID, examplePath)

	// Install the base chart
	// Set a random release name here so we can track it later
	releaseName := fmt.Sprintf("k8s-service-config-injection-%s", strings.ToLower(uniqueID))
	// Use the values file in the example and deploy the chart in the test namespace
	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		ValuesFiles:    []string{filepath.Join(examplePath, "values.yaml")},
		// Override the image tag
		SetValues: map[string]string{"containerImage.tag": uniqueID},
	}
	defer helm.Delete(t, options, releaseName, true)
	helm.Install(t, options, helmChartPath, releaseName)

	// Verify the app comes up cleanly and returns the expected text
	expectedText := "Hello from backend"
	validationFunction := sampleAppValidationFunctionGenerator(t, expectedText)
	verifyPodsCreatedSuccessfully(t, kubectlOptions, "sample-sinatra-app", releaseName, NumPodsExpected)
	verifyAllPodsAvailable(t, kubectlOptions, "sample-sinatra-app", releaseName, validationFunction)
	verifyServiceAvailable(t, kubectlOptions, "sample-sinatra-app", releaseName, validationFunction)

}

// Test the ConfigMap case of the k8s-service-config-injection example, where the server text is derived from a
// ConfigMap that is injected as an environment variable. This test will check that:
//
// 1. The docker container can be built
// 2. The provided kubernetes resource file can be used to create a ConfigMap containing a modified server text.
// 3. The base values.yaml file can be combined with the extension for pulling in the server text from a config map.
// 4. The deployed docker container responds with the server text derived from the config map
func TestK8SServiceConfigInjectionConfigMapExample(t *testing.T) {
	t.Parallel()

	// Setup paths for testing the example chart
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)
	examplePath, err := filepath.Abs(filepath.Join("..", "examples", "k8s-service-config-injection"))
	require.NoError(t, err)

	// Create a test namespace to deploy resources into, to avoid colliding with other tests
	kubectlOptions := k8s.NewKubectlOptions("", "", "")
	uniqueID := random.UniqueId()
	testNamespace := fmt.Sprintf("k8s-service-config-injection-%s", strings.ToLower(uniqueID))
	k8s.CreateNamespace(t, kubectlOptions, testNamespace)
	defer k8s.DeleteNamespace(t, kubectlOptions, testNamespace)
	kubectlOptions.Namespace = testNamespace

	// Build the docker image
	createSampleAppDockerImage(t, uniqueID, examplePath)

	// Install the configmap
	kubeResourceConfigPath := filepath.Join(examplePath, "kubernetes", "config-map.yaml")
	defer k8s.KubectlDelete(t, kubectlOptions, kubeResourceConfigPath)
	k8s.KubectlApply(t, kubectlOptions, kubeResourceConfigPath)

	// Install the chart with the base values and config map values
	// Set a random release name here so we can track it later
	releaseName := fmt.Sprintf("k8s-service-config-injection-%s", strings.ToLower(uniqueID))
	// Use the values file in the example, override it with the configmap extension and deploy the chart in the test
	// namespace
	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		ValuesFiles: []string{
			// Base example values
			filepath.Join(examplePath, "values.yaml"),
			// Example config map extensions values
			filepath.Join(examplePath, "extensions", "config_map_values.yaml"),
		},
		// Override the image tag
		SetValues: map[string]string{"containerImage.tag": uniqueID},
	}
	defer helm.Delete(t, options, releaseName, true)
	helm.Install(t, options, helmChartPath, releaseName)

	// Verify the app comes up cleanly and returns the expected text
	expectedText := "Hello! I was configured using a ConfigMap!"
	validationFunction := sampleAppValidationFunctionGenerator(t, expectedText)
	verifyPodsCreatedSuccessfully(t, kubectlOptions, "sample-sinatra-app", releaseName, NumPodsExpected)
	verifyAllPodsAvailable(t, kubectlOptions, "sample-sinatra-app", releaseName, validationFunction)
	verifyServiceAvailable(t, kubectlOptions, "sample-sinatra-app", releaseName, validationFunction)
}

// Test the Secret case of the k8s-service-config-injection example, where the server text is derived from a
// Secret that is injected as an environment variable. This test will check that:
//
// 1. The docker container can be built
// 2. Create a Secret used to inject the server text.
// 3. The base values.yaml file can be combined with the extension for pulling in the server text from a secret.
// 4. The deployed docker container responds with the server text derived from the secret
func TestK8SServiceConfigInjectionSecretExample(t *testing.T) {
	t.Parallel()

	// Setup paths for testing the example chart
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)
	examplePath, err := filepath.Abs(filepath.Join("..", "examples", "k8s-service-config-injection"))
	require.NoError(t, err)

	// Create a test namespace to deploy resources into, to avoid colliding with other tests
	kubectlOptions := k8s.NewKubectlOptions("", "", "")
	uniqueID := random.UniqueId()
	testNamespace := fmt.Sprintf("k8s-service-config-injection-%s", strings.ToLower(uniqueID))
	k8s.CreateNamespace(t, kubectlOptions, testNamespace)
	defer k8s.DeleteNamespace(t, kubectlOptions, testNamespace)
	kubectlOptions.Namespace = testNamespace

	// Build the docker image
	createSampleAppDockerImage(t, uniqueID, examplePath)

	// Create a secret using kubectl
	// Make sure to delete the secret in the undeploy process
	defer k8s.RunKubectl(t, kubectlOptions, "delete", "secret", "sample-sinatra-app-server-text")
	// Create Secret from a string literal
	// kubectl create secret generic sample-sinatra-app-server-text --from-literal server_text='Hello! I was configured using a Secret!'
	k8s.RunKubectl(
		t,
		kubectlOptions,
		"create",
		"secret",
		"generic",
		"sample-sinatra-app-server-text",
		"--from-literal",
		"server_text=Hello! I was configured using a Secret!",
	)

	// Install the chart with the base values and config map values
	// Set a random release name here so we can track it later
	releaseName := fmt.Sprintf("k8s-service-config-injection-%s", strings.ToLower(uniqueID))
	// Use the values file in the example, override it with the secret extension and deploy the chart in the test
	// namespace
	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		ValuesFiles: []string{
			// Base example values
			filepath.Join(examplePath, "values.yaml"),
			// Example config map extensions values
			filepath.Join(examplePath, "extensions", "secret_values.yaml"),
		},
		// Override the image tag
		SetValues: map[string]string{"containerImage.tag": uniqueID},
	}
	defer helm.Delete(t, options, releaseName, true)
	helm.Install(t, options, helmChartPath, releaseName)

	// Verify the app comes up cleanly and returns the expected text
	expectedText := "Hello! I was configured using a Secret!"
	validationFunction := sampleAppValidationFunctionGenerator(t, expectedText)
	verifyPodsCreatedSuccessfully(t, kubectlOptions, "sample-sinatra-app", releaseName, NumPodsExpected)
	verifyAllPodsAvailable(t, kubectlOptions, "sample-sinatra-app", releaseName, validationFunction)
	verifyServiceAvailable(t, kubectlOptions, "sample-sinatra-app", releaseName, validationFunction)
}
