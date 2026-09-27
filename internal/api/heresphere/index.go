package heresphere

import (
	"stash-vr/internal/library"
	"stash-vr/internal/multipart"
)

type indexDto struct {
	Access  int          `json:"access"`
	Library []libraryDto `json:"library"`
}

type libraryDto struct {
	Name string   `json:"name"`
	List []string `json:"list"`
}

func buildIndex(sections []library.Section, vds map[string]*library.VideoData, baseUrl string) (indexDto, error) {
	index := indexDto{Access: 1, Library: make([]libraryDto, 0, len(sections))}

	for _, section := range sections {
		l := libraryDto{
			Name: section.Name,
			List: make([]string, 0),
		}

		for _, sceneId := range section.Ids {
			if vd, ok := vds[sceneId]; ok {
				for _, item := range multipart.GetPlaybackItems(sceneId, vd.SceneParts.Files) {
					l.List = append(l.List, getVideoDataUrl(baseUrl, item.VideoId))
				}
			}
		}
		index.Library = append(index.Library, l)
	}

	return index, nil
}
