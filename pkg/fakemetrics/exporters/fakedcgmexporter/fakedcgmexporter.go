package fakedcgmexporter

import (
	"fmt"

	"github.com/kuoss/dummy-exporter/pkg/fakemetrics"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// https://github.com/NVIDIA/dcgm-exporter/blob/4.2.3-4.2.0/etc/default-counters.csv
// https://github.com/NVIDIA/dcgm-exporter/blob/4.2.3-4.2.0/grafana/dcgm-exporter-dashboard.json
// https://docs.nvidia.com/datacenter/dcgm/latest/dcgm-api/dcgm-api-field-ids.html

var metricOptions = []fakemetrics.Options{
	{
		Name: "DCGM_FI_DEV_SM_CLOCK",
		Type: v1.MetricTypeGauge,
		Help: "SM clock frequency (in MHz).",
	},
	{
		Name: "DCGM_FI_DEV_MEM_CLOCK",
		Type: v1.MetricTypeGauge,
		Help: "Memory clock frequency (in MHz).",
	},
	{
		Name: "DCGM_FI_DEV_MEMORY_TEMP",
		Type: v1.MetricTypeGauge,
		Help: "Memory temperature (in C).",
	},
	{
		Name: "DCGM_FI_DEV_GPU_TEMP",
		Type: v1.MetricTypeGauge,
		Help: "GPU temperature (in C).",
	},
	{
		Name: "DCGM_FI_DEV_POWER_USAGE",
		Type: v1.MetricTypeGauge,
		Help: "Power draw (in W).",
	},
	{
		Name: "DCGM_FI_DEV_TOTAL_ENERGY_CONSUMPTION",
		Type: v1.MetricTypeCounter,
		Help: "Total energy consumption since boot (in mJ).",
	},
	{
		Name: "DCGM_FI_DEV_PCIE_REPLAY_COUNTER",
		Type: v1.MetricTypeCounter,
		Help: "Total number of PCIe retries.",
	},
	{
		Name: "DCGM_FI_DEV_GPU_UTIL",
		Type: v1.MetricTypeGauge,
		Help: "GPU utilization (in %).",
	},
	{
		Name: "DCGM_FI_DEV_MEM_COPY_UTIL",
		Type: v1.MetricTypeGauge,
		Help: "Memory utilization (in %).",
	},
	{
		Name: "DCGM_FI_DEV_ENC_UTIL",
		Type: v1.MetricTypeGauge,
		Help: "Encoder utilization (in %).",
	},
	{
		Name: "DCGM_FI_DEV_DEC_UTIL",
		Type: v1.MetricTypeGauge,
		Help: "Decoder utilization (in %).",
	},
	{
		Name: "DCGM_FI_DEV_XID_ERRORS",
		Type: v1.MetricTypeGauge,
		Help: "Value of the last XID error encountered.",
	},
	{
		Name: "DCGM_FI_DEV_FB_FREE",
		Type: v1.MetricTypeGauge,
		Help: "Framebuffer memory free (in MiB).",
	},
	{
		Name: "DCGM_FI_DEV_FB_USED",
		Type: v1.MetricTypeGauge,
		Help: "Framebuffer memory used (in MiB).",
	},
	{
		Name: "DCGM_FI_DEV_NVLINK_BANDWIDTH_TOTAL",
		Type: v1.MetricTypeCounter,
		Help: "Total number of NVLink bandwidth counters for all lanes.",
	},
	{
		Name: "DCGM_FI_DEV_VGPU_LICENSE_STATUS",
		Type: v1.MetricTypeGauge,
		Help: "vGPU License status.",
	},
	{
		Name: "DCGM_FI_DEV_UNCORRECTABLE_REMAPPED_ROWS",
		Type: v1.MetricTypeCounter,
		Help: "Number of remapped rows for uncorrectable errors.",
	},
	{
		Name: "DCGM_FI_DEV_CORRECTABLE_REMAPPED_ROWS",
		Type: v1.MetricTypeCounter,
		Help: "Number of remapped rows for correctable errors.",
	},
	{
		Name: "DCGM_FI_DEV_ROW_REMAP_FAILURE",
		Type: v1.MetricTypeGauge,
		Help: "Whether remapping of rows has failed.",
	},
	{
		Name: "DCGM_FI_PROF_GR_ENGINE_ACTIVE",
		Type: v1.MetricTypeGauge,
		Help: "Ratio of time the graphics engine is active.",
	},
	{
		Name: "DCGM_FI_PROF_PIPE_TENSOR_ACTIVE",
		Type: v1.MetricTypeGauge,
		Help: "Ratio of cycles the tensor (HMMA) pipe is active.",
	},
	{
		Name: "DCGM_FI_PROF_DRAM_ACTIVE",
		Type: v1.MetricTypeGauge,
		Help: "Ratio of cycles the device memory interface is active sending or receiving data.",
	},
	{
		Name: "DCGM_FI_PROF_PCIE_TX_BYTES",
		Type: v1.MetricTypeGauge,
		Help: "The rate of data transmitted over PCIe in bytes per second.",
	},
	{
		Name: "DCGM_FI_PROF_PCIE_RX_BYTES",
		Type: v1.MetricTypeGauge,
		Help: "The rate of data received over PCIe in bytes per second.",
	},
}

type Options struct {
	NodeCount  int
	GPUPerNode int
}

func GetMetrics(opts Options) ([]fakemetrics.Metric, error) {
	var metrics []fakemetrics.Metric

	// Set default values
	if opts.NodeCount <= 0 {
		opts.NodeCount = 1
	}
	if opts.GPUPerNode <= 0 {
		opts.GPUPerNode = 1
	}

	for i := 0; i < opts.NodeCount; i++ {
		nodeName := fmt.Sprintf("node%d", i+1)

		for g := 0; g < opts.GPUPerNode; g++ {
			gpuName := fmt.Sprintf("gpu%d", g+1)

			for _, option := range metricOptions {
				gpuOption := option
				if gpuOption.Labels == nil {
					gpuOption.Labels = make(map[string]string)
				}
				gpuOption.Labels["node"] = nodeName
				gpuOption.Labels["gpu"] = gpuName

				metric, err := fakemetrics.New(gpuOption)
				if err != nil {
					return nil, fmt.Errorf("failed to create DCGM metric %s for %s %s: %w",
						option.Name, nodeName, gpuName, err)
				}

				metrics = append(metrics, *metric)
			}
		}
	}

	return metrics, nil
}
