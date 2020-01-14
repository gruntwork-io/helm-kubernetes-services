// +build all integration

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ghodss/yaml"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	WaitTimerRetries = 60
	WaitTimerSleep   = 5 * time.Second
	NumPodsExpected  = 3
)

// verifyPodsCreatedSuccessfully waits until the pods for the given helm release are created.
func verifyPodsCreatedSuccessfully(
	t *testing.T,
	kubectlOptions *k8s.KubectlOptions,
	appName string,
	releaseName string,
) {
	// Get the pods and wait until they are all ready
	filters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s", appName, releaseName),
	}

	k8s.WaitUntilNumPodsCreated(t, kubectlOptions, filters, NumPodsExpected, WaitTimerRetries, WaitTimerSleep)
	pods := k8s.ListPods(t, kubectlOptions, filters)

	for _, pod := range pods {
		k8s.WaitUntilPodAvailable(t, kubectlOptions, pod.Name, WaitTimerRetries, WaitTimerSleep)
	}
}

// verifyCanaryAndMainPodsCreatedSuccessfully uses gruntwork.io/deployment-type labels to ensure availability of both main and canary pods of a given release
func verifyCanaryAndMainPodsCreatedSuccessfully(
	t *testing.T,
	kubectlOptions *k8s.KubectlOptions,
	appName string,
	releaseName string,
) {

	filters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s,gruntwork.io/deployment-type=canary", appName, releaseName),
	}

	k8s.WaitUntilNumPodsCreated(t, kubectlOptions, filters, NumPodsExpected, WaitTimerRetries, WaitTimerSleep)
	pods := k8s.ListPods(t, kubectlOptions, filters)

	for _, pod := range pods {
		k8s.WaitUntilPodAvailable(t, kubectlOptions, pod.Name, WaitTimerRetries, WaitTimerSleep)
	}

	mainDeploymentFilters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s,gruntwork.io/deployment-type=main", appName, releaseName),
	}

	k8s.WaitUntilNumPodsCreated(t, kubectlOptions, mainDeploymentFilters, NumPodsExpected, WaitTimerRetries, WaitTimerSleep)
	mainPods := k8s.ListPods(t, kubectlOptions, mainDeploymentFilters)

	for _, mainPod := range mainPods {
		k8s.WaitUntilPodAvailable(t, kubectlOptions, mainPod.Name, WaitTimerRetries, WaitTimerSleep)
	}

}

// verifyDifferentContainerTagsForCanaryPods ensures that the pods that comprise the main deployment
// and the pods that comprise the canary deployment are running different image tags
func verifyDifferentContainerTagsForCanaryPods(
	t *testing.T,
	kubectlOptions *k8s.KubectlOptions,
	releaseName string,
) {
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

	assert.NotEqual(t, canaryTag, mainTag)
}

// verifyAllPodsAvailable waits until all the pods from the release are up and ready to serve traffic. The
// validationFunction is used to verify a successful response from the Pod.
func verifyAllPodsAvailable(
	t *testing.T,
	kubectlOptions *k8s.KubectlOptions,
	appName string,
	releaseName string,
	validationFunction func(int, string) bool,
) {
	filters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s", appName, releaseName),
	}
	pods := k8s.ListPods(t, kubectlOptions, filters)
	for _, pod := range pods {
		verifySinglePodAvailable(t, kubectlOptions, pod, validationFunction)
	}
}

// verifySinglePodAvailable waits until the given pod is ready to serve traffic. Does so by pinging port 80 on the Pod
// container. The validationFunction is used to verify a successful response from the Pod.
func verifySinglePodAvailable(
	t *testing.T,
	kubectlOptions *k8s.KubectlOptions,
	pod corev1.Pod,
	validationFunction func(int, string) bool,
) {
	// Open a tunnel from any available port locally
	localPort := k8s.GetAvailablePort(t)
	tunnel := k8s.NewTunnel(kubectlOptions, k8s.ResourceTypePod, pod.Name, localPort, 80)
	defer tunnel.Close()
	tunnel.ForwardPort(t)

	// Try to access the service on the local port, retrying until we get a good response for up to 5 minutes
	http_helper.HttpGetWithRetryWithCustomValidation(
		t,
		fmt.Sprintf("http://%s", tunnel.Endpoint()),
		nil,
		WaitTimerRetries,
		WaitTimerSleep,
		validationFunction,
	)
}

// verifyServiceAvailable waits until the service associated with the helm release is available and ready to serve
// traffic. The validationFunction is used to verify a successful response from the Pod.
func verifyServiceAvailable(
	t *testing.T,
	kubectlOptions *k8s.KubectlOptions,
	appName string,
	releaseName string,
	validationFunction func(int, string) bool,
) {
	// Get the service and wait until it is available
	filters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s", appName, releaseName),
	}
	services := k8s.ListServices(t, kubectlOptions, filters)
	require.Equal(t, len(services), 1)
	service := services[0]
	k8s.WaitUntilServiceAvailable(t, kubectlOptions, service.Name, WaitTimerRetries, WaitTimerSleep)

	// Now hit the service endpoint to verify it is accessible
	// Refresh service object in memory
	service = *k8s.GetService(t, kubectlOptions, service.Name)
	serviceEndpoint := k8s.GetServiceEndpoint(t, kubectlOptions, &service, 80)
	http_helper.HttpGetWithRetryWithCustomValidation(
		t,
		fmt.Sprintf("http://%s", serviceEndpoint),
		nil,
		WaitTimerRetries,
		WaitTimerSleep,
		validationFunction,
	)
}

// verifyServiceRoutesToMainAndCanaryPods ensures that the service is routing to both the main and the canary pods
// It does this by repeatedly issuing requests to the service and inspecting the nginx Server header
// Once both nginx tags have been seen in this header - we can be confident that we've reached both types of pod via the service
func verifyServiceRoutesToMainAndCanaryPods(
	t *testing.T,
	kubectlOptions *k8s.KubectlOptions,
	appName string,
	releaseName string,
) {

	// Get the service and wait until it is available
	filters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s", appName, releaseName),
	}
	services := k8s.ListServices(t, kubectlOptions, filters)
	require.Equal(t, len(services), 1)
	service := services[0]
	k8s.WaitUntilServiceAvailable(t, kubectlOptions, service.Name, WaitTimerRetries, WaitTimerSleep)

	service = *k8s.GetService(t, kubectlOptions, service.Name)
	serviceEndpoint := k8s.GetServiceEndpoint(t, kubectlOptions, &service, 80)

	// Ensure that the service routes to both the main and canary deployment pods
	// Read the latest values dynamically in case the fixtures file changes
	valuesFile, err := ioutil.ReadFile(filepath.Join("fixtures", "canary_and_main_deployment_values.yaml"))
	assert.NoError(t, err)

	rendered := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(valuesFile), &rendered)

	mainImageTag := rendered["containerImage"].(map[string]interface{})["tag"].(string)

	canaryImageTag := rendered["canary"].(map[string]interface{})["containerImage"].(map[string]interface{})["tag"].(string)

	// We haven't seen either tag come back in the nginx Server header yet
	seen := make(map[string]bool)
	seen[mainImageTag] = false
	seen[canaryImageTag] = false

	for seen[mainImageTag] == false || seen[canaryImageTag] == false {
		resp, err := http.Get(fmt.Sprintf("http://%s", serviceEndpoint))

		assert.NoError(t, err)

		serverNginxHeader := resp.Header.Get("Server")
		// Nginx returns a server header in the format "nginx/1.16.0"
		serverTag := strings.ReplaceAll(serverNginxHeader, "nginx/", "")

		// When we see a header value, update it as seen
		seen[serverTag] = true

		time.Sleep(1 * time.Second)
	}
}
