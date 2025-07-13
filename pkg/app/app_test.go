package app

import (
	"testing"

	"github.com/kuoss/dummy-exporter/pkg/config"
	"github.com/kuoss/dummy-exporter/pkg/fakemetrics"
	"github.com/stretchr/testify/assert"
)

func TestGetMetrics_WithCustomMetricsOnly(t *testing.T) {
	cfg := &config.Config{
		Exporters: config.Exporters{Enabled: false},
		CustomMetrics: []fakemetrics.Options{
			{
				Name: "test_metric",
				Type: "gauge",
				Help: "a test metric",
			},
		},
	}

	metrics, err := getMetrics(cfg)
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "test_metric", metrics[0].Name)
}
