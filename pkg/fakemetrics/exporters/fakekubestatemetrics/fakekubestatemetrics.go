package fakekubestatemetrics

import (
	"fmt"

	"github.com/kuoss/dummy-exporter/pkg/fakemetrics"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

var metricOptions = []fakemetrics.Options{
	{
		Name: "kube_node_info",
		Type: v1.MetricTypeGauge,
		Help: "Information about a node",
	},
	{
		Name: "kube_node_status_condition",
		Type: v1.MetricTypeGauge,
		Help: "The condition of a node (Ready, MemoryPressure, etc.)",
	},
	{
		Name: "kube_pod_info",
		Type: v1.MetricTypeGauge,
		Help: "Information about a pod",
	},
	{
		Name: "kube_pod_status_phase",
		Type: v1.MetricTypeGauge,
		Help: "The pods current phase",
	},
	{
		Name: "kube_pod_container_status_ready",
		Type: v1.MetricTypeGauge,
		Help: "Describes whether the container is ready to serve requests",
	},
	{
		Name: "kube_pod_container_status_restarts_total",
		Type: v1.MetricTypeCounter,
		Help: "Container restarts total",
	},
	{
		Name: "kube_deployment_status_replicas_available",
		Type: v1.MetricTypeGauge,
		Help: "The number of available replicas per deployment",
	},
	{
		Name: "kube_deployment_spec_replicas",
		Type: v1.MetricTypeGauge,
		Help: "Number of desired pods for a deployment",
	},
	{
		Name: "kube_deployment_status_replicas_unavailable",
		Type: v1.MetricTypeGauge,
		Help: "The number of unavailable replicas per deployment",
	},
	{
		Name: "kube_statefulset_status_replicas_ready",
		Type: v1.MetricTypeGauge,
		Help: "The number of ready replicas in a StatefulSet",
	},
}

type Options struct {
	NamespaceCount  int
	PodPerNamespace int
	ContainerPerPod int
}

func GetMetrics(opts Options) ([]fakemetrics.Metric, error) {
	var metrics []fakemetrics.Metric

	if opts.NamespaceCount <= 0 {
		opts.NamespaceCount = 2
	}
	if opts.PodPerNamespace <= 0 {
		opts.PodPerNamespace = 2
	}
	if opts.ContainerPerPod <= 0 {
		opts.ContainerPerPod = 2
	}

	for ns := 0; ns < opts.NamespaceCount; ns++ {
		namespaceName := fmt.Sprintf("namespace%d", ns+1)

		for pod := 0; pod < opts.PodPerNamespace; pod++ {
			podName := fmt.Sprintf("pod%d", pod+1)

			for _, option := range metricOptions {
				// Create metrics for each pod
				if option.Name == "kube_pod_status_phase" {
					podOption := option
					if podOption.Labels == nil {
						podOption.Labels = make(map[string]string)
					}
					podOption.Labels["namespace"] = namespaceName
					podOption.Labels["pod"] = podName

					metric, err := fakemetrics.New(podOption)
					if err != nil {
						return nil, fmt.Errorf("failed to create pod metric %s for %s in namespace %s: %w",
							option.Name, podName, namespaceName, err)
					}

					metrics = append(metrics, *metric)
				}
			}
		}

		// Create node condition metrics
		for _, option := range metricOptions {
			switch option.Name {
			case "kube_node_status_condition":
				nodeOption := option
				if nodeOption.Labels == nil {
					nodeOption.Labels = make(map[string]string)
				}
				nodeOption.Labels["namespace"] = namespaceName

				metric, err := fakemetrics.New(nodeOption)
				if err != nil {
					return nil, fmt.Errorf("failed to create node condition metric %s for namespace %s: %w",
						option.Name, namespaceName, err)
				}

				metrics = append(metrics, *metric)
			case "kube_deployment_status_replicas":
				// Deployment metrics
				deploymentOption := option
				if deploymentOption.Labels == nil {
					deploymentOption.Labels = make(map[string]string)
				}
				deploymentOption.Labels["namespace"] = namespaceName

				metric, err := fakemetrics.New(deploymentOption)
				if err != nil {
					return nil, fmt.Errorf("failed to create deployment metric %s for namespace %s: %w",
						option.Name, namespaceName, err)
				}

				metrics = append(metrics, *metric)
			}
		}
	}

	return metrics, nil
}
