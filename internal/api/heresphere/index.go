package heresphere

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	_library "stash-vr/internal/sections"
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
	ss := _library.Get(ctx, client)

	index := index{Access: 1, Library: fromSections(baseUrl, ss)}

	return index
}

func fromSections(baseUrl string, sections []section.Section) []library {
	return util.Transform[section.Section, library](func(section section.Section) *library {
		return util.Ptr(fromSection(baseUrl, section))
	}).Ordered(sections)
}

func fromSection(baseUrl string, section section.Section) library {
	o := library{Name: section.Name}
	for _, p := range section.PreviewPartsList {
		o.List = append(o.List, getVideoDataUrl(baseUrl, p.Id))
	}
	return o
}
