//go:build all || tpl
// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	// "fmt"
	// "github.com/gruntwork-io/terratest/modules/random"

	// "github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	// "strings"
	"testing"
)

func TestK8SServiceDeploymentCheckSecretStoreCSIBlock(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"secrets.dbsettings.as":           "volume",
			"secrets.dbsettings.mountPath":    "/etc/db",
			"secrets.dbsettings.csi.driver":   "secrets-store.csi.k8s.io",
			"secrets.dbsettings.csi.readOnly": "true",

			"secrets.dbsettings.csi.volumeAttributes.secretProviderClass": "backend-deployment-aws-secrets",
		},
	)

	// Verify that there is only one container and only one volume
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	// appContainer := renderedPodContainers[0]
	renderedPodVolumes := deployment.Spec.Template.Spec.Volumes
	require.Equal(t, len(renderedPodVolumes), 1)
	podVolume := renderedPodVolumes[0]

	// Check that the pod volume is a secret volume
	assert.Equal(t, podVolume.Name, "dbsettings-volume")
	require.NotNil(t, podVolume.Secret)
	assert.Equal(t, podVolume.Secret.SecretName, "dbsettings")

	// Check that the pod volume has CSI block
	require.NotNil(t, podVolume.CSI)

	assert.Equal(t, podVolume.CSI.Driver, "secrets-store.csi.k8s.io")
	assert.NotNil(t, podVolume.CSI.VolumeAttributes)
	assert.Equal(t, podVolume.CSI.VolumeAttributes, map[string]string{
		"secretProviderClass": "backend-deployment-aws-secrets",
	})

}
