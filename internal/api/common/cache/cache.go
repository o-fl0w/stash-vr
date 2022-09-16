package cache

import (
	"stash-vr/internal/api/common/section"
	"sync"
)

var Store store

type store struct {
	Index index
}

type index struct {
	lock     sync.RWMutex
	sections []section.Section
}

func (c *index) Get() []section.Section {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.sections
}

func (c *index) Set(ss []section.Section) {
	c.lock.Lock()
	c.sections = ss
	c.lock.Unlock()
}
