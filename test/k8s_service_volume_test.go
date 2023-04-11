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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestK8SServiceScratchSpaceIsTmpfs(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// Create a test namespace to deploy resources into, to avoid colliding with other tests
	kubectlOptions := k8s.NewKubectlOptions("", "", "")
	uniqueID := random.UniqueId()
	testNamespace := fmt.Sprintf("k8s-service-scratch-%s", strings.ToLower(uniqueID))
	k8s.CreateNamespace(t, kubectlOptions, testNamespace)
	defer k8s.DeleteNamespace(t, kubectlOptions, testNamespace)
	kubectlOptions.Namespace = testNamespace

	// Construct the values to run a pod with scratch space
	releaseName := fmt.Sprintf("k8s-service-scratch-%s", strings.ToLower(uniqueID))
	appName := "scratch-tester"
	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		SetValues: map[string]string{
			"applicationName":           appName,
			"containerImage.repository": "alpine",
			"containerImage.tag":        "3.13",
			"containerImage.pullPolicy": "IfNotPresent",
			"containerCommand[0]":       "sh",
			"containerCommand[1]":       "-c",
			"containerCommand[2]":       "mount && sleep 9999999",
			"scratchPaths.scratch-mnt":  "/mnt/scratch",
		},
	}
	defer helm.Delete(t, options, releaseName, true)
	helm.Install(t, options, helmChartPath, releaseName)

	// Make sure all the pods are deployed and available
	verifyPodsCreatedSuccessfully(t, kubectlOptions, appName, releaseName, 1)

	// Get the logs from the pod to verify /mnt/scratch is mounted as tmpfs.
	pods := k8s.ListPods(t, kubectlOptions, metav1.ListOptions{})
	require.Equal(t, 1, len(pods))
	pod := pods[0]
	logs, err := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, "logs", pod.Name)
	require.NoError(t, err)
	require.Contains(t, logs, "tmpfs on /mnt/scratch type tmpfs (rw,relatime")
}
