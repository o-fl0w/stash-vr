package index

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/section"
	"stash-vr/internal/section/model"
	"stash-vr/internal/stash"
	"stash-vr/internal/util"
)

type Index struct {
	Authorized string  `json:"authorized"`
	Scenes     []Scene `json:"scenes"`
}

type Scene struct {
	Name string        `json:"name"`
	List []PreviewData `json:"list"`
}

type PreviewData struct {
	Id           string `json:"id"`
	ThumbnailUrl string `json:"thumbnailUrl"`
	Title        string `json:"title"`
	VideoLength  int    `json:"videoLength"`
	VideoUrl     string `json:"video_url"`
}

func Build(ctx context.Context, client graphql.Client, baseUrl string) Index {
	sections := section.Get(ctx, client)

	scenes := fromSections(baseUrl, sections)

	index := Index{Authorized: "1", Scenes: scenes}

	return index
}

func fromSections(baseUrl string, sections []model.Section) []Scene {
	return util.Transform[model.Section, Scene](func(section model.Section) *Scene {
		return util.Ptr(fromSection(baseUrl, section))
	}).Ordered(sections)
}

func fromSection(baseUrl string, section model.Section) Scene {
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
