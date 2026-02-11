package app

import "github.com/prometheus/client_golang/prometheus"

// Metrics captures gateway operational telemetry.
type Metrics struct {
	RequestsTotal *prometheus.CounterVec
	LatencySec    *prometheus.HistogramVec
	TokensTotal   *prometheus.CounterVec
	CostTotalUSD  *prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_requests_total",
				Help: "Total requests processed by the gateway.",
			},
			[]string{"team", "model", "status"},
		),
		LatencySec: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_request_latency_seconds",
				Help:    "Latency distribution for gateway requests.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"team", "model", "status"},
		),
		TokensTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_tokens_total",
				Help: "Token usage grouped by team/model/type.",
			},
			[]string{"team", "model", "type"},
		),
		CostTotalUSD: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_cost_usd_total",
				Help: "Estimated gateway cost (USD).",
			},
			[]string{"team", "model"},
		),
	}

	reg.MustRegister(
		m.RequestsTotal,
		m.LatencySec,
		m.TokensTotal,
		m.CostTotalUSD,
	)
	return m
}
