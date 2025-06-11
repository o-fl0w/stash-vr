package library

import (
	"github.com/Khan/genqlient/graphql"
	"golang.org/x/sync/singleflight"
	"maps"
	"sync"
)

type Service struct {
	StashClient graphql.Client
	vdCache     map[string]*VideoData
	mu          sync.RWMutex
	single      singleflight.Group
	Stats       Stats
}

func (service *Service) snapshot() map[string]*VideoData {
	service.mu.RLock()
	defer service.mu.RUnlock()
	return maps.Clone(service.vdCache)
}

func NewService(client graphql.Client) *Service {
	return &Service{
		StashClient: client,
		vdCache:     make(map[string]*VideoData),
	}
}

type Stats struct {
	Links  int
	Scenes int
}
