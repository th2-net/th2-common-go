package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusServer struct {
	Port string
	Host string

	stopped bool
	server  *http.Server
}

func (prmServ *PrometheusServer) Run() {
	if prmServ.server == nil || prmServ.stopped {
		prmServ.server = &http.Server{Addr: ":2112"}
		http.Handle("/metrics", promhttp.Handler())
		go prmServ.server.ListenAndServe()
	}
}

func (prmServ *PrometheusServer) Stop() {
	if prmServ.server != nil && !prmServ.stopped {
		prmServ.server.Shutdown(context.Background())
	}
}
