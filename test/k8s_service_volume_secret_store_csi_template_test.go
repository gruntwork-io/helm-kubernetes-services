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

func TestK8SServiceDeploymentCheckSecretStoreCSIBlock(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"secrets.dbsettings.as":        "csi",
			"secrets.dbsettings.mountPath": "/etc/db",
			"secrets.dbsettings.readOnly":  "true",

			"secrets.dbsettings.csi.driver":              "secrets-store.csi.k8s.io",
			"secrets.dbsettings.csi.secretProviderClass": "secret-provider-class",

			"secrets.dbsettings.items.host.envVarName": "DB_HOST",
			"secrets.dbsettings.items.port.envVarName": "DB_PORT",
		},
	)

	// Verify that there is only one container and only one volume
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	renderedPodVolumes := deployment.Spec.Template.Spec.Volumes
	require.Equal(t, len(renderedPodVolumes), 1)
	podVolume := renderedPodVolumes[0]

	// Check that the pod volume has a correct name
	assert.Equal(t, podVolume.Name, "dbsettings-volume")
	
	// Check that the pod volume has CSI block
	assert.NotNil(t, podVolume.CSI)

	// Check that the pod volume has correct CSI driver and attributes
	assert.Equal(t, podVolume.CSI.Driver, "secrets-store.csi.k8s.io")
	assert.NotNil(t, podVolume.CSI.VolumeAttributes)
	assert.Equal(t, podVolume.CSI.VolumeAttributes, map[string]string{
		"secretProviderClass": "secret-provider-class",
	})
}
