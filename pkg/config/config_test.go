package config_test

import (
	"os"
	"testing"

	"github.com/kuoss/dummy-exporter/pkg/config"
	"github.com/kuoss/dummy-exporter/pkg/fakemetrics"
	"github.com/kuoss/dummy-exporter/pkg/fakemetrics/generator"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

const testConfigPath = "../../etc/exporter.yaml"

func resetFlagsAndArgs(args []string) {
	os.Args = args
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
}

func TestLoadConfig_DefaultYAML(t *testing.T) {
	resetFlagsAndArgs([]string{"cmd"}) // no CLI flags

	cfg, err := config.LoadConfig(testConfigPath)
	assert.NoError(t, err)

	assert.Equal(t, "0.0.0.0", cfg.Server.Address, "server address should match")
	assert.Equal(t, 9100, cfg.Server.Port, "server port should match")
	assert.Equal(t, "3s", cfg.Server.UpdateInterval, "updateInterval should match default")

	assert.True(t, cfg.Exporters.Enabled, "exporters.enabled should be true")
	assert.True(t, cfg.Exporters.DcgmExporter.Enabled, "dcgmExporter.enabled should be true")
	assert.True(t, cfg.Exporters.KubeStateMetrics.Enabled, "kubeStateMetrics.enabled should be true")
	assert.True(t, cfg.Exporters.NodeExporter.Enabled, "nodeExporter.enabled should be true")

	assert.Len(t, cfg.CustomMetrics, 10, "expected 10 custom metrics")
	assert.Equal(t, "custom_always_up", cfg.CustomMetrics[0].Name, "first metric name should match")

	want0 := fakemetrics.Options{
		Name:       "custom_always_up",
		Type:       "gauge",
		Help:       "Always up metric",
		Generator:  generator.Options{Type: "fixed", Value: 1, Values: []float64(nil)},
		Labels:     map[string]string(nil),
		LabelNames: []string(nil),
	}
	assert.Equal(t, want0, cfg.CustomMetrics[0])

	want := []fakemetrics.Options{
		{Name: "custom_always_up", Type: "gauge", Help: "Always up metric", Generator: generator.Options{Type: "fixed", Value: 1, Values: []float64(nil)}, Labels: map[string]string(nil), LabelNames: []string(nil)},
		{Name: "custom_the_answer", Type: "gauge", Help: "The answer to life", Generator: generator.Options{Type: "fixed", Value: 42, Values: []float64(nil)}, Labels: map[string]string(nil), LabelNames: []string(nil)},
		{Name: "custom_cycle_up", Type: "gauge", Help: "Cycles an up/down state (0 = down, 1 = up)", Generator: generator.Options{Type: "round_robin", Value: 0, Values: []float64{0, 1}}, Labels: map[string]string(nil), LabelNames: []string(nil)},
		{Name: "custom_cycle_code", Type: "gauge", Help: "Cycles through HTTP response codes", Generator: generator.Options{Type: "round_robin", Value: 0, Values: []float64{200, 301, 400, 404, 500}}, Labels: map[string]string(nil), LabelNames: []string(nil)},
		{Name: "custom_coin", Type: "gauge", Help: "Random coin flip (0 = tails, 1 = heads)", Generator: generator.Options{Type: "rand_int", Value: 0, Values: []float64{0, 1}}, Labels: map[string]string(nil), LabelNames: []string(nil)},
		{Name: "custom_dice", Type: "gauge", Help: "Simulates a dice roll (1-6)", Generator: generator.Options{Type: "rand_int", Value: 0, Values: []float64{1, 6}}, Labels: map[string]string(nil), LabelNames: []string(nil)},
		{Name: "custom_temperature", Type: "gauge", Help: "Simulated temperature (°C)", Generator: generator.Options{Type: "rand_float", Value: 0, Values: []float64{-5, 30}}, Labels: map[string]string(nil), LabelNames: []string(nil)},
		{Name: "custom_http_requests_total", Type: "counter", Help: "Total HTTP requests", Generator: generator.Options{Type: "", Value: 0, Values: []float64(nil)}, Labels: map[string]string(nil), LabelNames: []string(nil)},
		{Name: "custom_request_duration_seconds", Type: "histogram", Help: "Request durations", Generator: generator.Options{Type: "", Value: 0, Values: []float64(nil)}, Labels: map[string]string(nil), LabelNames: []string(nil)},
		{Name: "custom_latency_summary", Type: "summary", Help: "Latency quantiles", Generator: generator.Options{Type: "", Value: 0, Values: []float64(nil)}, Labels: map[string]string(nil), LabelNames: []string(nil)},
	}

	assert.Equal(t, want, cfg.CustomMetrics)
}

func TestLoadConfig_OverrideWithEmptyCustomMetrics(t *testing.T) {
	resetFlagsAndArgs([]string{
		"cmd",
		"--exporters.dcgmExporter.enabled=false",
		"--customMetrics=",
	})

	cfg, err := config.LoadConfig(testConfigPath)
	assert.NoError(t, err)

	assert.False(t, cfg.Exporters.DcgmExporter.Enabled, "dcgmExporter should be overridden to false")
	assert.Len(t, cfg.CustomMetrics, 0, "customMetrics should be empty")
}

func TestLoadConfig_OverrideWithEmptyArray(t *testing.T) {
	resetFlagsAndArgs([]string{
		"cmd",
		"--exporters.dcgmExporter.enabled=false",
		"--customMetrics=[]",
	})

	cfg, err := config.LoadConfig(testConfigPath)
	assert.NoError(t, err)

	assert.False(t, cfg.Exporters.DcgmExporter.Enabled, "dcgmExporter should be overridden to false")
	assert.Len(t, cfg.CustomMetrics, 0, "customMetrics should be empty array")
}

func TestLoadConfig_OverrideWithValidCustomMetrics(t *testing.T) {
	resetFlagsAndArgs([]string{
		"cmd",
		`--customMetrics=[{"name":"cli_metric","type":"gauge","help":"from cli","generator":{"type":"fixed","value":1}}]`,
	})

	cfg, err := config.LoadConfig(testConfigPath)
	assert.NoError(t, err)

	assert.Len(t, cfg.CustomMetrics, 1, "should have 1 CLI-defined metric")
	assert.Equal(t, "cli_metric", cfg.CustomMetrics[0].Name)
	assert.Equal(t, "gauge", string(cfg.CustomMetrics[0].Type))
}
