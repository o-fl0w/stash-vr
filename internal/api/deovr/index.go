package deovr

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/sections"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash"
	"stash-vr/internal/stimhub"
	"stash-vr/internal/util"
)

type index struct {
	Authorized string  `json:"authorized"`
	Scenes     []scene `json:"scenes"`
}

type scene struct {
	Name string        `json:"name"`
	List []previewData `json:"list"`
}

type previewData struct {
	Id           string `json:"id"`
	ThumbnailUrl string `json:"thumbnailUrl"`
	Title        string `json:"title"`
	VideoLength  int    `json:"videoLength"`
	VideoUrl     string `json:"video_url"`
}

func buildIndex(ctx context.Context, client graphql.Client, baseUrl string) index {
	ss := sections.Get(ctx, client, nil)

	scenes := fromSections(baseUrl, ss)

	index := index{Authorized: "1", Scenes: scenes}

	return index
}

func fromSections(baseUrl string, sections []section.Section) []scene {
	return util.Transform[section.Section, scene](func(section section.Section) (scene, error) {
		if section.FilterId == stimhub.FilterId {
			return scene{}, nil
		}
		return fromSection(baseUrl, section), nil
	}).Ordered(sections)
}

func fromSection(baseUrl string, section section.Section) scene {
	s := scene{
		Name: section.Name,
		List: make([]previewData, len(section.Scenes)),
	}
	for i, p := range section.Scenes {
		s.List[i] = previewData{
			Id:           p.Id(),
			ThumbnailUrl: stash.ApiKeyed(p.Paths.Screenshot),
			Title:        p.ScenePreviewParts.Title,
			VideoLength:  int(p.Files[0].Duration),
			VideoUrl:     getVideoDataUrl(baseUrl, p.Id()),
		}
	}
	return s
}
