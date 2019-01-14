package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// ConnGauge connection count
	ConnGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "gmkv",
			Subsystem: "server",
			Name:      "connections",
			Help:      "Number of connections.",
		})
	// GcAddCounter GC add
	GcAddCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gmkv",
			Subsystem: "store",
			Name:      "gc_add",
			Help:      "Gauge of jobs.",
		}, []string{"group", "service", "type"})

	// GcExecCounter GC exec
	GcExecCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gmkv",
			Subsystem: "store",
			Name:      "gc_exec",
			Help:      "Gauge of jobs.",
		}, []string{"group", "service", "type"})
	// ExpireExecCounter expire exec
	ExpireExecCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gmkv",
			Subsystem: "store",
			Name:      "expire_exec",
			Help:      "Gauge of jobs.",
		}, []string{"group", "service", "type"})

	// TxnFailedCounter txn failed count
	TxnFailedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gmkv",
			Subsystem: "store",
			Name:      "txn_failed_total",
			Help:      "Gauge of jobs.",
		}, []string{"group", "service", "type", "event"})
)

func init() {
	prometheus.MustRegister(ConnGauge)
	prometheus.MustRegister(GcAddCounter)
	prometheus.MustRegister(GcExecCounter)
	prometheus.MustRegister(ExpireExecCounter)
	prometheus.MustRegister(TxnFailedCounter)
}
