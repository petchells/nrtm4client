package nrtm4serve

import (
	"net/http"

	"gitlab.com/etchells/nrtm4client/internal/nrtm4/persist"
	"gitlab.com/etchells/nrtm4client/internal/nrtm4serve/rpc"
)

// WebAPI defines the RPC functions used by the web client
type WebAPI struct {
	rpc.API
	Repo persist.Repository
}

// GetAuth implements interface -- allows requests to all methods
func (api WebAPI) GetAuth(w http.ResponseWriter, r *http.Request, req rpc.JSONRPCRequest) (rpc.WebSession, bool) {
	return rpc.WebSession{}, true
}

// GetSources returns a list of sources
func (api WebAPI) GetSources() ([]persist.NRTMSource, error) {
	return api.Repo.GetSources()
}
