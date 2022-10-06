package section

import "stash-vr/internal/stash/gql"

type Section struct {
	Name             string
	FilterId         string
	PreviewPartsList []gql.ScenePreviewParts
}

func ContainsFilterId(id string, list []Section) bool {
	for _, v := range list {
		if id == v.FilterId {
			return true
		}
	}
	return false
}
