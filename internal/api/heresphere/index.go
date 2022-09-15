package heresphere

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/api/common"
	"stash-vr/internal/api/common/types"
	"stash-vr/internal/util"
)

type Index struct {
	Access  int       `json:"access"`
	Library []Library `json:"library"`
}

type VideoDataUrl string

type Library struct {
	Name string         `json:"name"`
	List []VideoDataUrl `json:"list"`
}

func buildIndex(ctx context.Context, client graphql.Client, baseUrl string) (Index, error) {
	sections := common.GetIndex(ctx, client)

	index := Index{Access: 1, Library: fromSections(baseUrl, sections)}

	return index, nil
}

func fromSections(baseUrl string, sections []types.Section) []Library {
	return util.Transformation[types.Section, Library]{
		Transform: func(section types.Section) (Library, error) {
			return fromSection(baseUrl, section), nil
		},
	}.Ordered(sections)
}

func fromSection(baseUrl string, section types.Section) Library {
	o := Library{Name: section.Name}
	for _, p := range section.PreviewPartsList {
		o.List = append(o.List, videoDataUrl(baseUrl, p.Id))
	}
	return o
}

func videoDataUrl(baseUrl string, id string) VideoDataUrl {
	return VideoDataUrl(fmt.Sprintf("%s/heresphere/%s", baseUrl, id))
}
