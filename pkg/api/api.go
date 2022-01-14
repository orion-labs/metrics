package api

import (
	"github.com/orion-labs/metrics/pkg/server"
	"github.com/segmentio/stats/v4"
	"github.com/segmentio/stats/v4/prometheus"
	log "github.com/sirupsen/logrus"
	"time"
)

const DefaultStatsPort = 7418
var DefaultStatsHandler = &prometheus.Handler{}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

type MetricsConfig struct {
	Prefix    string `json:"prefix"`
	StatsPort int    `json:"stats_port"`
}

type Metrics struct {
	Config      *MetricsConfig
	StatsHandler *prometheus.Handler
	MetricsSvr  *metrics.Server
	MetricsEngine *stats.Engine
}

func NewDefaultMetrics() *Metrics {
	return NewMetrics(DefaultStatsPort, DefaultStatsHandler, NewStatsEngine(""))
}

func NewMetrics(port int, h *prometheus.Handler, eng *stats.Engine) *Metrics {
	return &Metrics{
		Config:       &MetricsConfig{
			Prefix:    eng.Prefix,
			StatsPort: port,
		},
		StatsHandler: h,
		MetricsEngine: eng,
		MetricsSvr:   NewMetricsServer(h, port, eng),
	}
}

func NewMetricsServer(h *prometheus.Handler, port int, eng *stats.Engine) *metrics.Server {
	return metrics.NewPrometheusMetricServer(port, h, eng)
}

func NewStatsEngine(prefix string, tags... stats.Tag) *stats.Engine {
	return stats.NewEngine(
		prefix,
		nil,
		tags...
	)
}

func (m *Metrics) WithTags(tags []stats.Tag) *Metrics {
	if m.MetricsEngine != nil {
		prefix := m.MetricsEngine.Prefix
		m.MetricsEngine = m.MetricsEngine.WithPrefix(prefix, tags...)
		m.MetricsSvr.StatsEngine = m.MetricsEngine
		stats.DefaultEngine = m.MetricsEngine
	}
	return m
}

func (m *Metrics) Run() error {
	if m.MetricsSvr != nil {
		svr := m.MetricsSvr
		*m.MetricsEngine = *m.MetricsSvr.StatsEngine
		m.MetricsSvr = NewMetricsServer(m.StatsHandler, m.Config.StatsPort, m.MetricsEngine)
		_ = svr.Close()
	}

	m.startupStats()
	return m.MetricsSvr.Run()
}

func (m *Metrics) Close() error {
	if m.MetricsSvr != nil {
		return m.MetricsSvr.Close()
	}
	return nil
}

func (m *Metrics) Flush() {
	if m.MetricsEngine != nil {
		m.MetricsEngine.Flush()
	}
}

func (m *Metrics) startupStats() {
	defer stats.Flush()
	now := time.Now()
	stats.Set("start.time", now.Sub(time.Time{}).Seconds())
}

