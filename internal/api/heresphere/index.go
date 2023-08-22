package heresphere

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/sections"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/util"
)

type index struct {
	Access  int       `json:"access"`
	Library []library `json:"library"`
}

type library struct {
	Name string   `json:"name"`
	List []string `json:"list"`
}

func buildIndex(ctx context.Context, client graphql.Client, baseUrl string) index {
	ss := sections.Get(ctx, client)

	index := index{Access: 1, Library: fromSections(baseUrl, ss)}

	return index
}

func fromSections(baseUrl string, sections []section.Section) []library {
	return util.Transform[section.Section, library](func(section section.Section) (library, error) {
		return fromSection(baseUrl, section), nil
	}).Ordered(sections)
}

func fromSection(baseUrl string, section section.Section) library {
	o := library{Name: section.Name, List: make([]string, len(section.Scenes))}
	for i, p := range section.Scenes {
		o.List[i] = getVideoDataUrl(baseUrl, p.Id())
	}
	return o
}
