// +build all integration

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	verifyPodsCreatedSuccessfully(t, kubectlOptions, releaseName)
	verifyAllPodNginxAvailable(t, kubectlOptions, releaseName)
	verifyServiceAvailable(t, kubectlOptions, releaseName)
}

func verifyPodsCreatedSuccessfully(t *testing.T, kubectlOptions *k8s.KubectlOptions, releaseName string) {
	// Get the pods and wait until they are all ready
	filters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=nginx,app.kubernetes.io/instance=%s", releaseName),
	}
	k8s.WaitUntilNumPodsCreated(t, kubectlOptions, filters, 3, 60, 1*time.Second)
	pods := k8s.ListPods(t, kubectlOptions, filters)
	for _, pod := range pods {
		k8s.WaitUntilPodAvailable(t, kubectlOptions, pod.Name, 60, 1*time.Second)
	}
}

func verifyAllPodNginxAvailable(t *testing.T, kubectlOptions *k8s.KubectlOptions, releaseName string) {
	filters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=nginx,app.kubernetes.io/instance=%s", releaseName),
	}
	pods := k8s.ListPods(t, kubectlOptions, filters)
	for _, pod := range pods {
		verifySinglePodNginxAvailable(t, kubectlOptions, pod)
	}
}

func verifySinglePodNginxAvailable(t *testing.T, kubectlOptions *k8s.KubectlOptions, pod corev1.Pod) {
	// Open a tunnel from any available port locally
	localPort := k8s.GetAvailablePort(t)
	tunnel := k8s.NewTunnel(kubectlOptions, k8s.ResourceTypePod, pod.Name, localPort, 80)
	defer tunnel.Close()
	tunnel.ForwardPort(t)

	// Try to access the nginx service on the local port, retrying until we get a good response for up to 5 minutes
	http_helper.HttpGetWithRetryWithCustomValidation(
		t,
		fmt.Sprintf("http://localhost:%d", localPort),
		60,
		5*time.Second,
		func(statusCode int, body string) bool {
			return statusCode == 200
		},
	)
}

func verifyServiceAvailable(t *testing.T, kubectlOptions *k8s.KubectlOptions, releaseName string) {
	// Get the service and wait until it is available
	filters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=nginx,app.kubernetes.io/instance=%s", releaseName),
	}
	services := k8s.ListServices(t, kubectlOptions, filters)
	require.Equal(t, len(services), 1)
	service := services[0]
	k8s.WaitUntilServiceAvailable(t, kubectlOptions, service.Name, 60, 5*time.Second)

	// Now hit the service endpoint to verify it is accessible
	// Refresh service object in memory
	service := k8s.GetService(t, kubectlOptions, service.Name)
	serviceEndpoint := k8s.GetServiceEndpoint(t, service, 80)
	http_helper.HttpGetWithRetryWithCustomValidation(
		t,
		serviceEndpoint,
		60,
		5*time.Second,
		func(statusCode int, body string) bool {
			return statusCode == 200
		},
	)
}
