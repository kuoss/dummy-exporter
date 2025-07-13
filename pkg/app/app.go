package app

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kuoss/dummy-exporter/pkg/config"
	"github.com/kuoss/dummy-exporter/pkg/fakemetrics"
	"github.com/kuoss/dummy-exporter/pkg/fakemetrics/exporters/fakedcgmexporter"
	"github.com/kuoss/dummy-exporter/pkg/fakemetrics/exporters/fakekubestatemetrics"
	"github.com/kuoss/dummy-exporter/pkg/fakemetrics/exporters/fakenodeexporter"
)

func Start(version string) error {
	log.Printf("🚀 Starting dummy-exporter version %s", version)

	cfg, err := config.LoadConfig("etc/exporter.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	printConfigSummary(cfg)

	metrics, err := getMetrics(cfg)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	updateInterval := 3 * time.Second
	if cfg.Server.UpdateInterval != "" {
		if d, err := time.ParseDuration(cfg.Server.UpdateInterval); err == nil {
			updateInterval = d
		} else {
			log.Printf("⚠️  Invalid updateInterval %q, defaulting to 3s", cfg.Server.UpdateInterval)
		}
	}
	log.Printf("⏱️  Update interval: %s", updateInterval)

	// Create a new registry instead of using the default global registry
	reg := prometheus.NewRegistry()

	// Register built-in collectors from config
	if cfg.Collectors.Enabled {
		if cfg.Collectors.GoCollector {
			if err := reg.Register(collectors.NewGoCollector()); err != nil {
				log.Printf("⚠️  Failed to register GoCollector: %v", err)
			} else {
				log.Printf("✅ Registered GoCollector")
			}
		}

		if cfg.Collectors.ProcessCollector {
			if err := reg.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
				log.Printf("⚠️  Failed to register ProcessCollector: %v", err)
			} else {
				log.Printf("✅ Registered ProcessCollector")
			}
		}
	} else {
		log.Printf("🚫 Collectors disabled by config")
	}

	// Register metrics with duplicate checking
	for _, m := range metrics {
		collector := m.Collector()

		if err := reg.Register(collector); err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
				log.Printf("⚠️  Metric already registered, skipping")
				continue
			}
			return fmt.Errorf("failed to register metric: %w", err)
		}

		log.Printf("✅ Registered metric: %s (%s)", m.Name, m.Type)
	}

	// Set up HTTP handler with the custom registry
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		Registry: reg,
	}))

	// Add health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// Add status endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html>
<head><title>Dummy Exporter</title></head>
<body>
<h1>Dummy Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
<p><a href="/health">Health</a></p>
<h2>Configuration</h2>
<ul>
<li>Node Exporter: %+v</li>
<li>DCGM Exporter: %+v</li>
<li>KSM Exporter: %+v</li>
<li>Custom Metrics: %d</li>
</ul>
<p>Total Metrics: %d</p>
<p>Update Interval: %s</p>
</body>
</html>
`,
			cfg.Exporters.NodeExporter,
			cfg.Exporters.DcgmExporter,
			cfg.Exporters.KubeStateMetrics,
			len(cfg.CustomMetrics),
			len(metrics),
			updateInterval)
	})

	// Set initial values for all metrics before starting update loop
	for _, m := range metrics {
		m.Update()
	}

	// Start the update goroutine
	go func() {
		ticker := time.NewTicker(updateInterval)
		defer ticker.Stop()

		log.Printf("📊 Starting metrics update loop with %d metrics", len(metrics))

		for {
			select {
			case <-ticker.C:
				for _, m := range metrics {
					m.Update()
				}
			}
		}
	}()

	addr := fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port)
	log.Printf("🔊 Listening on %s", addr)
	log.Printf("📈 Metrics endpoint: http://%s/metrics", addr)

	return http.ListenAndServe(addr, nil)
}

func printConfigSummary(cfg *config.Config) {
	log.Println("📄 Loaded config:")
	log.Printf("  Listen:      %s:%d", cfg.Server.Address, cfg.Server.Port)
	log.Printf("  Update:      %s", cfg.Server.UpdateInterval)
	log.Printf("  Exporters:")
	log.Printf("    dcgmExporter:       %v", cfg.Exporters.DcgmExporter.Enabled)
	if cfg.Exporters.DcgmExporter.Enabled {
		log.Printf("      nodes: %d, gpuPerNode: %d",
			cfg.Exporters.DcgmExporter.NodeCount,
			cfg.Exporters.DcgmExporter.GPUPerNode)
	}
	log.Printf("    kubeStateMetrics:   %v", cfg.Exporters.KubeStateMetrics.Enabled)
	if cfg.Exporters.KubeStateMetrics.Enabled {
		log.Printf("      namespaces: %d, podsPerNs: %d, containersPerPod: %d",
			cfg.Exporters.KubeStateMetrics.NamespaceCount,
			cfg.Exporters.KubeStateMetrics.PodPerNamespace,
			cfg.Exporters.KubeStateMetrics.ContainerPerPod)
	}
	log.Printf("    nodeExporter:       %v", cfg.Exporters.NodeExporter.Enabled)
	if cfg.Exporters.NodeExporter.Enabled {
		log.Printf("      nodes: %d", cfg.Exporters.NodeExporter.NodeCount)
	}
	log.Printf("  Custom metrics: %d", len(cfg.CustomMetrics))
}

func getMetrics(cfg *config.Config) ([]fakemetrics.Metric, error) {
	var metrics []fakemetrics.Metric

	// Get metrics from built-in exporters
	if cfg.Exporters.Enabled {
		if cfg.Exporters.DcgmExporter.Enabled {
			log.Printf("🔧 Loading DCGM exporter metrics...")
			ms, err := fakedcgmexporter.GetMetrics(
				cfg.Exporters.DcgmExporter.NodeCount,
				cfg.Exporters.DcgmExporter.GPUPerNode,
			)
			if err != nil {
				return nil, fmt.Errorf("fakedcgmexporter.GetMetrics err: %w", err)
			}
			metrics = append(metrics, ms...)
			log.Printf("✅ Loaded %d DCGM metrics", len(ms))
		}

		if cfg.Exporters.KubeStateMetrics.Enabled {
			log.Printf("🔧 Loading KSM exporter metrics...")
			ms, err := fakekubestatemetrics.GetMetrics(
				cfg.Exporters.KubeStateMetrics.NamespaceCount,
				cfg.Exporters.KubeStateMetrics.PodPerNamespace,
				cfg.Exporters.KubeStateMetrics.ContainerPerPod,
			)
			if err != nil {
				return nil, fmt.Errorf("fakekubestatemetrics.GetMetrics err: %w", err)
			}
			metrics = append(metrics, ms...)
			log.Printf("✅ Loaded %d KSM metrics", len(ms))
		}

		if cfg.Exporters.NodeExporter.Enabled {
			log.Printf("🔧 Loading Node exporter metrics...")
			ms, err := fakenodeexporter.GetMetrics(fakenodeexporter.Options{
				NodeCount:   cfg.Exporters.NodeExporter.NodeCount,
				CPUPerNode:  cfg.Exporters.NodeExporter.CPUPerNode,
				DiskPerNode: cfg.Exporters.NodeExporter.DiskPerNode,
				EthPerNode:  cfg.Exporters.NodeExporter.EthPerNode,
			})
			if err != nil {
				return nil, fmt.Errorf("fakenodeexporter.GetMetrics err: %w", err)
			}
			metrics = append(metrics, ms...)
			log.Printf("✅ Loaded %d Node exporter metrics", len(ms))
		}
	}

	// Add custom metrics from config
	if len(cfg.CustomMetrics) > 0 {
		log.Printf("🔧 Loading custom metrics...")
		for i, o := range cfg.CustomMetrics {
			m, err := fakemetrics.New(o)
			if err != nil {
				return nil, fmt.Errorf("fakemetrics.New err for custom metric %d: %w", i, err)
			}
			metrics = append(metrics, *m)
		}
		log.Printf("✅ Loaded %d custom metrics", len(cfg.CustomMetrics))
	}

	log.Printf("📊 Total unique metrics loaded: %d", len(metrics))
	return metrics, nil
}
