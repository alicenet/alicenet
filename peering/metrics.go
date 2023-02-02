package peering

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricLabelState = "state"

	metricStateActive   = "active"
	metricStateInactive = "inactive"
)

// Peers manages peer metrics
type Peers struct {
	peerCount *prometheus.GaugeVec
}

// NewPeerMetrics takes in a prometheus registry and initializes
// and registers relay metrics. It returns those registered Peers.
func NewPeerMetrics(r prometheus.Registerer) *Peers {
	return &Peers{
		peerCount: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "peer_count",
				Help: "the total count of peers connected and their state",
			}, []string{metricLabelState}),
	}
}
