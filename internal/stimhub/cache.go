package stimhub

import "sync"

var cache struct {
	stimScenes []StimScene
	lock       sync.RWMutex
}

func Get(audioCrc32 string, sceneId string) *StimScene {
	cache.lock.RLock()
	defer cache.lock.RUnlock()
	for i := range cache.stimScenes {
		if cache.stimScenes[i].AudioCrc32 == audioCrc32 && cache.stimScenes[i].SceneId == sceneId {
			return &cache.stimScenes[i]
		}
	}
	return nil
}
func Set(stimScenes []StimScene) {
	cache.lock.Lock()
	defer cache.lock.Unlock()
	cache.stimScenes = stimScenes
}
