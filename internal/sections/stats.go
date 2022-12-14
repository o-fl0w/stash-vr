package sections

import "stash-vr/internal/sections/section"

type Stats struct {
	Links  int
	Scenes int
}

func Count(sections []section.Section) Stats {
	var linkCount int
	sceneIds := make(map[string]any)
	for _, s := range sections {
		linkCount += len(s.PreviewPartsList)
		for _, p := range s.PreviewPartsList {
			sceneIds[p.Id] = struct{}{}
		}
	}
	return Stats{
		Links:  linkCount,
		Scenes: len(sceneIds),
	}
}
