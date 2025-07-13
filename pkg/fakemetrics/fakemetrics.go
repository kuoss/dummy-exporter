package fakemetrics

import (
	"fmt"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kuoss/dummy-exporter/pkg/fakemetrics/generator"
)

type Metric struct {
	Name          string
	Type          v1.MetricType
	Help          string
	Labels        map[string]string `yaml:"labels,omitempty"`
	LabelNames    []string          `yaml:"label_names,omitempty"`
	PromCollector prometheus.Collector
	Generator     generator.Generator
	IsVector      bool
}

type Options struct {
	Name       string            `yaml:"name"`
	Type       v1.MetricType     `yaml:"type"`
	Help       string            `yaml:"help"`
	Generator  generator.Options `yaml:"generator,omitempty"`
	Labels     map[string]string `yaml:"labels,omitempty"`
	LabelNames []string          `yaml:"label_names,omitempty"`
}

func New(opts Options) (*Metric, error) {
	isVector := len(opts.LabelNames) > 0

	var pc prometheus.Collector
	switch opts.Type {
	case v1.MetricTypeCounter:
		if isVector {
			pc = prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: opts.Name,
				Help: opts.Help,
			}, opts.LabelNames)
		} else {
			pc = prometheus.NewCounter(prometheus.CounterOpts{
				Name:        opts.Name,
				Help:        opts.Help,
				ConstLabels: opts.Labels,
			})
		}
	case v1.MetricTypeGauge:
		if isVector {
			pc = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: opts.Name,
				Help: opts.Help,
			}, opts.LabelNames)
		} else {
			pc = prometheus.NewGauge(prometheus.GaugeOpts{
				Name:        opts.Name,
				Help:        opts.Help,
				ConstLabels: opts.Labels,
			})
		}
	case v1.MetricTypeHistogram:
		if isVector {
			pc = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name:    opts.Name,
				Help:    opts.Help,
				Buckets: prometheus.LinearBuckets(0.1, 0.1, 5),
			}, opts.LabelNames)
		} else {
			pc = prometheus.NewHistogram(prometheus.HistogramOpts{
				Name:        opts.Name,
				Help:        opts.Help,
				Buckets:     prometheus.LinearBuckets(0.1, 0.1, 5),
				ConstLabels: opts.Labels,
			})
		}
	case v1.MetricTypeSummary:
		if isVector {
			pc = prometheus.NewSummaryVec(prometheus.SummaryOpts{
				Name:       opts.Name,
				Help:       opts.Help,
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			}, opts.LabelNames)
		} else {
			pc = prometheus.NewSummary(prometheus.SummaryOpts{
				Name:        opts.Name,
				Help:        opts.Help,
				Objectives:  map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
				ConstLabels: opts.Labels,
			})
		}
	default:
		return nil, fmt.Errorf("unsupported metric type: %v", opts.Type)
	}

	gen := generator.New(opts.Type, opts.Generator)

	return &Metric{
		Name:          opts.Name,
		Help:          opts.Help,
		Type:          opts.Type,
		Labels:        opts.Labels,
		LabelNames:    opts.LabelNames,
		PromCollector: pc,
		Generator:     gen,
		IsVector:      isVector,
	}, nil
}

func (m *Metric) Update() {
	if m.Generator != nil {
		m.Generator.Update(m.PromCollector)
	}
}

func (m *Metric) Collector() prometheus.Collector {
	if m.IsVector && len(m.LabelNames) > 0 && len(m.Labels) > 0 {
		labelVals := []string{}
		for _, name := range m.LabelNames {
			labelVals = append(labelVals, m.Labels[name])
		}
		switch v := m.PromCollector.(type) {
		case *prometheus.GaugeVec:
			return v.WithLabelValues(labelVals...)
		case *prometheus.CounterVec:
			return v.WithLabelValues(labelVals...)
			// HistogramVec 등도 가능
		}
	}
	return m.PromCollector
}
