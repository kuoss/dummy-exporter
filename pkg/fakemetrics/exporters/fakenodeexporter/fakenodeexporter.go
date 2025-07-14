package fakenodeexporter

import (
	"fmt"

	"github.com/kuoss/dummy-exporter/pkg/fakemetrics"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

var metricOptions = []fakemetrics.Options{
	{
		Name: "node_cpu_seconds_total",
		Type: v1.MetricTypeCounter,
		Help: "Total seconds the CPUs spent in each mode",
	},
	{
		Name: "node_memory_MemAvailable_bytes",
		Type: v1.MetricTypeGauge,
		Help: "Memory information field MemAvailable_bytes",
	},
	{
		Name: "node_memory_MemTotal_bytes",
		Type: v1.MetricTypeGauge,
		Help: "Memory information field MemTotal_bytes",
	},
	{
		Name: "node_disk_read_bytes_total",
		Type: v1.MetricTypeCounter,
		Help: "The total number of bytes read successfully",
	},
	{
		Name: "node_disk_written_bytes_total",
		Type: v1.MetricTypeCounter,
		Help: "The total number of bytes written successfully",
	},
	{
		Name: "node_network_receive_bytes_total",
		Type: v1.MetricTypeCounter,
		Help: "Network device statistic receive_bytes",
	},
	{
		Name: "node_network_transmit_bytes_total",
		Type: v1.MetricTypeCounter,
		Help: "Network device statistic transmit_bytes",
	},
	{
		Name: "node_filesystem_size_bytes",
		Type: v1.MetricTypeGauge,
		Help: "Filesystem size in bytes",
	},
	{
		Name: "node_filesystem_free_bytes",
		Type: v1.MetricTypeGauge,
		Help: "Filesystem free space in bytes",
	},
	{
		Name: "node_load1",
		Type: v1.MetricTypeGauge,
		Help: "1m load average",
	},
	{
		Name: "node_load5",
		Type: v1.MetricTypeGauge,
		Help: "5m load average",
	},
	{
		Name: "node_load15",
		Type: v1.MetricTypeGauge,
		Help: "15m load average",
	},
}

// CPU 모드 정의
var cpuModes = []string{"idle", "user", "system", "iowait", "irq", "softirq", "steal", "guest", "guest_nice"}

type Options struct {
	NodeCount   int
	CPUPerNode  int
	DiskPerNode int
	EthPerNode  int
}

func GetMetrics(opts Options) ([]fakemetrics.Metric, error) {
	var metrics []fakemetrics.Metric

	if opts.NodeCount <= 0 {
		opts.NodeCount = 2
	}
	if opts.CPUPerNode <= 0 {
		opts.CPUPerNode = 2
	}
	if opts.DiskPerNode <= 0 {
		opts.DiskPerNode = 2
	}
	if opts.EthPerNode <= 0 {
		opts.EthPerNode = 2
	}

	for i := 0; i < opts.NodeCount; i++ {
		nodeName := fmt.Sprintf("node%d", i+1)

		for _, option := range metricOptions {
			switch option.Name {
			case "node_cpu_seconds_total":
				for cpu := 0; cpu < opts.CPUPerNode; cpu++ {
					for _, mode := range cpuModes {
						cpuOption := option
						if cpuOption.Labels == nil {
							cpuOption.Labels = make(map[string]string)
						}
						cpuOption.Labels["node"] = nodeName
						cpuOption.Labels["cpu"] = fmt.Sprintf("%d", cpu)
						cpuOption.Labels["mode"] = mode

						metric, err := fakemetrics.New(cpuOption)
						if err != nil {
							return nil, fmt.Errorf("failed to create CPU metric %s for %s cpu %d mode %s: %w",
								option.Name, nodeName, cpu, mode, err)
						}

						metrics = append(metrics, *metric)
					}
				}
			case "node_network_receive_bytes_total", "node_network_transmit_bytes_total":
				// 네트워크 메트릭: EthPerNode만큼 생성
				for eth := 0; eth < opts.EthPerNode; eth++ {
					nodeOption := option
					if nodeOption.Labels == nil {
						nodeOption.Labels = make(map[string]string)
					}
					nodeOption.Labels["node"] = nodeName
					nodeOption.Labels["device"] = fmt.Sprintf("eth%d", eth)

					metric, err := fakemetrics.New(nodeOption)
					if err != nil {
						return nil, fmt.Errorf("failed to create network metric %s for %s device eth%d: %w",
							option.Name, nodeName, eth, err)
					}

					metrics = append(metrics, *metric)
				}
			case "node_disk_read_bytes_total", "node_disk_written_bytes_total":
				for disk := 0; disk < opts.DiskPerNode; disk++ {
					nodeOption := option
					if nodeOption.Labels == nil {
						nodeOption.Labels = make(map[string]string)
					}
					nodeOption.Labels["node"] = nodeName
					nodeOption.Labels["device"] = fmt.Sprintf("sd%c", 'a'+disk)

					metric, err := fakemetrics.New(nodeOption)
					if err != nil {
						return nil, fmt.Errorf("failed to create disk metric %s for %s device sd%c: %w",
							option.Name, nodeName, 'a'+disk, err)
					}

					metrics = append(metrics, *metric)
				}
			case "node_filesystem_size_bytes", "node_filesystem_free_bytes":
				for disk := 0; disk < opts.DiskPerNode; disk++ {
					nodeOption := option
					if nodeOption.Labels == nil {
						nodeOption.Labels = make(map[string]string)
					}
					nodeOption.Labels["node"] = nodeName
					nodeOption.Labels["device"] = fmt.Sprintf("/dev/sd%c1", 'a'+disk)
					nodeOption.Labels["mountpoint"] = fmt.Sprintf("/mnt/disk%d", disk)
					nodeOption.Labels["fstype"] = "ext4"

					metric, err := fakemetrics.New(nodeOption)
					if err != nil {
						return nil, fmt.Errorf("failed to create filesystem metric %s for %s device /dev/sd%c1: %w",
							option.Name, nodeName, 'a'+disk, err)
					}

					metrics = append(metrics, *metric)
				}
			default:
				nodeOption := option
				if nodeOption.Labels == nil {
					nodeOption.Labels = make(map[string]string)
				}
				nodeOption.Labels["node"] = nodeName

				metric, err := fakemetrics.New(nodeOption)
				if err != nil {
					return nil, fmt.Errorf("failed to create metric %s for %s: %w",
						option.Name, nodeName, err)
				}

				metrics = append(metrics, *metric)
			}
		}
	}

	return metrics, nil
}
