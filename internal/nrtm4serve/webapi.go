package nrtm4serve

import (
	"gitlab.com/etchells/nrtm4client/internal/nrtm4/persist"
	"gitlab.com/etchells/nrtm4client/internal/nrtm4serve/rpc"
)

// WebAPI defines the RPC functions used by the web client
type WebAPI struct {
	rpc.API
	repo persist.Repository
}

func (api WebAPI) GetSources() ([]persist.NRTMSource, error) {
	return api.repo.GetSources()
}
