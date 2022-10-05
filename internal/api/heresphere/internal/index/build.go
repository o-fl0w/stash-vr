package index

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/api/common"
	"stash-vr/internal/api/common/section"
	"stash-vr/internal/util"
)

func Build(ctx context.Context, client graphql.Client, baseUrl string) Index {
	sections := common.GetIndex(ctx, client)

	index := Index{Access: 1, Library: fromSections(baseUrl, sections)}

	return index
}

func fromSections(baseUrl string, sections []section.Section) []Library {
	return util.Transform[section.Section, Library](func(section section.Section) *Library {
		return util.Ptr(fromSection(baseUrl, section))
	}).Ordered(sections)
}

func fromSection(baseUrl string, section section.Section) Library {
	o := Library{Name: section.Name}
	for _, p := range section.PreviewPartsList {
		o.List = append(o.List, videoDataUrl(baseUrl, p.Id))
	}
	return o
}

func videoDataUrl(baseUrl string, id string) VideoDataUrl {
	return VideoDataUrl(fmt.Sprintf("%s/heresphere/%s", baseUrl, id))
}
