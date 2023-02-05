package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusServer struct {
	Port string
	Host string

	stopped bool
	server  *http.Server
}

func NewPrometheusServer(host string, port string) *PrometheusServer {
	return &PrometheusServer{
		Port: port,
		Host: host,
	}
}

func (prmServ *PrometheusServer) Run() {
	if prmServ.server == nil || prmServ.stopped {
		prmServ.server = &http.Server{Addr: fmt.Sprintf("%v:%v", prmServ.Host, prmServ.Port)}
		http.Handle("/metrics", promhttp.Handler())
		go prmServ.server.ListenAndServe()
	}
}

func (prmServ *PrometheusServer) Stop() {
	if prmServ.server != nil && !prmServ.stopped {
		prmServ.server.Shutdown(context.Background())
	}
}
