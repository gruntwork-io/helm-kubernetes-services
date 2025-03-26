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
	"golang.org/x/mod/semver"
)

// Test that:
//
// 1. We can deploy the example Job
// 2. The Job succeeds without errors

func TestK8SJobBusyboxExample(t *testing.T) {
	t.Parallel()

	workingDir := filepath.Join(".", "stages", t.Name())

	//os.Setenv("SKIP_setup", "true")
	//os.Setenv("SKIP_create_namespace", "true")
	//os.Setenv("SKIP_install", "true")
	//os.Setenv("SKIP_validate_job_deployment", "true")
	//os.Setenv("SKIP_upgrade", "true")
	//os.Setenv("SKIP_validate_upgrade", "true")
	//os.Setenv("SKIP_delete", "true")
	//os.Setenv("SKIP_delete_namespace", "true")

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-job"))
	require.NoError(t, err)
	examplePath, err := filepath.Abs(filepath.Join("..", "examples", "k8s-job-busybox"))
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
	testNamespace := fmt.Sprintf("k8s-job-busybox-%s", strings.ToLower(uniqueID))

	defer test_structure.RunTestStage(t, "delete_namespace", func() {
		k8s.DeleteNamespace(t, kubectlOptions, testNamespace)
	})

	test_structure.RunTestStage(t, "create_namespace", func() {
		k8s.CreateNamespace(t, kubectlOptions, testNamespace)
	})

	kubectlOptions.Namespace = testNamespace

	// Use the values file in the example and deploy the chart in the test namespace
	// Set a random release name
	releaseName := fmt.Sprintf("k8s-job-busybox-%s", strings.ToLower(uniqueID))
	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		ValuesFiles:    []string{filepath.Join(examplePath, "values.yaml")},
	}

	defer test_structure.RunTestStage(t, "delete", func() {
		helm.Delete(t, options, releaseName, true)
	})

	test_structure.RunTestStage(t, "install", func() {
		helm.Install(t, options, helmChartPath, releaseName)
	})

	test_structure.RunTestStage(t, "validate_job_deployment", func() {
		verifyPodsCreatedSuccessfully(t, kubectlOptions, "busybox", releaseName, NumPodsExpected)

	})
}
