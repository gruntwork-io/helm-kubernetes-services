//go:build all || tpl
// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestK8SServiceDeploymentAddingScratchVolumes(t *testing.T) {
	t.Parallel()

	volName := "scratch"
	volMountPath := "/mnt/scratch"

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			fmt.Sprintf("scratchPaths.%s", volName): volMountPath,
		},
	)

	// Verify that there is only one container
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	podContainer := renderedPodContainers[0]

	// Verify that a mount has been created for the scratch path
	mounts := podContainer.VolumeMounts
	assert.Equal(t, len(mounts), 1)
	mount := mounts[0]
	assert.Equal(t, volName, mount.Name)
	assert.Equal(t, volMountPath, mount.MountPath)

	// Verify that a volume has been declared for the scratch path and is using tmpfs
	volumes := deployment.Spec.Template.Spec.Volumes
	assert.Equal(t, len(volumes), 1)
	volume := volumes[0]
	assert.Equal(t, volName, volume.Name)
	assert.Equal(t, corev1.StorageMediumMemory, volume.EmptyDir.Medium)

}

func TestK8SServiceDeploymentAddingPersistentVolumes(t *testing.T) {
	t.Parallel()

	volName := "pv-1"
	volClaim := "claim-1"
	volMountPath := "/mnt/path/1"

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"persistentVolumes.pv-1.claimName": volClaim,
			"persistentVolumes.pv-1.mountPath": volMountPath,
		},
	)

	// Verify that there is only one container
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)

	// Verify that a mount has been created for the PV
	mounts := renderedPodContainers[0].VolumeMounts
	assert.Equal(t, len(mounts), 1)
	mount := mounts[0]
	assert.Equal(t, volName, mount.Name)
	assert.Equal(t, volMountPath, mount.MountPath)

	// Verify that a volume has been declared for the PV
	volumes := deployment.Spec.Template.Spec.Volumes
	assert.Equal(t, len(volumes), 1)
	volume := volumes[0]
	assert.Equal(t, volName, volume.Name)
	assert.Equal(t, volClaim, volume.PersistentVolumeClaim.ClaimName)
}

func TestK8SServiceDeploymentAddingEmptyDirs(t *testing.T) {
	t.Parallel()

	volName := "empty-dir"
	volMountPath := "/mnt/empty"

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			fmt.Sprintf("emptyDirs.%s", volName): volMountPath,
		},
	)

	// Verify that there is only one container
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	podContainer := renderedPodContainers[0]

	// Verify that a mount has been created for the emptyDir
	mounts := podContainer.VolumeMounts
	assert.Equal(t, len(mounts), 1)
	mount := mounts[0]
	assert.Equal(t, volName, mount.Name)
	assert.Equal(t, volMountPath, mount.MountPath)

	// Verify that a volume has been declared for the emptyDir
	volumes := deployment.Spec.Template.Spec.Volumes
	assert.Equal(t, len(volumes), 1)
	volume := volumes[0]
	assert.Equal(t, volName, volume.Name)
	assert.Empty(t, volume.EmptyDir)
}

func TestK8SServiceDeploymentAddingTerminationGracePeriod(t *testing.T) {

	gracePeriod := "30"

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"terminationGracePeriodSeconds": gracePeriod,
		},
	)

	// Verify that there is only one container
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)

	expectedGracePeriodInt64, err := strconv.ParseInt(gracePeriod, 10, 64)

	// Verify termination grace period has been set for container
	assert.NoError(t, err)
	renderedTerminationGracePeriodSeconds := deployment.Spec.Template.Spec.TerminationGracePeriodSeconds
	require.Equal(t, expectedGracePeriodInt64, *renderedTerminationGracePeriodSeconds)
}
