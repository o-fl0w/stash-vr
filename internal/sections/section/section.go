package section

import (
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/stimhub"
	"stash-vr/internal/util"
)

type Section struct {
	Name     string
	FilterId string
	Scenes   []ScenePreview
}

type ScenePreview struct {
	gql.ScenePreviewParts
	StimAudioCrc32 string
}

func (s ScenePreview) Title() string {
	if s.StimAudioCrc32 != "" {
		return stimhub.Get(s.StimAudioCrc32, s.GetId()).Title
	}
	return util.FirstNonEmpty(s.ScenePreviewParts.Title, s.ScenePreviewParts.Files[0].Basename)
}

func (s ScenePreview) Id() string {
	if s.StimAudioCrc32 != "" {
		return stimhub.MakeStimSceneId(s.GetId(), s.StimAudioCrc32)
	}
	return s.GetId()
}
