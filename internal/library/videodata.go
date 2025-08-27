package library

import (
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
)

type VideoData struct {
	SceneParts *gql.SceneParts
}

func (vd VideoData) Title() string {
	return util.FirstNonEmpty(vd.SceneParts.Title, &vd.SceneParts.Files[0].Basename)
}

func (vd VideoData) Id() string {
	return vd.SceneParts.Id
}
