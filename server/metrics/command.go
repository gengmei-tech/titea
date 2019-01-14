package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// CommandDuration command duration times
	CommandDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "gmkv",
			Subsystem: "command",
			Name:      "duration_time",
			Help:      " Duration time of command",
			Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 10),
		}, []string{"group", "service", "type"})
)

func init() {
	prometheus.MustRegister(CommandDuration)
}
