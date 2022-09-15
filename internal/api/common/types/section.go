package types

import "stash-vr/internal/stash/gql"

type Section struct {
	Name             string
	FilterId         string
	PreviewPartsList []gql.ScenePreviewParts
}
