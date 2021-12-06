// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

// Test that setting the `envVars` input value to empty object leaves env vars out of the pod.
func TestK8SServiceEnvVarConfigMapsSecretsEmptyDoesNotAddEnvVarsAndVolumesToPod(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(t, map[string]string{})

	// Verify that there is only one container
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]

	// ... and that there are no environments
	environments := appContainer.Env
	assert.Equal(t, len(environments), 0)

	// ... or volumes configured
	renderedPodVolumes := deployment.Spec.Template.Spec.Volumes
	assert.Equal(t, len(renderedPodVolumes), 0)
}

// Test that setting the `envVars` input value will include those environment vars
// We test by injecting to the envVars:
// DB_HOST: "mysql.default.svc.cluster.local"
// DB_PORT: 3306
func TestK8SServiceEnvVarAddsEnvVarsToPod(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"envVars.DB_HOST": "mysql.default.svc.cluster.local",
			"envVars.DB_PORT": "3306",
		},
	)

	// Verify that there is only one container and that the environments section is populated.
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	environments := appContainer.Env
	assert.Equal(t, len(environments), 2)

	renderedEnvVar := map[string]string{}
	for _, env := range environments {
		renderedEnvVar[env.Name] = env.Value
	}
	assert.Equal(t, renderedEnvVar["DB_HOST"], "mysql.default.svc.cluster.local")
	assert.Equal(t, renderedEnvVar["DB_PORT"], "3306")
}

// Test that setting the `additionalContainerEnv` input value will include those environment vars
// We test by injecting:
// additionalContainerEnv:
//   - name: DD_AGENT_HOST
//     valueFrom:
//       fieldRef:
//         fieldPath: status.hostIP
//   - name: DD_ENTITY_ID
//     valueFrom:
//       fieldRef:
//         fieldPath: metadata.uid
func TestK8SServiceAdditionalEnvVarAddsEnvVarsToPod(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"additionalContainerEnv[0].name":                         "DD_AGENT_HOST",
			"additionalContainerEnv[0].valueFrom.fieldRef.fieldPath": "status.hostIP",
			"additionalContainerEnv[1].name":                         "DD_ENTITY_ID",
			"additionalContainerEnv[1].valueFrom.fieldRef.fieldPath": "metadata.uid",
		},
	)

	// Verify that there is only one container and that the environments section is populated.
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	environments := appContainer.Env
	assert.Equal(t, len(environments), 2)

	renderedEnvVar := map[string]string{}
	for _, env := range environments {
		renderedEnvVar[env.Name] = env.ValueFrom.FieldRef.FieldPath
	}
	assert.Equal(t, renderedEnvVar["DD_AGENT_HOST"], "status.hostIP")
	assert.Equal(t, renderedEnvVar["DD_ENTITY_ID"], "metadata.uid")
}

// Test that setting the `configMaps` input value with environment include those environment vars
// We test by injecting to configMaps:
// configMaps:
//   dbsettings:
//     as: environment
//     items:
//       host:
//         envVarName: DB_HOST
//       port:
//         envVarName: DB_PORT
func TestK8SServiceEnvironmentConfigMapAddsEnvVarsToPod(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"configMaps.dbsettings.as":                    "environment",
			"configMaps.dbsettings.items.host.envVarName": "DB_HOST",
			"configMaps.dbsettings.items.port.envVarName": "DB_PORT",
		},
	)

	// Verify that there is only one container and that the environments section is empty.
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	environments := appContainer.Env
	assert.Equal(t, len(environments), 2)

	// Read in the configured env vars for convenient mapping of env var name
	renderedEnvVar := map[string]corev1.EnvVar{}
	for _, env := range environments {
		renderedEnvVar[env.Name] = env
	}

	// Verify the DB_HOST env var comes from config map host key of dbsettings
	assert.Equal(t, renderedEnvVar["DB_HOST"].Value, "")
	require.NotNil(t, renderedEnvVar["DB_HOST"].ValueFrom)
	require.NotNil(t, renderedEnvVar["DB_HOST"].ValueFrom.ConfigMapKeyRef)
	assert.Equal(t, renderedEnvVar["DB_HOST"].ValueFrom.ConfigMapKeyRef.Key, "host")
	assert.Equal(t, renderedEnvVar["DB_HOST"].ValueFrom.ConfigMapKeyRef.Name, "dbsettings")

	// Verify the DB_PORT env var comes from config map port key of dbsettings
	assert.Equal(t, renderedEnvVar["DB_PORT"].Value, "")
	require.NotNil(t, renderedEnvVar["DB_PORT"].ValueFrom)
	require.NotNil(t, renderedEnvVar["DB_PORT"].ValueFrom.ConfigMapKeyRef)
	assert.Equal(t, renderedEnvVar["DB_PORT"].ValueFrom.ConfigMapKeyRef.Key, "port")
	assert.Equal(t, renderedEnvVar["DB_PORT"].ValueFrom.ConfigMapKeyRef.Name, "dbsettings")
}

// Test that setting the `configMaps` input value with volume include the volume mount for the config map
// We test by injecting to configMaps:
// configMaps:
//   dbsettings:
//     as: volume
//     mountPath: /etc/db
func TestK8SServiceVolumeConfigMapAddsVolumeAndVolumeMountWithoutSubPathToPod(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"configMaps.dbsettings.as":        "volume",
			"configMaps.dbsettings.mountPath": "/etc/db",
		},
	)

	// Verify that there is only one container and only one volume
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	renderedPodVolumes := deployment.Spec.Template.Spec.Volumes
	require.Equal(t, len(renderedPodVolumes), 1)
	podVolume := renderedPodVolumes[0]

	// Check that the pod volume is a configmap volume
	assert.Equal(t, podVolume.Name, "dbsettings-volume")
	require.NotNil(t, podVolume.ConfigMap)
	assert.Equal(t, podVolume.ConfigMap.Name, "dbsettings")

	// Check that the pod volume will be mounted
	require.Equal(t, len(appContainer.VolumeMounts), 1)
	volumeMount := appContainer.VolumeMounts[0]
	assert.Equal(t, volumeMount.Name, "dbsettings-volume")
	assert.Equal(t, volumeMount.MountPath, "/etc/db")
	assert.Empty(t, volumeMount.SubPath)
}

// Test that setting the `configMaps` input value with volume include the volume mount and subpath for the config map
// We test by injecting to configMaps:
// configMaps:
//   dbsettings:
//     as: volume
//     mountPath: /etc/db/host.txt
//     subPath: host.xt
func TestK8SServiceVolumeConfigMapAddsVolumeAndVolumeMountWithSubPathToPod(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"configMaps.dbsettings.as":        "volume",
			"configMaps.dbsettings.mountPath": "/etc/db/host.txt",
			"configMaps.dbsettings.subPath":   "host.txt",
		},
	)

	// Verify that there is only one container and only one volume
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	renderedPodVolumes := deployment.Spec.Template.Spec.Volumes
	require.Equal(t, len(renderedPodVolumes), 1)
	podVolume := renderedPodVolumes[0]

	// Check that the pod volume is a configmap volume
	assert.Equal(t, podVolume.Name, "dbsettings-volume")
	require.NotNil(t, podVolume.ConfigMap)
	assert.Equal(t, podVolume.ConfigMap.Name, "dbsettings")

	// Check that the pod volume will be mounted
	require.Equal(t, len(appContainer.VolumeMounts), 1)
	volumeMount := appContainer.VolumeMounts[0]
	assert.Equal(t, volumeMount.Name, "dbsettings-volume")
	assert.Equal(t, volumeMount.MountPath, "/etc/db/host.txt")
	assert.Equal(t, volumeMount.SubPath, "host.txt")
}

// Test that setting the `configMaps` input value with volume and individual file mount paths will set the appropriate
// settings
// We test by injecting to configMaps:
// configMaps:
//   dbsettings:
//     as: volume
//     mountPath: /etc/db
//     items:
//       host:
//         filePath: host.txt
func TestK8SServiceVolumeConfigMapWithKeyFilePathAddsVolumeWithKeyFilePathToPod(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"configMaps.dbsettings.as":                  "volume",
			"configMaps.dbsettings.mountPath":           "/etc/db",
			"configMaps.dbsettings.items.host.filePath": "host.txt",
		},
	)

	// Verify that there is only one volume
	renderedPodVolumes := deployment.Spec.Template.Spec.Volumes
	require.Equal(t, len(renderedPodVolumes), 1)
	podVolume := renderedPodVolumes[0]

	// Check that the pod volume is a configmap volume and has a file path instruction for host key
	assert.Equal(t, podVolume.Name, "dbsettings-volume")
	require.NotNil(t, podVolume.ConfigMap)
	assert.Equal(t, podVolume.ConfigMap.Name, "dbsettings")
	require.Equal(t, len(podVolume.ConfigMap.Items), 1)
	keyToPath := podVolume.ConfigMap.Items[0]
	assert.Equal(t, keyToPath.Key, "host")
	assert.Equal(t, keyToPath.Path, "host.txt")
}

// Test the file mode calculation. We test by injecting to configMaps the following with different file mode octals and
// validating the decimal:
// configMaps:
//   dbsettings:
//     as: volume
//     mountPath: /etc/db
//     items:
//       host:
//         filePath: host.txt
//         fileMode: 644
func TestK8SServiceVolumeConfigMapWithKeyFileModeConvertsOctalToDecimal(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		octal   string
		decimal int32
	}{
		{"644", 420},
		{"700", 448},
		{"755", 493},
		{"777", 511},
	}
	for _, testCase := range testCases {
		// Capture range variable to force scope
		testCase := testCase
		t.Run(testCase.octal, func(t *testing.T) {
			t.Parallel()
			checkFileMode(t, "configMaps", testCase.octal, testCase.decimal)
		})
	}
}

// Test that setting the `secrets` input value with environment include those environment vars
// We test by injecting to secrets:
// secrets:
//   dbsettings:
//     as: environment
//     items:
//       host:
//         envVarName: DB_HOST
//       port:
//         envVarName: DB_PORT
func TestK8SServiceEnvironmentSecretAddsEnvVarsToPod(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"secrets.dbsettings.as":                    "environment",
			"secrets.dbsettings.items.host.envVarName": "DB_HOST",
			"secrets.dbsettings.items.port.envVarName": "DB_PORT",
		},
	)

	// Verify that there is only one container and that the environments section is empty.
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	environments := appContainer.Env
	assert.Equal(t, len(environments), 2)

	// Read in the configured env vars for convenient mapping of env var name
	renderedEnvVar := map[string]corev1.EnvVar{}
	for _, env := range environments {
		renderedEnvVar[env.Name] = env
	}

	// Verify the DB_HOST env var comes from secret host key of dbsettings
	assert.Equal(t, renderedEnvVar["DB_HOST"].Value, "")
	require.NotNil(t, renderedEnvVar["DB_HOST"].ValueFrom)
	require.NotNil(t, renderedEnvVar["DB_HOST"].ValueFrom.SecretKeyRef)
	assert.Equal(t, renderedEnvVar["DB_HOST"].ValueFrom.SecretKeyRef.Key, "host")
	assert.Equal(t, renderedEnvVar["DB_HOST"].ValueFrom.SecretKeyRef.Name, "dbsettings")

	// Verify the DB_PORT env var comes from secret port key of dbsettings
	assert.Equal(t, renderedEnvVar["DB_PORT"].Value, "")
	require.NotNil(t, renderedEnvVar["DB_PORT"].ValueFrom)
	require.NotNil(t, renderedEnvVar["DB_PORT"].ValueFrom.SecretKeyRef)
	assert.Equal(t, renderedEnvVar["DB_PORT"].ValueFrom.SecretKeyRef.Key, "port")
	assert.Equal(t, renderedEnvVar["DB_PORT"].ValueFrom.SecretKeyRef.Name, "dbsettings")
}

// Test that setting the `secrets` input value with volume include the volume mount for the secret
// We test by injecting to secrets:
// secrets:
//   dbsettings:
//     as: volume
//     mountPath: /etc/db
func TestK8SServiceVolumeSecretAddsVolumeAndVolumeMountToPod(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"secrets.dbsettings.as":        "volume",
			"secrets.dbsettings.mountPath": "/etc/db",
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
	require.NotNil(t, podVolume.Secret)
	assert.Equal(t, podVolume.Secret.SecretName, "dbsettings")

	// Check that the pod volume will be mounted
	require.Equal(t, len(appContainer.VolumeMounts), 1)
	volumeMount := appContainer.VolumeMounts[0]
	assert.Equal(t, volumeMount.Name, "dbsettings-volume")
	assert.Equal(t, volumeMount.MountPath, "/etc/db")
}

// Test that setting the `secrets` input value with volume and individual file mount paths will set the appropriate
// settings
// We test by injecting to secrets:
// secrets:
//   dbsettings:
//     as: volume
//     mountPath: /etc/db
//     items:
//       host:
//         filePath: host.txt
func TestK8SServiceVolumeSecretWithKeyFilePathAddsVolumeWithKeyFilePathToPod(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"secrets.dbsettings.as":                  "volume",
			"secrets.dbsettings.mountPath":           "/etc/db",
			"secrets.dbsettings.items.host.filePath": "host.txt",
		},
	)

	// Verify that there is only one volume
	renderedPodVolumes := deployment.Spec.Template.Spec.Volumes
	require.Equal(t, len(renderedPodVolumes), 1)
	podVolume := renderedPodVolumes[0]

	// Check that the pod volume is a secret volume and has a file path instruction for host key
	assert.Equal(t, podVolume.Name, "dbsettings-volume")
	require.NotNil(t, podVolume.Secret)
	assert.Equal(t, podVolume.Secret.SecretName, "dbsettings")
	require.Equal(t, len(podVolume.Secret.Items), 1)
	keyToPath := podVolume.Secret.Items[0]
	assert.Equal(t, keyToPath.Key, "host")
	assert.Equal(t, keyToPath.Path, "host.txt")
}

// Test the file mode calculation. We test by injecting to secrets the following with different file mode octals and
// validating the decimal:
// secrets:
//   dbsettings:
//     as: volume
//     mountPath: /etc/db
//     items:
//       host:
//         filePath: host.txt
//         fileMode: 644
func TestK8SServiceVolumeSecretWithKeyFileModeConvertsOctalToDecimal(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		octal   string
		decimal int32
	}{
		{"644", 420},
		{"700", 448},
		{"755", 493},
		{"777", 511},
	}
	for _, testCase := range testCases {
		// Capture range variable to force scope
		testCase := testCase
		t.Run(testCase.octal, func(t *testing.T) {
			t.Parallel()
			checkFileMode(t, "secrets", testCase.octal, testCase.decimal)
		})
	}
}

// Test the file mode calculation assertions. We test by injecting to secrets the following with different file mode
// octals and checking that it fails
// secrets:
//   dbsettings:
//     as: volume
//     mountPath: /etc/db
//     items:
//       host:
//         filePath: host.txt
//         fileMode: 644
func TestK8SServiceFileModeOctalToDecimalAssertions(t *testing.T) {
	t.Parallel()
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	testCases := []string{
		"800",  // First digit greater than max (7)
		"080",  // Second digit greater than max (7)
		"008",  // Third digit greater than max (7)
		"nan",  // Not a number
		"75n",  // Not a number
		"0644", // Not three digits
		"44",   // Not three digits
	}
	for _, testCase := range testCases {
		// Capture range variable to force scope
		testCase := testCase
		t.Run(testCase, func(t *testing.T) {
			t.Parallel()
			// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values
			// defined.
			options := &helm.Options{
				ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
				SetValues: map[string]string{
					"secrets.dbsettings.as":                  "volume",
					"secrets.dbsettings.mountPath":           "/etc/db",
					"secrets.dbsettings.items.host.filePath": "host.txt",
					"secrets.dbsettings.items.host.fileMode": testCase,
				},
			}
			// Render just the deployment resource
			_, err := helm.RenderTemplateE(t, options, helmChartPath, strings.ToLower(t.Name()), []string{"templates/deployment.yaml"})
			assert.Error(t, err)
		})
	}
}

// Test that setting the `secrets` and `configMaps` input value with volume include the volume mount for both.
// We test by injecting to secrets and configMaps:
// configMaps:
//   dbsettings:
//     as: volume
//     mountPath: /etc/db
// secrets:
//   dbpassword:
//     as: volume
//     mountPath: /etc/dbpass
func TestK8SServiceVolumeSecretAndConfigMapAddsBothVolumesAndVolumeMountsToPod(t *testing.T) {
	t.Parallel()

	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			"configMaps.dbsettings.as":        "volume",
			"configMaps.dbsettings.mountPath": "/etc/db",
			"secrets.dbpassword.as":           "volume",
			"secrets.dbpassword.mountPath":    "/etc/dbpass",
		},
	)

	// Verify that there is only one container and only one volume
	renderedPodContainers := deployment.Spec.Template.Spec.Containers
	require.Equal(t, len(renderedPodContainers), 1)
	appContainer := renderedPodContainers[0]
	renderedPodVolumes := deployment.Spec.Template.Spec.Volumes
	require.Equal(t, len(renderedPodVolumes), 2)

	// Map volumes to a map for easy lookup
	volumes := map[string]corev1.Volume{}
	for _, volume := range renderedPodVolumes {
		volumes[volume.Name] = volume
	}

	// Check configMap pod volume exists
	volume := volumes["dbsettings-volume"]
	require.NotNil(t, volume.ConfigMap)
	assert.Equal(t, volume.ConfigMap.Name, "dbsettings")

	// Check secret pod volume exists
	volume = volumes["dbpassword-volume"]
	require.NotNil(t, volume.Secret)
	assert.Equal(t, volume.Secret.SecretName, "dbpassword")

	// Check that both volumes will be mounted on the pod in the specified paths
	volumeMounts := map[string]corev1.VolumeMount{}
	for _, mount := range appContainer.VolumeMounts {
		volumeMounts[mount.Name] = mount
	}
	assert.Equal(t, volumeMounts["dbsettings-volume"].MountPath, "/etc/db")
	assert.Equal(t, volumeMounts["dbpassword-volume"].MountPath, "/etc/dbpass")
}

func checkFileMode(t *testing.T, configMapsOrSecrets string, fileModeOctal string, fileModeDecimal int32) {
	deployment := renderK8SServiceDeploymentWithSetValues(
		t,
		map[string]string{
			fmt.Sprintf("%s.dbsettings.as", configMapsOrSecrets):                  "volume",
			fmt.Sprintf("%s.dbsettings.mountPath", configMapsOrSecrets):           "/etc/db",
			fmt.Sprintf("%s.dbsettings.items.host.filePath", configMapsOrSecrets): "host.txt",
			fmt.Sprintf("%s.dbsettings.items.host.fileMode", configMapsOrSecrets): fileModeOctal,
		},
	)

	// Verify that there is only one volume
	renderedPodVolumes := deployment.Spec.Template.Spec.Volumes
	require.Equal(t, len(renderedPodVolumes), 1)
	podVolume := renderedPodVolumes[0]
	assert.Equal(t, podVolume.Name, "dbsettings-volume")

	// Check that the pod volume is a configmap/secret volume and has a file mode instruction with decimal value of octal
	switch configMapsOrSecrets {
	case "configMaps":
		require.NotNil(t, podVolume.ConfigMap)
		assert.Equal(t, podVolume.ConfigMap.Name, "dbsettings")
		require.Equal(t, len(podVolume.ConfigMap.Items), 1)
		keyToPath := podVolume.ConfigMap.Items[0]
		assert.Equal(t, keyToPath.Key, "host")
		require.NotNil(t, keyToPath.Mode)
		assert.Equal(t, *keyToPath.Mode, fileModeDecimal)
	case "secrets":
		require.NotNil(t, podVolume.Secret)
		assert.Equal(t, podVolume.Secret.SecretName, "dbsettings")
		require.Equal(t, len(podVolume.Secret.Items), 1)
		keyToPath := podVolume.Secret.Items[0]
		assert.Equal(t, keyToPath.Key, "host")
		require.NotNil(t, keyToPath.Mode)
		assert.Equal(t, *keyToPath.Mode, fileModeDecimal)
	default:
		t.Fatalf("Unexpected attribute name: %s", configMapsOrSecrets)
	}
}
