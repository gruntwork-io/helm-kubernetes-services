//go:build all || tpl
// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test each of the required values. Here, we take advantage of the fact that linter_values.yaml is supposed to define
// all the required values, so we check the template rendering by nulling out each field.
func TestK8SJobRequiredValuesAreRequired(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-job"))
	require.NoError(t, err)

	eachRequired := []string{
		"containerImage.repository",
		"containerImage.tag",
		"applicationName",
	}
	for _, requiredVal := range eachRequired {
		// Capture the range value and force it into this scope. Otherwise, it is defined outside this block so it can
		// change when the subtests parallelize and switch contexts.
		requiredVal := requiredVal
		t.Run(requiredVal, func(t *testing.T) {
			t.Parallel()

			// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
			// We then use SetValues to null out the value.
			options := &helm.Options{
				ValuesFiles: []string{filepath.Join("..", "charts", "k8s-job", "linter_values.yaml")},
				SetValues:   map[string]string{requiredVal: "null"},
			}
			_, err := helm.RenderTemplateE(t, options, helmChartPath, strings.ToLower(t.Name()), []string{})
			assert.Error(t, err)
		})
	}
}

// Test each of the optional values defined in linter_values.yaml. Here, we take advantage of the fact that
// linter_values.yaml is supposed to define all the required values, so we check the template rendering by nulling out
// each field.
func TestK8SJobOptionalValuesAreOptional(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-job"))
	require.NoError(t, err)

	eachOptional := []string{
		"containerImage.pullPolicy",
	}
	for _, optionalVal := range eachOptional {
		// Capture the range value and force it into this scope. Otherwise, it is defined outside this block so it can
		// change when the subtests parallelize and switch contexts.
		optionalVal := optionalVal
		t.Run(optionalVal, func(t *testing.T) {
			t.Parallel()

			// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
			// We then use SetValues to null out the value.
			options := &helm.Options{
				ValuesFiles: []string{filepath.Join("..", "charts", "k8s-job", "linter_values.yaml")},
				SetValues:   map[string]string{optionalVal: "null"},
			}
			// Make sure it renders without error
			helm.RenderTemplate(t, options, helmChartPath, "all", []string{})
		})
	}
}

// Test that annotations render correctly to annotate the Job resource
func TestK8SJobAnnotationsRenderCorrectly(t *testing.T) {
	t.Parallel()

	uniqueID := random.UniqueId()
	// ERROR: Need to find function that can inject annotations into a job
	job := renderK8SServiceDeploymentWithSetValues(t, map[string]string{"jobAnnotations.unique-id": uniqueID})

	assert.Equal(t, len(job.Annotations), 1)
	assert.Equal(t, job.Annotations["unique-id"], uniqueID)
}

func TestK8SJobSecurityContextAnnotationRenderCorrectly(t *testing.T) {
	t.Parallel()
	job := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"securityContext.privileged": "true",
			"securityContext.runAsUser":  "1000",
		},
	)
	renderedContainers := job.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedContainers), 1)
	testContainer := renderedContainers[0]
	assert.NotNil(t, testContainer.SecurityContext)
	assert.True(t, *testContainer.SecurityContext.Privileged)
	assert.Equal(t, *testContainer.SecurityContext.RunAsUser, int64(1000))
}

// Test that default imagePullSecrets do not render any
func TestK8SJobNoImagePullSecrets(t *testing.T) {
	t.Parallel()

	job := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{},
	)

	renderedImagePullSecrets := job.Spec.Template.Spec.ImagePullSecrets
	require.Equal(t, len(renderedImagePullSecrets), 0)
}

func TestK8SJobMultipleImagePullSecrets(t *testing.T) {
	t.Parallel()

	job := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"imagePullSecrets[0]": "docker-private-registry-key",
			"imagePullSecrets[1]": "gcr-registry-key",
		},
	)

	renderedImagePullSecrets := job.Spec.Template.Spec.ImagePullSecrets
	require.Equal(t, len(renderedImagePullSecrets), 2)
	assert.Equal(t, renderedImagePullSecrets[0].Name, "docker-private-registry-key")
	assert.Equal(t, renderedImagePullSecrets[1].Name, "gcr-registry-key")
}

// Test that omitting containerCommand does not set command attribute on the Job container spec.
func TestK8SJobDefaultHasNullCommandSpec(t *testing.T) {
	t.Parallel()

	job := renderK8SServiceDeploymentWithSetValues(t, map[string]string{})
	renderedContainers := job.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedContainers), 1)
	appContainer := renderedContainers[0]
	assert.Nil(t, appContainer.Command)
}

// Test that setting containerCommand sets the command attribute on the Job container spec.
func TestK8SJobWithContainerCommandHasCommandSpec(t *testing.T) {
	t.Parallel()

	job := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"containerCommand[0]": "echo",
			"containerCommand[1]": "Hello world",
		},
	)
	renderedContainers := job.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedContainers), 1)
	appContainer := renderedContainers[0]
	assert.Equal(t, appContainer.Command, []string{"echo", "Hello world"})
}

func TestK8SJobMainJobContainersLabeledCorrectly(t *testing.T) {
	t.Parallel()
	job := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"containerImage.repository": "nginx",
			"containerImage.tag":        "1.16.0",
		},
	)
	// Ensure a "main" type job is properly labeled as such
	assert.Equal(t, job.Spec.Selector.MatchLabels["gruntwork.io/job-type"], "main")
}

func TestK8SJobAddingAdditionalLabels(t *testing.T) {
	t.Parallel()
	first_custom_job_label_value := "first-custom-value"
	second_custom_job_label_value := "second-custom-value"
	job := renderK8SServiceDeploymentWithSetValues(t,
		map[string]string{"additionalJobLabels.first-label": first_custom_job_label_value,
			"additionalJobLabels.second-label": second_custom_job_label_value})

	assert.Equal(t, job.Labels["first-label"], first_custom_job_label_value)
	assert.Equal(t, job.Labels["second-label"], second_custom_job_label_value)
}

func TestK8SJobFullnameOverride(t *testing.T) {
	t.Parallel()

	overiddenName := "overidden-name"

	job := renderK8SServiceDeploymentWithSetValues(t,
		map[string]string{
			"fullnameOverride": overiddenName,
		},
	)

	assert.Equal(t, job.Name, overiddenName)
}

func TestK8SJobEnvFrom(t *testing.T) {
	t.Parallel()

	t.Run("BothConfigMapsAndSecretsEnvFrom", func(t *testing.T) {
		job := renderK8SServiceDeploymentWithSetValues(t,
			map[string]string{
				"configMaps.test-configmap.as": "envFrom",
				"secrets.test-secret.as":       "envFrom",
			},
		)

		assert.NotNil(t, job.Spec.Template.Spec.Containers[0].EnvFrom)
		assert.Equal(t, len(job.Spec.Template.Spec.Containers[0].EnvFrom), 2)
		assert.Equal(t, job.Spec.Template.Spec.Containers[0].EnvFrom[0].ConfigMapRef.Name, "test-configmap")
		assert.Equal(t, job.Spec.Template.Spec.Containers[0].EnvFrom[1].SecretRef.Name, "test-secret")
	})

	t.Run("OnlyConfigMapsEnvFrom", func(t *testing.T) {
		job := renderK8SServiceDeploymentWithSetValues(t,
			map[string]string{
				"configMaps.test-configmap.as": "envFrom",
			},
		)

		assert.NotNil(t, job.Spec.Template.Spec.Containers[0].EnvFrom)
		assert.Equal(t, len(job.Spec.Template.Spec.Containers[0].EnvFrom), 1)
		assert.Equal(t, job.Spec.Template.Spec.Containers[0].EnvFrom[0].ConfigMapRef.Name, "test-configmap")
	})

	t.Run("OnlySecretsEnvFrom", func(t *testing.T) {
		job := renderK8SServiceDeploymentWithSetValues(t,
			map[string]string{
				"secrets.test-secret.as": "envFrom",
			},
		)

		assert.NotNil(t, job.Spec.Template.Spec.Containers[0].EnvFrom)
		assert.Equal(t, len(job.Spec.Template.Spec.Containers[0].EnvFrom), 1)
		assert.Equal(t, job.Spec.Template.Spec.Containers[0].EnvFrom[0].SecretRef.Name, "test-secret")
	})

}
