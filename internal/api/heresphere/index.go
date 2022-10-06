package heresphere

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/cache"
	"stash-vr/internal/section"
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
	sections := cache.GetSections(ctx, client)

	index := index{Access: 1, Library: fromSections(baseUrl, sections)}

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

func getVideoDataUrl(baseUrl string, id string) string {
	return fmt.Sprintf("%s/heresphere/%s", baseUrl, id)
}
