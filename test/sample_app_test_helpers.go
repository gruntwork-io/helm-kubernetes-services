// +build all integration

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/shell"
)

// Expected response from the sample app is a json
type SampleAppResponse struct {
	Text string `json:"text"`
}

// createSampleAppDockerImage builds the sample app docker image into the minikube environment, tagging it using the
// unique ID.
func createSampleAppDockerImage(t *testing.T, uniqueID string, examplePath string) {
	dockerWorkingDir := filepath.Join(examplePath, "docker")
	cmdsToRun := []string{
		// Build the docker environment to talk to minikube daemon
		"eval $(minikube docker-env)",
		// Build the image and tag using the unique ID
		fmt.Sprintf("docker build -t gruntwork-io/sample-sinatra-app:%s .", uniqueID),
	}
	cmd := shell.Command{
		Command: "sh",
		Args: []string{
			"-c",
			strings.Join(cmdsToRun, " && "),
		},
		WorkingDir: dockerWorkingDir,
	}
	shell.RunCommand(t, cmd)
}

// sampleAppValidationFunctionGenerator will output a validation function that can be used with the pod verification
// code in k8s_service_example_test_helpers.go.
func sampleAppValidationFunctionGenerator(t *testing.T, expectedText string) func(int, string) bool {
	return func(statusCode int, body string) bool {
		if statusCode != 200 {
			return false
		}

		var resp SampleAppResponse
		err := json.Unmarshal([]byte(body), &resp)
		if err != nil {
			logger.Logf(t, "Error unmarshalling sample app response: %s", err)
			return false
		}
		return resp.Text == expectedText
	}
}
