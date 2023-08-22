package section

import (
	"stash-vr/internal/efile"
	"stash-vr/internal/stash/gql"
)

type Section struct {
	Name     string
	FilterId string
	Scenes   []ScenePreview
}

type ScenePreview struct {
	gql.ScenePreviewParts
	*efile.EScene
}

func (s ScenePreview) Title() string {
	if s.EScene != nil && s.EScene.Title != "" {
		return s.EScene.Title
	}
	if s.ScenePreviewParts.Title != "" {
		return s.ScenePreviewParts.Title
	}
	return s.ScenePreviewParts.Files[0].Basename
}

func (s ScenePreview) Id() string {
	if s.EScene != nil {
		return efile.MakeESceneId(s.ScenePreviewParts.Id, s.EScene.Oshash)
	}
	return s.ScenePreviewParts.Id
}
func ContainsFilterId(id string, list []Section) bool {
	for _, v := range list {
		if id == v.FilterId {
			return true
		}
	}
	return false
}
