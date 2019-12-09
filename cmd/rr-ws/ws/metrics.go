package ws

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	ws "github.com/spiral/broadcast-ws"
	rr "github.com/spiral/roadrunner/cmd/rr/cmd"
	"github.com/spiral/roadrunner/service/metrics"
)

func init() {
	cobra.OnInitialize(func() {
		svc, _ := rr.Container.Get(metrics.ID)
		mtr, ok := svc.(*metrics.Service)
		if !ok || !mtr.Enabled() {
			return
		}

		ht, _ := rr.Container.Get(ws.ID)
		if bc, ok := ht.(*ws.Service); ok {
			collector := newCollector()

			// register metrics
			mtr.MustRegister(collector.connCounter)

			// collect events
			bc.AddListener(collector.listener)
		}
	})
}

// listener provide debug callback for system events. With colors!
type metricCollector struct {
	connCounter prometheus.Gauge
}

func newCollector() *metricCollector {
	return &metricCollector{
		connCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "rr_ws_conn_total",
				Help: "Total number of websocket connections.",
			},
		),
	}
}

// listener listens to http events and generates nice looking output.
func (c *metricCollector) listener(event int, ctx interface{}) {
	switch event {
	case ws.EventConnect:
		c.connCounter.Inc()
	case ws.EventDisconnect:
		c.connCounter.Dec()
	}
}
