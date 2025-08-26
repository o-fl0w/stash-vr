package library

import (
	"context"
	"stash-vr/internal/stash/gql"
)

type Tag struct {
	Id        string
	Name      string
	SortName  *string
	ParentIds []string
}

func (service *Service) LoadTags(ctx context.Context) error {
	resp, err := gql.FindAllTags(ctx, service.StashClient)
	if err != nil {
		return err
	}
	service.tags = make(map[string]*Tag)
	for _, st := range resp.FindTags.Tags {
		t := Tag{
			Id:       st.Id,
			Name:     st.Name,
			SortName: st.Sort_name,
		}

		for _, p := range st.Parents {
			t.ParentIds = append(t.ParentIds, p.Id)
		}
		service.tags[st.Id] = &t
	}
	return nil
}

func (service *Service) ancestorTagNames(ctx context.Context, tagId string) []string {
	visited := map[string]struct{}{tagId: {}}
	queue := []string{tagId}
	out := []string{}

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		t := service.tags[id]
		if t == nil {
			continue
		}
		for _, pid := range t.ParentIds {
			if _, seen := visited[pid]; seen {
				continue
			}
			visited[pid] = struct{}{}
			p := service.tags[pid]
			queue = append(queue, pid)

			if p.SortName != nil && *p.SortName == "hidden" {
				continue
			}
			out = append(out, p.Name)
		}
	}
	return out
}
