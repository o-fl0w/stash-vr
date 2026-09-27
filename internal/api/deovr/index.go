package deovr

import (
	"stash-vr/internal/library"
	"stash-vr/internal/multipart"
	"stash-vr/internal/stash"
	"stash-vr/internal/util"
)

type indexDto struct {
	Authorized string     `json:"authorized"`
	Scenes     []sceneDto `json:"scenes"`
}

type sceneDto struct {
	Name string           `json:"name"`
	List []previewDataDto `json:"list"`
}

type previewDataDto struct {
	Id           string  `json:"id"`
	ThumbnailUrl *string `json:"thumbnailUrl"`
	Title        string  `json:"title"`
	VideoLength  int     `json:"videoLength"`
	VideoUrl     string  `json:"video_url"`
}

func buildIndex(sections []library.Section, vds map[string]*library.VideoData, baseUrl string) (indexDto, error) {
	index := indexDto{Authorized: "1", Scenes: make([]sceneDto, 0, len(sections))}

	for _, section := range sections {
		s := sceneDto{Name: section.Name, List: make([]previewDataDto, 0)}

		for _, sceneId := range section.Ids {
			if vd, ok := vds[sceneId]; ok {
				for _, item := range multipart.GetPlaybackItems(sceneId, vd.SceneParts.Files) {
					videoUrl := getVideoDataUrl(baseUrl, item.VideoId)
					videoUrl = multipart.AppendQueryParam(videoUrl, "fileId", item.FileId)

					previewData := previewDataDto{
						Id:          multipart.HashVideoId(item.VideoId),
						Title:       multipart.FormatTitle(vd.Title(), item.Label),
						VideoLength: int(item.File.Duration),
						VideoUrl:    videoUrl,
					}

					if vd.SceneParts.Paths.Screenshot != nil {
						previewData.ThumbnailUrl = util.Ptr(stash.ApiKeyed(*vd.SceneParts.Paths.Screenshot))
					}

					s.List = append(s.List, previewData)
				}
			}
		}

		if len(s.List) > 0 {
			index.Scenes = append(index.Scenes, s)
		}
	}

	return index, nil
}
