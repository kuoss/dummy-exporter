package config

import (
	"encoding/json"
	"fmt"

	"github.com/kuoss/dummy-exporter/pkg/fakemetrics"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Server        Server                `yaml:"server,omitempty"`
	Collectors    Collectors            `yaml:"collectors,omitempty"`
	Exporters     Exporters             `yaml:"exporters,omitempty"`
	CustomMetrics []fakemetrics.Options `yaml:"customMetrics,omitempty"`
}

type Server struct {
	Address        string `mapstructure:"address"`
	Port           int    `mapstructure:"port"`
	UpdateInterval string `mapstructure:"updateInterval"`
}

type Collectors struct {
	Enabled          bool `yaml:"enabled" mapstructure:"enabled"`
	GoCollector      bool `yaml:"goCollector" mapstructure:"goCollector"`
	ProcessCollector bool `yaml:"processCollector" mapstructure:"processCollector"`
}

type Exporters struct {
	Enabled          bool             `mapstructure:"enabled"`
	DcgmExporter     DcgmExporter     `mapstructure:"dcgmExporter"`
	KubeStateMetrics KubeStateMetrics `mapstructure:"kubeStateMetrics"`
	NodeExporter     NodeExporter     `mapstructure:"nodeExporter"`
}

type DcgmExporter struct {
	Enabled    bool `mapstructure:"enabled"`
	NodeCount  int  `mapstructure:"nodeCount"`
	GPUPerNode int  `mapstructure:"gpuPerNode"`
}

type KubeStateMetrics struct {
	Enabled         bool `mapstructure:"enabled"`
	NamespaceCount  int  `mapstructure:"namespaceCount"`
	PodPerNamespace int  `mapstructure:"podPerNamespace"`
	ContainerPerPod int  `mapstructure:"containerPerPod"`
}

type NodeExporter struct {
	Enabled     bool `mapstructure:"enabled"`
	NodeCount   int  `mapstructure:"nodeCount"`
	CPUPerNode  int  `mapstructure:"cpuPerNode"`
	DiskPerNode int  `mapstructure:"diskPerNode"`
	EthPerNode  int  `mapstructure:"ethPerNode"`
}

func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetDefault("server.updateInterval", "3s")
	// Initialize customMetrics as empty slice to prevent nil pointer issues
	v.SetDefault("customMetrics", []fakemetrics.Options{})

	// Read YAML file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// CLI flags
	pflag.Bool("collectors.goCollector", true, "Enable Go collector")
	pflag.Bool("collectors.processCollector", true, "Enable Process collector")

	pflag.Bool("exporters.dcgmExporter.enabled", false, "Enable DCGM exporter")
	pflag.Bool("exporters.kubeStateMetrics.enabled", false, "Enable KSM exporter")
	pflag.Bool("exporters.nodeExporter.enabled", false, "Enable Node exporter")
	pflag.Int("exporters.nodeExporter.nodeCount", 0, "Number of nodes for Node exporter")
	pflag.Int("exporters.nodeExporter.cpuCount", 0, "Number of CPUs per node")
	pflag.Int("exporters.nodeExporter.diskCount", 0, "Number of disks per node")
	pflag.Int("exporters.nodeExporter.ethCount", 0, "Number of network interfaces per node")
	pflag.String("customMetrics", "", "Custom metrics JSON array")
	pflag.Parse()

	if err := v.BindPFlags(pflag.CommandLine); err != nil {
		return nil, fmt.Errorf("failed to bind CLI flags: %w", err)
	}

	// Handle customMetrics from CLI manually
	raw := v.GetString("customMetrics")
	var cliCustomMetrics []fakemetrics.Options
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &cliCustomMetrics); err != nil {
			return nil, fmt.Errorf("invalid customMetrics JSON: %w", err)
		}
		// Remove the string value to prevent unmarshaling conflicts
		v.Set("customMetrics", []fakemetrics.Options{})
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override with CLI customMetrics if provided
	if raw != "" {
		cfg.CustomMetrics = cliCustomMetrics
	}
	// Ensure CustomMetrics is never nil
	if cfg.CustomMetrics == nil {
		cfg.CustomMetrics = []fakemetrics.Options{}
	}

	// default fixups
	if cfg.Exporters.DcgmExporter.NodeCount <= 0 {
		cfg.Exporters.DcgmExporter.NodeCount = 2
	}
	if cfg.Exporters.DcgmExporter.GPUPerNode <= 0 {
		cfg.Exporters.DcgmExporter.GPUPerNode = 2
	}
	if cfg.Exporters.KubeStateMetrics.NamespaceCount <= 0 {
		cfg.Exporters.KubeStateMetrics.NamespaceCount = 2
	}
	if cfg.Exporters.KubeStateMetrics.PodPerNamespace <= 0 {
		cfg.Exporters.KubeStateMetrics.PodPerNamespace = 2
	}
	if cfg.Exporters.KubeStateMetrics.ContainerPerPod <= 0 {
		cfg.Exporters.KubeStateMetrics.ContainerPerPod = 2
	}
	if cfg.Exporters.NodeExporter.NodeCount <= 0 {
		cfg.Exporters.NodeExporter.NodeCount = 2
	}
	if cfg.Exporters.NodeExporter.CPUPerNode <= 0 {
		cfg.Exporters.NodeExporter.CPUPerNode = 2
	}
	if cfg.Exporters.NodeExporter.DiskPerNode <= 0 {
		cfg.Exporters.NodeExporter.DiskPerNode = 2
	}
	if cfg.Exporters.NodeExporter.EthPerNode <= 0 {
		cfg.Exporters.NodeExporter.EthPerNode = 2
	}

	return &cfg, nil
}
