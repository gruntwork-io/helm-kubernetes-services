// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gruntwork-io/terratest/modules/helm"
	prometheus_operator_v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that setting serviceMonitor.enabled = false will cause the helm template to not render the Service Monitor resource
func TestK8SServiceServiceMonitorEnabledFalseDoesNotCreateServiceMonitor(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{filepath.Join("..", "charts", "k8s-service", "linter_values.yaml")},
		SetValues:   map[string]string{"serviceMonitor.enabled": "false"},
	}
	_, err = helm.RenderTemplateE(t, options, helmChartPath, "servicemonitor", []string{"templates/servicemonitor.yaml"})
	require.Error(t, err)
}

// Test that configuring a service monitor will render correctly to something that will be accepted by the Prometheus
// operator
func TestK8SServiceServiceMonitorEnabledCreatesServiceMonitor(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "k8s-service"))
	require.NoError(t, err)

	// We make sure to pass in the linter_values.yaml values file, which we assume has all the required values defined.
	// We then use SetValues to override all the defaults.
	options := &helm.Options{
		ValuesFiles: []string{
			filepath.Join("..", "charts", "k8s-service", "linter_values.yaml"),
			filepath.Join("fixtures", "service_monitor_values.yaml"),
		},
	}
	out := helm.RenderTemplate(t, options, helmChartPath, "servicemonitor", []string{"templates/servicemonitor.yaml"})

	// We take the output and render it to a map to validate it is an empty yaml
	rendered := prometheus_operator_v1.ServiceMonitor{}
	require.NoError(t, yaml.Unmarshal([]byte(out), &rendered))
	require.Equal(t, 1, len(rendered.Spec.Endpoints))

	// check the default endpoint properties
	defaultEndpoint := rendered.Spec.Endpoints[0]
	assert.Equal(t, "10s", defaultEndpoint.Interval)
	assert.Equal(t, "10s", defaultEndpoint.ScrapeTimeout)
	assert.Equal(t, "/metrics", defaultEndpoint.Path)
	assert.Equal(t, "http", defaultEndpoint.Port)
	assert.Equal(t, "http", defaultEndpoint.Scheme)
}
