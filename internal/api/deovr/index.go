package deovr

import (
	"stash-vr/internal/library"
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
	index := indexDto{Authorized: "1", Scenes: make([]sceneDto, len(sections))}

	for i, section := range sections {
		s := sceneDto{
			Name: section.Name,
			List: make([]previewDataDto, len(section.Ids)),
		}
		index.Scenes[i] = s

		for j, sectionSceneId := range section.Ids {
			vd := vds[sectionSceneId]
			s.List[j] = previewDataDto{
				Id:          vd.SceneParts.Id,
				Title:       vd.Title(),
				VideoLength: int(vd.SceneParts.Files[0].Duration),
				VideoUrl:    getVideoDataUrl(baseUrl, vd.Id()),
			}
			if vd.SceneParts.Paths.Screenshot != nil {
				s.List[j].ThumbnailUrl = util.Ptr(stash.ApiKeyed(*vd.SceneParts.Paths.Screenshot))
			}
		}
	}

	return index, nil
}
