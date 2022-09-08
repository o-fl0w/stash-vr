package heresphere

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
)

type Index struct {
	Access  int       `json:"access"`
	Library []Library `json:"library"`
}

type VideoDataUrl string

type Library struct {
	Name string         `json:"name"`
	List []VideoDataUrl `json:"list"`

	_savedFilterId string
}

func buildIndex(ctx context.Context, client graphql.Client, baseUrl string) (Index, error) {
	index := Index{Access: 1}

	if err := sectionsCustom(ctx, client, baseUrl, "", &index.Library); err != nil {
		return Index{}, fmt.Errorf("sectionsCustom: %w", err)
	}

	if err := sectionsByFrontPage(ctx, client, baseUrl, "", &index.Library); err != nil {
		return Index{}, fmt.Errorf("sectionsByFrontPage: %w", err)
	}

	if err := sectionsBySavedFilters(ctx, client, baseUrl, "?:", &index.Library); err != nil {
		return Index{}, fmt.Errorf("sectionsBySavedFilters: %w", err)
	}

	//if err := sectionsByTags(ctx, serverUrl, baseUrl, "#:", &index.Library); err != nil {
	//	return Index{}, fmt.Errorf("sectionsByTags: %w", err)
	//}

	log.Info().Str("route", "index").Int("#categories", len(index.Library)).Send()

	return index, nil
}

func sectionsCustom(ctx context.Context, client graphql.Client, baseUrl string, prefix string, destination *[]Library) error {
	library := Library{
		Name: fmt.Sprintf("%s%s", prefix, "All"),
	}
	sceneFilter := gql.SceneFilterType{}
	scenesResponse, err := gql.FindScenesByFilter(ctx, client, &sceneFilter, "", gql.SortDirectionEnumAsc)
	if err != nil {
		return fmt.Errorf("FindScenesByFilter: %w", err)
	}
	for _, s := range scenesResponse.FindScenes.Scenes {
		library.List = append(library.List, videoDataUrl(baseUrl, s.Id))
	}
	*destination = append(*destination, library)

	return nil
}

func sectionsByFrontPage(ctx context.Context, client graphql.Client, baseUrl string, prefix string, destination *[]Library) error {
	ids, err := stash.FindFrontPageSavedFilterIds(ctx, client)
	if err != nil {
		return fmt.Errorf("FindFrontPageSavedFilterIds: %w", err)
	}

	for _, id := range ids {
		savedFilterResponse, err := gql.FindSavedFilter(ctx, client, id)
		if err != nil {
			return fmt.Errorf("FindSavedFilter: %w", err)
		}

		filter, err := stash.ParseJsonEncodedFilter(savedFilterResponse.FindSavedFilter.Filter)
		if err != nil {
			return fmt.Errorf("ParseJsonEncodedFilter: %w", err)
		}

		scenesResponse, err := gql.FindScenesByFilter(ctx, client, &filter.SceneFilter, filter.SortBy, filter.SortDir)
		if err != nil {
			return fmt.Errorf("FindScenesByFilter: %w", err)
		}

		library := Library{
			Name:           fmt.Sprintf("%s%s", prefix, savedFilterResponse.FindSavedFilter.Name),
			_savedFilterId: id,
		}
		for _, s := range scenesResponse.FindScenes.Scenes {
			library.List = append(library.List, videoDataUrl(baseUrl, s.Id))
		}
		*destination = append(*destination, library)
	}

	return nil
}

func sectionsBySavedFilters(ctx context.Context, client graphql.Client, baseUrl string, prefix string, destination *[]Library) error {
	savedFiltersResponse, err := gql.FindSavedFilters(ctx, client)
	if err != nil {
		return fmt.Errorf("FindSavedFilters: %w", err)
	}

	for _, savedFilter := range savedFiltersResponse.FindSavedFilters {
		if savedFilter.Name == "" || containsSavedFilterId(savedFilter.Id, *destination) {
			continue
		}

		filter, err := stash.ParseJsonEncodedFilter(savedFilter.Filter)
		if err != nil {
			return fmt.Errorf("ParseJsonEncodedFilter: %w", err)
		}

		scenesResponse, err := gql.FindScenesByFilter(ctx, client, &filter.SceneFilter, filter.SortBy, filter.SortDir)
		if err != nil {
			return fmt.Errorf("FindScenesByFilter: %w", err)
		}
		if len(scenesResponse.FindScenes.Scenes) == 0 {
			continue
		}

		library := Library{
			Name:           fmt.Sprintf("%s%s", prefix, savedFilter.Name),
			_savedFilterId: savedFilter.Id,
		}

		for _, s := range scenesResponse.FindScenes.Scenes {
			library.List = append(library.List, videoDataUrl(baseUrl, s.Id))
		}

		*destination = append(*destination, library)
	}

	return nil
}

func sectionsByTags(ctx context.Context, client graphql.Client, baseUrl string, prefix string, destination *[]Library) error {
	findTagsResponse, err := gql.FindTags(ctx, client)
	if err != nil {
		return fmt.Errorf("FindTags: %w", err)
	}

	var tagIds []string
	for _, tag := range findTagsResponse.FindTags.Tags {
		tagIds = append(tagIds, tag.Id)
	}

	scenesResponse, err := gql.FindScenesByTags(ctx, client, tagIds)
	if err != nil {
		return fmt.Errorf("FindScenesByTags: %w", err)
	}

	tagMap := make(map[string]map[string]struct{})
	for _, s := range scenesResponse.FindScenes.Scenes {
		for _, tag := range s.Tags {
			hsTagName := fmt.Sprintf("%s%s", prefix, tag.Name)
			if tagMap[hsTagName] == nil {
				tagMap[hsTagName] = make(map[string]struct{})
			}
			tagMap[hsTagName][s.Id] = struct{}{}
		}
	}

	for k := range tagMap {
		if len(tagMap[k]) == 0 {
			continue
		}
		library := Library{
			Name: k,
		}
		for id := range tagMap[k] {
			library.List = append(library.List, videoDataUrl(baseUrl, id))
		}
		*destination = append(*destination, library)
	}

	return nil
}

func videoDataUrl(baseUrl string, id string) VideoDataUrl {
	return VideoDataUrl(fmt.Sprintf("%s/heresphere/%s", baseUrl, id))
}

func containsSavedFilterId(id string, list []Library) bool {
	for _, v := range list {
		if id == v._savedFilterId {
			return true
		}
	}
	return false
}
