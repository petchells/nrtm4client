package nrtm4serve

import (
	"net/http"
	"time"

	"gitlab.com/etchells/nrtm4client/internal/nrtm4/service"
	"gitlab.com/etchells/nrtm4client/internal/nrtm4serve/rpc"
)

// Launch sets up the rpc handler and starts the server
func Launch(config service.AppConfig, port int, webDir string) {
	rpcHandler := rpc.Handler{API: WebAPI{}}
	logger.Info("Cashdash api server starting", "port", port)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Recovered from Panic in launcher", "recover", r)
			time.Sleep(time.Second * 20)
		}
	}()
	s := rpc.NewServer()
	s.POSTHandler("/api", rpcHandler.RPCServiceWrapper)
	if len(webDir) > 0 {
		s.Router().Handle("/", http.FileServer(http.Dir(webDir)))
	}
	s.Serve(port)
}
