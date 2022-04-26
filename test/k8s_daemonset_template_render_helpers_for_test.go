//go:build all || tpl
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
	networkingv1 "k8s.io/api/networking/v1"

	certapi "github.com/GoogleCloudPlatform/gke-managed-certs/pkg/apis/networking.gke.io/v1beta1"
)

func renderK8SServiceDeploymentWithSetValues(t *testing.T, setValues map[string]string) appsv1.DaemonSet {
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-daemonset"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-daemonset", "linter_values.yaml")},
		SetValues:   setValues,
	}
	// Render just the daemonset resource
	out := helm.RenderTemplate(t, options, helmChartPath, "daemonset", []string{"templates/daemonset.yaml"})

	// Parse the daemonset and return it
	var daemonset appsv1.DaemonSet
	helm.UnmarshalK8SYaml(t, out, &daemonset)
	return daemonset
}

func renderK8SServiceCanaryDeploymentWithSetValues(t *testing.T, setValues map[string]string) appsv1.DaemonSet {
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-daemonset"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-daemonset", "linter_values.yaml")},
		SetValues:   setValues,
	}
	// Render just the canary daemonset resource
	out := helm.RenderTemplate(t, options, helmChartPath, "canarydaemonset", []string{"templates/canarydaemonset.yaml"})

	// Parse the canary daemonset and return it
	var canarydaemonset appsv1.DaemonSet
	helm.UnmarshalK8SYaml(t, out, &canarydaemonset)
	return canarydaemonset
}

func renderK8SServiceIngressWithSetValues(t *testing.T, setValues map[string]string) networkingv1.Ingress {
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-daemonset"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-daemonset", "linter_values.yaml")},
		SetValues:   setValues,
	}
	// Render just the ingress resource
	out := helm.RenderTemplate(t, options, helmChartPath, "ingress", []string{"templates/ingress.yaml"})

	// Parse the ingress and return it
	var ingress networkingv1.Ingress
	helm.UnmarshalK8SYaml(t, out, &ingress)
	return ingress
}

func renderK8SServiceIngressWithValuesFile(t *testing.T, valuesFilePath string) networkingv1.Ingress {
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-daemonset"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{
			filepath.Join("..", "charts", "k8s-daemonset", "linter_values.yaml"),
			valuesFilePath,
		},
	}
	// Render just the ingress resource
	out := helm.RenderTemplate(t, options, helmChartPath, "ingress", []string{"templates/ingress.yaml"})

	// Parse the ingress and return it
	var ingress networkingv1.Ingress
	helm.UnmarshalK8SYaml(t, out, &ingress)
	return ingress
}

func renderK8SServiceExtV1Beta1IngressWithSetValues(t *testing.T, setValues map[string]string) extv1beta1.Ingress {
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-daemonset"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-daemonset", "linter_values.yaml")},
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
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-daemonset"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-daemonset", "linter_values.yaml")},
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
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-daemonset"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-daemonset", "linter_values.yaml")},
		SetValues:   setValues,
	}
	// Render just the service account resource
	out := helm.RenderTemplate(t, options, helmChartPath, "serviceaccount", []string{"templates/serviceaccount.yaml"})

	// Parse the service account and return it
	var serviceaccount corev1.ServiceAccount
	helm.UnmarshalK8SYaml(t, out, &serviceaccount)
	return serviceaccount
}

func renderK8SServiceWithSetValues(t *testing.T, setValues map[string]string) corev1.Service {
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-daemonset"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-daemonset", "linter_values.yaml")},
		SetValues:   setValues,
	}
	// Render just the service resource
	out := helm.RenderTemplate(t, options, helmChartPath, "service", []string{"templates/service.yaml"})

	// Parse the service and return it
	var service corev1.Service
	helm.UnmarshalK8SYaml(t, out, &service)
	return service
}
