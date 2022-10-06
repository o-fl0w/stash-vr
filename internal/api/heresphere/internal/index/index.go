package index

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/section"
	"stash-vr/internal/section/model"
	"stash-vr/internal/util"
)

type Index struct {
	Access  int       `json:"access"`
	Library []library `json:"library"`
}

type VideoDataUrl string

type library struct {
	Name string         `json:"name"`
	List []VideoDataUrl `json:"list"`
}

func Build(ctx context.Context, client graphql.Client, baseUrl string) Index {
	sections := section.Get(ctx, client)

	index := Index{Access: 1, Library: fromSections(baseUrl, sections)}

	return index
}

func fromSections(baseUrl string, sections []model.Section) []library {
	return util.Transform[model.Section, library](func(section model.Section) *library {
		return util.Ptr(fromSection(baseUrl, section))
	}).Ordered(sections)
}

func fromSection(baseUrl string, section model.Section) library {
	o := library{Name: section.Name}
	for _, p := range section.PreviewPartsList {
		o.List = append(o.List, GetVideoDataUrl(baseUrl, p.Id))
	}
	return o
}

func GetVideoDataUrl(baseUrl string, id string) VideoDataUrl {
	return VideoDataUrl(fmt.Sprintf("%s/heresphere/%s", baseUrl, id))
}
