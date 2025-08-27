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
	muVdCache   sync.RWMutex
	single      singleflight.Group
	Stats       Stats

	tagCache map[string]*Tag
}

func (libraryService *Service) snapshot() map[string]*VideoData {
	libraryService.muVdCache.RLock()
	defer libraryService.muVdCache.RUnlock()
	return maps.Clone(libraryService.vdCache)
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
