// +build all tpl

// NOTE: We use build flags to differentiate between template tests and integration tests so that you can conveniently
// run just the template tests. See the test README for more information.

package test

import (
	"path/filepath"
	"testing"

	promethues_operator_v1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/ghodss/yaml"
	"github.com/gruntwork-io/terratest/modules/helm"
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
	out := helm.RenderTemplate(t, options, helmChartPath, []string{"templates/servicemonitor.yaml"})

	// We take the output and render it to a map to validate it is an empty yaml
	rendered := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(out), &rendered)
	assert.NoError(t, err)
	assert.Equal(t, len(rendered), 0)
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
	out := helm.RenderTemplate(t, options, helmChartPath, []string{"templates/servicemonitor.yaml"})

	// We take the output and render it to a map to validate it is an empty yaml
	rendered := promethues_operator_v1.ServiceMonitor{}
	require.NoError(t, yaml.Unmarshal([]byte(out), &rendered))
	require.Equal(t, len(rendered.Spec.Endpoints), 1)

	// check the default endpoint properties
	defaultEndpoint := rendered.Spec.Endpoints[0]
	assert.Equal(t, defaultEndpoint.Interval, "10s")
	assert.Equal(t, defaultEndpoint.ScrapeTimeout, "10s")
	assert.Equal(t, defaultEndpoint.Path, "/metrics")
	assert.Equal(t, defaultEndpoint.Port, "http")
	assert.Equal(t, defaultEndpoint.Scheme, "http")
}
