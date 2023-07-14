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

			"serviceAccount.name":   "secret-sa",
			"secrets.dbsettings.as": "csi",

			"secrets.dbsettings.mountPath":    "/etc/db",
			"secrets.dbsettings.csi.driver":   "secrets-store.csi.k8s.io",
			"secrets.dbsettings.csi.readOnly": "true",

			"secrets.dbsettings.csi.volumeAttributes.secretProviderClass": "secret-provider-class",

			"secrets.dbsettings.items[0].name":                        "ENV_1",
			"secrets.dbsettings.items[0].valueFrom.secretKeyRef.name": "dbsettings",
			"secrets.dbsettings.items[0].valueFrom.secretKeyRef.key":  "ENV_1",
			"secrets.dbsettings.items[1].name":                        "ENV_2",
			"secrets.dbsettings.items[1].valueFrom.secretKeyRef.name": "dbsettings",
			"secrets.dbsettings.items[1].valueFrom.secretKeyRef.key":  "ENV_2",
		},
	)

	// Verify that there is only one container and only one volume
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	renderedPodVolumes := deployment.Spec.Template.Spec.Volumes
	require.Equal(t, len(renderedPodVolumes), 1)
	podVolume := renderedPodVolumes[0]

	// Check that the pod volume is a secret volume
	assert.Equal(t, podVolume.Name, "dbsettings-volume")

	// Check that the pod volume has CSI block
	require.NotNil(t, podVolume.CSI)
	assert.Equal(t, podVolume.CSI.Driver, "secrets-store.csi.k8s.io")
	assert.NotNil(t, podVolume.CSI.VolumeAttributes)
	assert.Equal(t, podVolume.CSI.VolumeAttributes, map[string]string{
		"secretProviderClass": "secret-provider-class",
	})

	// Check that the container env contains the ENV from secrets
	require.NotNil(t, appContainer.Env)
	assert.Equal(t, appContainer.Env[0].Name, "ENV_1")
	assert.Equal(t, appContainer.Env[0].ValueFrom.SecretKeyRef.Name, "dbsettings")
	assert.Equal(t, appContainer.Env[0].ValueFrom.SecretKeyRef.Key, "ENV_1")
	assert.Equal(t, appContainer.Env[1].Name, "ENV_2")
	assert.Equal(t, appContainer.Env[1].ValueFrom.SecretKeyRef.Name, "dbsettings")
	assert.Equal(t, appContainer.Env[1].ValueFrom.SecretKeyRef.Key, "ENV_2")
}
