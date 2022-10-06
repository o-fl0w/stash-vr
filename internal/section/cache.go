package section

import (
	"stash-vr/internal/section/model"
	"sync"
)

var cache sections

type sections struct {
	lock     sync.RWMutex
	sections []model.Section
}

func (c *sections) Get() []model.Section {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.sections
}

func (c *sections) Set(ss []model.Section) {
	c.lock.Lock()
	c.sections = ss
	c.lock.Unlock()
}
