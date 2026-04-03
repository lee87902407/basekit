package mempool

import (
	"bytes"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

type NopStats struct{}

func (NopStats) OnGet(size int, bucket int, pooled bool)     {}
func (NopStats) OnPut(capacity int, bucket int, pooled bool) {}
func (NopStats) OnDrop(capacity int, reason string)          {}

type PrometheusStats struct {
	registry          *prometheus.Registry
	bucketGets        *prometheus.CounterVec
	requestsTotal     prometheus.Counter
	releasesTotal     prometheus.Counter
	dropsTotal        *prometheus.CounterVec
	requestsPerSecond prometheus.Gauge
	currentSecond     atomic.Uint64
	stopCh            chan struct{}
}

func NewPrometheusStats() *PrometheusStats {
	registry := prometheus.NewRegistry()
	return NewPrometheusStatsWithRegistry(registry)
}

func NewPrometheusStatsWithRegistry(registry *prometheus.Registry) *PrometheusStats {
	stats := &PrometheusStats{
		registry: registry,
		bucketGets: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mempool_get_total",
				Help: "Total mempool get requests grouped by bucket and pooled source.",
			},
			[]string{"bucket", "pooled"},
		),
		requestsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "mempool_requests_total",
				Help: "Total number of mempool get requests.",
			},
		),
		releasesTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "mempool_releases_total",
				Help: "Total number of mempool release operations.",
			},
		),
		dropsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mempool_drop_total",
				Help: "Total number of dropped buffers grouped by reason.",
			},
			[]string{"reason"},
		),
		requestsPerSecond: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "mempool_requests_per_second",
				Help: "Approximate mempool request count observed during the current one-second window.",
			},
		),
		stopCh: make(chan struct{}),
	}

	registry.MustRegister(
		stats.bucketGets,
		stats.requestsTotal,
		stats.releasesTotal,
		stats.dropsTotal,
		stats.requestsPerSecond,
	)

	go stats.runRPSLoop()

	return stats
}

func (s *PrometheusStats) OnGet(size int, bucket int, pooled bool) {
	s.bucketGets.WithLabelValues(strconv.Itoa(bucket), strconv.FormatBool(pooled)).Inc()
	s.requestsTotal.Inc()
	count := s.currentSecond.Add(1)
	s.requestsPerSecond.Set(float64(count))
}

func (s *PrometheusStats) OnPut(capacity int, bucket int, pooled bool) {
	s.releasesTotal.Inc()
}

func (s *PrometheusStats) OnDrop(capacity int, reason string) {
	s.dropsTotal.WithLabelValues(reason).Inc()
}

func (s *PrometheusStats) Registry() *prometheus.Registry {
	return s.registry
}

func (s *PrometheusStats) GatherText() (string, error) {
	metrics, err := s.registry.Gather()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	for _, metric := range metrics {
		if _, err := expfmt.MetricFamilyToText(&buf, metric); err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

func (s *PrometheusStats) Close() {
	select {
	case <-s.stopCh:
		return
	default:
		close(s.stopCh)
	}
}

func (s *PrometheusStats) runRPSLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			count := s.currentSecond.Swap(0)
			s.requestsPerSecond.Set(float64(count))
		case <-s.stopCh:
			return
		}
	}
}
