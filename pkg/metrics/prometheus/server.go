/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package prometheus

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	Port int
	Host string

	stopped bool
	server  *http.Server
}

func NewServer(host string, port int) *Server {
	return &Server{
		Port: port,
		Host: host,
	}
}

func (prmServ *Server) Run() {
	if prmServ.server == nil || prmServ.stopped {
		prmServ.server = &http.Server{Addr: fmt.Sprintf("%s:%d", prmServ.Host, prmServ.Port)}
		http.Handle("/metrics", promhttp.Handler())
		go prmServ.server.ListenAndServe()
	}
}

func (prmServ *Server) Stop() error {
	if prmServ.server != nil && !prmServ.stopped {
		return prmServ.server.Shutdown(context.Background())
	}
	return nil
}
