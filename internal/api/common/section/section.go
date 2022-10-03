package section

import "stash-vr/internal/stash/gql"

type Section struct {
	Name             string
	FilterId         string
	PreviewPartsList []gql.ScenePreviewParts
}

type SectionCount struct {
	Links  int
	Scenes int
}

func Count(sections []Section) SectionCount {
	var linkCount int
	sceneIds := make(map[string]any)
	for _, s := range sections {
		linkCount += len(s.PreviewPartsList)
		for _, p := range s.PreviewPartsList {
			sceneIds[p.Id] = struct{}{}
		}
	}
	return SectionCount{
		Links:  linkCount,
		Scenes: len(sceneIds),
	}
}

func ContainsFilterId(id string, list []Section) bool {
	for _, v := range list {
		if id == v.FilterId {
			return true
		}
	}
	return false
}
