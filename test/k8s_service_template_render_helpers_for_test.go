// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"

	certapi "github.com/GoogleCloudPlatform/gke-managed-certs/pkg/apis/networking.gke.io/v1beta1"
)

func renderK8SServiceDeploymentWithSetValues(t *testing.T, setValues map[string]string) appsv1.Deployment {
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   setValues,
	}
	// Render just the deployment resource
	out := helm.RenderTemplate(t, options, helmChartPath, "deployment", []string{"templates/deployment.yaml"})

	// Parse the deployment and return it
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(t, out, &deployment)
	return deployment
}

func renderK8SServiceCanaryDeploymentWithSetValues(t *testing.T, setValues map[string]string) appsv1.Deployment {
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   setValues,
	}
	// Render just the canary deployment resource
	out := helm.RenderTemplate(t, options, helmChartPath, "canarydeployment", []string{"templates/canarydeployment.yaml"})

	// Parse the canary deployment and return it
	var canarydeployment appsv1.Deployment
	helm.UnmarshalK8SYaml(t, out, &canarydeployment)
	return canarydeployment
}

func renderK8SServiceIngressWithSetValues(t *testing.T, setValues map[string]string) extv1beta1.Ingress {
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   setValues,
	}
	// Render just the ingress resource
	out := helm.RenderTemplate(t, options, helmChartPath, "ingress", []string{"templates/ingress.yaml"})

	// Parse the ingress and return it
	var ingress extv1beta1.Ingress
	helm.UnmarshalK8SYaml(t, out, &ingress)
	return ingress
}

func renderK8SServiceManagedCertificateWithSetValues(t *testing.T, setValues map[string]string) certapi.ManagedCertificate {
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   setValues,
	}
	// Render just the google managed certificate resource
	out := helm.RenderTemplate(t, options, helmChartPath, "gmc", []string{"templates/gmc.yaml"})

	// Parse the google managed certificate and return it
	var cert certapi.ManagedCertificate
	helm.UnmarshalK8SYaml(t, out, &cert)
	return cert
}

func renderK8SServiceAccountWithSetValues(t *testing.T, setValues map[string]string) corev1.ServiceAccount {
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   setValues,
	}
	// Render just the service account resource
	out := helm.RenderTemplate(t, options, helmChartPath, "serviceaccount", []string{"templates/serviceaccount.yaml"})

	// Parse the service account and return it
	var serviceaccount corev1.ServiceAccount
	helm.UnmarshalK8SYaml(t, out, &serviceaccount)
	return serviceaccount
}
