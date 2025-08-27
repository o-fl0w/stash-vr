package library

import (
	"context"
	"sort"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
)

type Tag struct {
	Id        string
	Name      string
	SortName  *string
	ParentIds []string
}

func (libraryService *Service) LoadTags(ctx context.Context) error {
	resp, err := gql.FindAllTags(ctx, libraryService.StashClient)
	if err != nil {
		return err
	}
	libraryService.tagCache = make(map[string]*Tag)
	for _, st := range resp.FindTags.Tags {
		t := Tag{
			Id:       st.Id,
			Name:     st.Name,
			SortName: st.Sort_name,
		}

		for _, p := range st.Parents {
			t.ParentIds = append(t.ParentIds, p.Id)
		}
		libraryService.tagCache[st.Id] = &t
	}
	return nil
}

func (libraryService *Service) ancestors(tagId string) []Tag {
	visited := map[string]struct{}{tagId: {}}
	queue := []string{tagId}
	out := []Tag{}

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		t := libraryService.tagCache[id]
		if t == nil {
			continue
		}
		for _, pid := range t.ParentIds {
			if _, seen := visited[pid]; seen {
				continue
			}
			visited[pid] = struct{}{}
			p := libraryService.tagCache[pid]
			queue = append(queue, pid)

			out = append(out, *p)
		}
	}
	return out
}

func (libraryService *Service) decorateTags(vd *VideoData) {
	allAncestors := map[string]Tag{}
	for _, t := range vd.SceneParts.Tags {
		ancestors := libraryService.ancestors(t.Id)
		for _, a := range ancestors {
			allAncestors[a.Id] = a
		}
	}

	ordered := make([]Tag, 0, len(allAncestors))
	for _, t := range allAncestors {
		ordered = append(ordered, t)
	}

	sort.Slice(ordered, func(i, j int) bool {
		si := util.FirstNonEmpty(ordered[i].SortName, &ordered[i].Name)
		sj := util.FirstNonEmpty(ordered[j].SortName, &ordered[j].Name)
		if si == sj {
			return ordered[i].Name < ordered[j].Name
		}
		return si < sj
	})

	for _, a := range ordered {
		vd.SceneParts.Tags = append(vd.SceneParts.Tags, &gql.TagPartsArrayTagsTag{TagParts: gql.TagParts{
			Id:        a.Id,
			Name:      "#" + a.Name,
			Sort_name: a.SortName},
		})
	}
}
