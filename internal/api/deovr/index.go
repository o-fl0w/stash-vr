package deovr

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/api/common"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
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
	var l []Scene
	for _, section := range sections {
		l = append(l, fromSection(baseUrl, section))
	}
	return l
}

func fromSection(baseUrl string, section common.Section) Scene {
	o := Scene{Name: section.Name}
	for _, p := range section.PreviewPartsList {
		o.List = append(o.List, fromPreviewParts(baseUrl, p))
	}
	return o
}

func fromPreviewParts(baseUrl string, s gql.ScenePreviewParts) PreviewData {
	return PreviewData{
		Id:           s.Id,
		ThumbnailUrl: stash.ApiKeyed(s.Paths.Screenshot),
		Title:        s.Title,
		VideoLength:  int(s.File.Duration),
		VideoUrl:     videoDataUrl(baseUrl, s.Id),
	}
}

func videoDataUrl(baseUrl string, id string) string {
	return fmt.Sprintf("%s/deovr/%s", baseUrl, id)
}
