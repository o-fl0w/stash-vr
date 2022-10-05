package index

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/api/common"
	"stash-vr/internal/api/common/section"
	"stash-vr/internal/stash"
	"stash-vr/internal/util"
)

func Build(ctx context.Context, client graphql.Client, baseUrl string) Index {
	sections := common.GetIndex(ctx, client)

	scenes := fromSections(baseUrl, sections)

	index := Index{Authorized: "1", Scenes: scenes}

	return index
}

func fromSections(baseUrl string, sections []section.Section) []Scene {
	return util.Transform[section.Section, Scene](func(section section.Section) *Scene {
		return util.Ptr(fromSection(baseUrl, section))
	}).Ordered(sections)
}

func fromSection(baseUrl string, section section.Section) Scene {
	s := Scene{
		Name: section.Name,
		List: make([]PreviewData, len(section.PreviewPartsList)),
	}
	for i, p := range section.PreviewPartsList {
		s.List[i] = PreviewData{
			Id:           p.Id,
			ThumbnailUrl: stash.ApiKeyed(p.Paths.Screenshot),
			Title:        p.Title,
			VideoLength:  int(p.File.Duration),
			VideoUrl:     videoDataUrl(baseUrl, p.Id),
		}
	}
	return s
}

func videoDataUrl(baseUrl string, id string) string {
	return fmt.Sprintf("%s/deovr/%s", baseUrl, id)
}
