package heresphere

import (
	"stash-vr/internal/library"
)

type indexDto struct {
	Access  int          `json:"access"`
	Library []libraryDto `json:"library"`
}

type libraryDto struct {
	Name string   `json:"name"`
	List []string `json:"list"`
}

func buildIndex(sections []library.Section, baseUrl string) (indexDto, error) {
	index := indexDto{Library: make([]libraryDto, 0, len(sections))}

	for _, section := range sections {
		l := libraryDto{
			Name: section.Name,
			List: make([]string, len(section.Ids)),
		}
		index.Library = append(index.Library, l)
		for i, sceneId := range section.Ids {
			l.List[i] = getVideoDataUrl(baseUrl, sceneId)
		}
	}

	return index, nil
}
