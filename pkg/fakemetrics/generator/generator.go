package generator

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
)

type Generator interface {
	Update(collector prometheus.Collector)
}

type Options struct {
	Type   string    `yaml:"type,omitempty"`   // fixed, inc, rand_float, etc.
	Value  float64   `yaml:"value,omitempty"`  // for fixed, inc (step)
	Values []float64 `yaml:"values,omitempty"` // for random, round_robin
}

type Type string

const (
	TypeFixed      Type = "fixed"
	TypeInc        Type = "inc"
	TypeRandFloat  Type = "rand_float"
	TypeRandInt    Type = "rand_int"
	TypeRandSample Type = "rand_sample"
	TypeRoundRobin Type = "round_robin"
)

// ========== FixedGenerator ==========
type FixedGenerator struct {
	Value float64
}

func (g *FixedGenerator) Update(collector prometheus.Collector) {
	if m, ok := collector.(interface{ Set(float64) }); ok {
		m.Set(g.Value)
	}
}

// ========== IncGenerator ==========
type IncGenerator struct {
	Counter float64
	Step    float64
	mu      sync.Mutex
}

func (g *IncGenerator) Update(collector prometheus.Collector) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Counter += g.Step
	if m, ok := collector.(interface{ Set(float64) }); ok {
		m.Set(g.Counter)
	}
}

// ========== RandFloatGenerator ==========
type RandFloatGenerator struct {
	Min float64
	Max float64
	rng *rand.Rand
}

func (g *RandFloatGenerator) Update(collector prometheus.Collector) {
	val := g.Min + g.rng.Float64()*(g.Max-g.Min)
	if m, ok := collector.(interface{ Set(float64) }); ok {
		m.Set(val)
	}
}

// ========== RandIntGenerator ==========
type RandIntGenerator struct {
	Min int
	Max int
	rng *rand.Rand
}

func (g *RandIntGenerator) Update(collector prometheus.Collector) {
	val := float64(g.rng.Intn(g.Max-g.Min+1) + g.Min)
	if m, ok := collector.(interface{ Set(float64) }); ok {
		m.Set(val)
	}
}

// ========== RandSampleGenerator ==========
type RandSampleGenerator struct {
	Min float64
	Max float64
	rng *rand.Rand
}

func (g *RandSampleGenerator) Update(collector prometheus.Collector) {
	val := g.Min + g.rng.Float64()*(g.Max-g.Min)
	if m, ok := collector.(interface{ Observe(float64) }); ok {
		m.Observe(val)
	}
}

// ========== RoundRobinGenerator ==========
type RoundRobinGenerator struct {
	Values []float64
	index  int
	mu     sync.Mutex
}

func (g *RoundRobinGenerator) Update(collector prometheus.Collector) {
	if len(g.Values) == 0 {
		return
	}
	g.mu.Lock()
	val := g.Values[g.index]
	g.index = (g.index + 1) % len(g.Values)
	g.mu.Unlock()

	if m, ok := collector.(interface{ Set(float64) }); ok {
		m.Set(val)
	}
}

// ========== Generator Factory ==========
func New(mt v1.MetricType, opts Options) Generator {
	typ := Type(opts.Type)
	value := opts.Value
	values := opts.Values

	// Infer type if not specified
	if typ == "" {
		switch mt {
		case v1.MetricTypeCounter:
			typ = TypeInc
		case v1.MetricTypeGauge:
			typ = TypeRandFloat
		case v1.MetricTypeHistogram, v1.MetricTypeSummary:
			typ = TypeRandSample
		default:
			typ = TypeFixed
		}
	}

	// Apply defaults
	switch typ {
	case TypeInc:
		if value == 0 {
			value = 1
		}
	case TypeRandFloat, TypeRandInt, TypeRandSample:
		if len(values) != 2 {
			values = []float64{0, 100}
		}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	switch typ {
	case TypeFixed:
		return &FixedGenerator{Value: value}
	case TypeInc:
		return &IncGenerator{Step: value}
	case TypeRandFloat:
		return &RandFloatGenerator{Min: values[0], Max: values[1], rng: rng}
	case TypeRandInt:
		return &RandIntGenerator{Min: int(values[0]), Max: int(values[1]), rng: rng}
	case TypeRandSample:
		return &RandSampleGenerator{Min: values[0], Max: values[1], rng: rng}
	case TypeRoundRobin:
		return &RoundRobinGenerator{Values: values}
	default:
		panic(fmt.Sprintf("unknown generator type: %s", typ))
	}
}
