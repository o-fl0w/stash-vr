package deovr

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/api/common"
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

func buildIndex(ctx context.Context, client graphql.Client, baseUrl string) (Index, error) {
	sections := common.BuildIndex(ctx, client)

	index := Index{Authorized: "1", Scenes: fromSections(baseUrl, sections)}

	return index, nil
}

func fromSections(baseUrl string, sections []common.Section) []Scene {
	return util.Transformation[common.Section, Scene]{
		Transform: func(section common.Section) (Scene, error) {
			return fromSection(baseUrl, section), nil
		},
	}.Ordered(sections)
}

func fromSection(baseUrl string, section common.Section) Scene {
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
