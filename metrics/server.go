package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alicenet/alicenet/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultPath = "/metrics"
)

type MetricServer struct {
	promPort uint32
	registry *prometheus.Registry
	server   http.Server
}

func NewServer(promPort uint16, registry *prometheus.Registry) *MetricServer {
	if registry == nil {
		registry = prometheus.NewRegistry()
	}
	m := http.NewServeMux()
	m.Handle(defaultPath, promhttp.Handler())
	logging.GetLogger("metrics").
		WithField("path", defaultPath).
		WithField("msg", "initialising prometheus metrics")
	return &MetricServer{
		server: http.Server{
			Addr:    fmt.Sprintf(":%d", promPort),
			Handler: m,
		},
		registry: registry,
	}
}

func (s *MetricServer) Start() {
	logging.GetLogger("metrics").
		WithField("path", defaultPath).
		WithField("msg", "starting prometheus metric server")
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logging.GetLogger("metrics").
			WithField("error", err.Error()).
			WithField("msg", "exiting prometheus metric server")
	}
}

func (s *MetricServer) Stop(ctx context.Context) error {
	logging.GetLogger("metrics").
		WithField("msg", "exiting prometheus metric server")
	err := s.server.Shutdown(ctx)
	if err != nil {
		logging.GetLogger("metrics").
			WithField("msg", "failed to exit prometheus metric server cleanly")
		return err
	}
	logging.GetLogger("metrics").
		WithField("msg", "successfully exited prometheus metric server")
	return nil
}
