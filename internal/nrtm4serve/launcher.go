package nrtm4serve

import (
	"log"
	"net/http"
	"time"

	"gitlab.com/etchells/nrtm4client/internal/nrtm4/pg"
	"gitlab.com/etchells/nrtm4client/internal/nrtm4/service"
	"gitlab.com/etchells/nrtm4client/internal/nrtm4serve/rpc"
)

// Launch sets up the rpc handler and starts the server
func Launch(config service.AppConfig, port int, webDir string) {
	repo := pg.PostgresRepository{}
	if err := repo.Initialize(config.PgDatabaseURL); err != nil {
		log.Fatal("Failed to initialize repository")
	}
	defer repo.Close()
	rpcHandler := rpc.Handler{API: WebAPI{Repo: repo}}
	logger.Info("NRTM4serve is starting", "port", port)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Recovered from Panic in launcher", "recover", r)
			time.Sleep(time.Second * 20)
		}
	}()
	s := rpc.NewServer()
	s.POSTHandler("/rpc", rpcHandler.RPCServiceWrapper)
	if len(webDir) > 0 {
		s.Router().Handle("/", http.FileServer(http.Dir(webDir)))
	}
	s.Serve(port)
}
