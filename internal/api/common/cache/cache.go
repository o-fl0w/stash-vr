package cache

import (
	"stash-vr/internal/api/common/types"
	"sync"
)

var Store store

type store struct {
	Index index
}

type index struct {
	lock     sync.RWMutex
	sections []types.Section
}

func (c *index) Get() []types.Section {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.sections
}

func (c *index) Set(ss []types.Section) {
	c.lock.Lock()
	c.sections = ss
	c.lock.Unlock()
}
