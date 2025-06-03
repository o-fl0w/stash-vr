package library

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"golang.org/x/sync/singleflight"
	"maps"
	"stash-vr/internal/stash/gql"
	"sync"
)

type Service struct {
	StashClient graphql.Client
	vdCache     map[string]*VideoData
	mu          sync.RWMutex
	single      singleflight.Group
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

func (service *Service) GetClientVersions(ctx context.Context) (map[string]string, error) {
	version, err := gql.Version(ctx, service.StashClient)
	if err != nil {
		return nil, err
	}
	return map[string]string{"stash": *version.Version.Version}, nil
}
