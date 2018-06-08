package consumer

import (
	"net/http"
	"fmt"
	log "github.com/sirupsen/logrus"
	"encoding/json"
)

type Status struct{
	Connected bool
	LatestErrors []*Node
}

type HealthCheckConfig struct{
	Port uint
}

type HealthCheckServer struct{
	cfg *HealthCheckConfig
	statusFunc func() bool
	errorChannel ErrorChannel
}

func NewHealthCheckServer(cfg *HealthCheckConfig, health func() bool, ec ErrorChannel) *HealthCheckServer {
	return &HealthCheckServer{
		cfg:cfg,
		statusFunc:health,
		errorChannel:ec,
	}
}

func (svr *HealthCheckServer) SetupRoutes() {
	log.Infof("setting up healthcheck routes")
	http.Handle("/health", http.HandlerFunc(svr.GetStatus))
}

func (svr *HealthCheckServer) Run() {
	log.Infof("starting healthCheck server on port %v", svr.cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v",svr.cfg.Port), svr.setHeaders(http.DefaultServeMux)))
	log.Infof("exiting healthCheck server", svr.cfg.Port)
}

func(svr *HealthCheckServer) setHeaders (h http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		w.Header().Add("accept", "application/json")
		h.ServeHTTP(w, r)
	})
}


func (svr *HealthCheckServer) GetStatus(w http.ResponseWriter, r *http.Request){
	c:=svr.statusFunc()
	status := &Status{
		Connected:c,
		LatestErrors:make([]*Node,0),
	}

	for _, e := range svr.errorChannel.GetErrors(){
		if e != nil {
			status.LatestErrors = append(status.LatestErrors, e)
		}
	}

	if !c{
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	js, _ := json.Marshal(status)
	w.Write(js)
	w.WriteHeader(http.StatusOK)
}