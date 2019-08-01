// +build all integration

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"

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
		&tls.Config{},
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
		&tls.Config{},
		WaitTimerRetries,
		WaitTimerSleep,
		validationFunction,
	)
}
