//go:build all || tpl
// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that setting shutdownDelay to 0 will disable the preStop hook
func TestK8SServiceShutdownDelayZeroDisablesPreStopHook(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(t, map[string]string{"shutdownDelay": "0"})

	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	assert.Nil(t, appContainer.Lifecycle)
}

// Test that setting shutdownDelay to something greater than 0 will include a preStop hook
func TestK8SServiceNonZeroShutdownDelayIncludesPreStopHook(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(t, map[string]string{"shutdownDelay": "5"})

	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	require.NotNil(t, appContainer.Lifecycle)
	require.NotNil(t, appContainer.Lifecycle.PreStop)
	require.NotNil(t, appContainer.Lifecycle.PreStop.Exec)
	require.Equal(t, appContainer.Lifecycle.PreStop.Exec.Command, []string{"sleep", "5"})
}

func TestK8SServiceDeploymentAddingOnlyPostStartLifecycleHooks(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			// Disable shutdown delay to ensure it doesn't enable preStop hooks.
			"shutdownDelay": "0",

			"lifecycleHooks.enabled":                  "true",
			"lifecycleHooks.postStart.exec.command[0]": "echo",
			"lifecycleHooks.postStart.exec.command[1]": "run after start",
		},
	)

	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	require.NotNil(t, appContainer.Lifecycle)

	assert.Nil(t, appContainer.Lifecycle.PreStop)

	require.NotNil(t, appContainer.Lifecycle.PostStart)
	require.NotNil(t, appContainer.Lifecycle.PostStart.Exec)
	require.Equal(t, appContainer.Lifecycle.PostStart.Exec.Command, []string{"echo", "run after start"})
}

func TestK8SServiceDeploymentAddingOnlyPreStopLifecycleHooks(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			// Disable shutdown delay to ensure it doesn't enable preStop hooks.
			"shutdownDelay": "0",

			"lifecycleHooks.enabled":                 "true",
			"lifecycleHooks.preStop.exec.command[0]": "echo",
			"lifecycleHooks.preStop.exec.command[1]": "run before stop",
		},
	)

	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	require.NotNil(t, appContainer.Lifecycle)

	assert.Nil(t, appContainer.Lifecycle.PostStart)

	require.NotNil(t, appContainer.Lifecycle.PreStop)
	require.NotNil(t, appContainer.Lifecycle.PreStop.Exec)
	require.Equal(t, appContainer.Lifecycle.PreStop.Exec.Command, []string{"echo", "run before stop"})
}

func TestK8SServiceDeploymentAddingBothLifecycleHooks(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			// Disable shutdown delay to ensure it doesn't enable preStop hooks.
			"shutdownDelay": "0",

			"lifecycleHooks.enabled":                   "true",
			"lifecycleHooks.postStart.exec.command[0]": "echo",
			"lifecycleHooks.postStart.exec.command[1]": "run after start",
			"lifecycleHooks.preStop.exec.command[0]":   "echo",
			"lifecycleHooks.preStop.exec.command[1]":   "run before stop",
		},
	)

	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	require.NotNil(t, appContainer.Lifecycle)

	require.NotNil(t, appContainer.Lifecycle.PostStart)
	require.NotNil(t, appContainer.Lifecycle.PostStart.Exec)
	require.Equal(t, appContainer.Lifecycle.PostStart.Exec.Command, []string{"echo", "run after start"})

	require.NotNil(t, appContainer.Lifecycle.PreStop)
	require.NotNil(t, appContainer.Lifecycle.PreStop.Exec)
	require.Equal(t, appContainer.Lifecycle.PreStop.Exec.Command, []string{"echo", "run before stop"})
}

func TestK8SServiceDeploymentPreferExplicitPreStopOverShutdownDelay(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"shutdownDelay":                          "5",
			"lifecycleHooks.enabled":                 "true",
			"lifecycleHooks.preStop.exec.command[0]": "echo",
			"lifecycleHooks.preStop.exec.command[1]": "run before stop",
		},
	)

	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	require.NotNil(t, appContainer.Lifecycle)

	assert.Nil(t, appContainer.Lifecycle.PostStart)

	require.NotNil(t, appContainer.Lifecycle.PreStop)
	require.NotNil(t, appContainer.Lifecycle.PreStop.Exec)
	require.Equal(t, appContainer.Lifecycle.PreStop.Exec.Command, []string{"echo", "run before stop"})
}

func TestK8SServiceDeploymentEnabledFalseDisablesLifecycleHooksEvenWhenAddingBoth(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			// Disable shutdown delay to ensure it doesn't enable preStop hooks.
			"shutdownDelay": "0",

			"lifecycleHooks.enabled":                   "false",
			"lifecycleHooks.postStart.exec.command[0]": "echo",
			"lifecycleHooks.postStart.exec.command[1]": "run after start",
			"lifecycleHooks.preStop.exec.command[0]":   "echo",
			"lifecycleHooks.preStop.exec.command[1]":   "run before stop",
		},
	)

	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	require.Nil(t, appContainer.Lifecycle)
}
