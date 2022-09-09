package deovr

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
)

type Index struct {
	Authorized string  `json:"authorized"`
	Scenes     []Scene `json:"scenes"`
}

type PreviewVideoData struct {
	Id           string `json:"id"`
	ThumbnailUrl string `json:"thumbnailUrl"`
	Title        string `json:"title"`
	VideoLength  int    `json:"videoLength"`
	VideoUrl     string `json:"video_url"`
}

type Scene struct {
	Name string             `json:"name"`
	List []PreviewVideoData `json:"list"`

	_savedFilterId string
}

func buildIndex(ctx context.Context, client graphql.Client, baseUrl string) (Index, error) {
	index := Index{Authorized: "1", Scenes: []Scene{}}

	if err := sectionsCustom(ctx, client, baseUrl, "", &index.Scenes); err != nil {
		return Index{}, fmt.Errorf("sectionsCustom: %w", err)
	}

	if err := sectionsByFrontPage(ctx, client, baseUrl, "", &index.Scenes); err != nil {
		return Index{}, fmt.Errorf("sectionsByFrontPage: %w", err)
	}

	if err := sectionsBySavedFilters(ctx, client, baseUrl, "?:", &index.Scenes); err != nil {
		return Index{}, fmt.Errorf("sectionsBySavedFilters: %w", err)
	}

	//if err := sectionsByTags(ctx, client, baseUrl, "#:", &index.Scenes); err != nil {
	//	return Index{}, fmt.Errorf("sectionsByTags: %w", err)
	//}

	log.Info().Str("route", "index").Int("#categories", len(index.Scenes)).Send()

	return index, nil
}

func sectionsCustom(ctx context.Context, client graphql.Client, baseUrl string, prefix string, destination *[]Scene) error {
	scene := Scene{
		Name: fmt.Sprintf("%s%s", prefix, "All"),
	}
	sceneFilter := gql.SceneFilterType{}
	scenesResponse, err := gql.FindScenesByFilter(ctx, client, &sceneFilter, "", gql.SortDirectionEnumAsc)
	if err != nil {
		return fmt.Errorf("FindScenesByFilter: %w", err)
	}
	for _, s := range scenesResponse.FindScenes.Scenes {
		scene.List = append(scene.List, getPreviewVideoData(baseUrl, s.ScenePreviewParts))
	}
	*destination = append(*destination, scene)

	return nil
}

func sectionsByFrontPage(ctx context.Context, client graphql.Client, baseUrl string, prefix string, destination *[]Scene) error {
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

		scene := Scene{
			Name:           fmt.Sprintf("%s%s", prefix, savedFilterResponse.FindSavedFilter.Name),
			_savedFilterId: id,
		}
		for _, s := range scenesResponse.FindScenes.Scenes {
			scene.List = append(scene.List, getPreviewVideoData(baseUrl, s.ScenePreviewParts))
		}
		*destination = append(*destination, scene)
	}

	return nil
}

func sectionsBySavedFilters(ctx context.Context, client graphql.Client, baseUrl string, prefix string, destination *[]Scene) error {
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
			return fmt.Errorf("FindSceneIdsByFilters: %w", err)
		}
		if len(scenesResponse.FindScenes.Scenes) == 0 {
			continue
		}

		scene := Scene{
			Name:           fmt.Sprintf("%s%s", prefix, savedFilter.Name),
			_savedFilterId: savedFilter.Id,
		}

		for _, s := range scenesResponse.FindScenes.Scenes {
			scene.List = append(scene.List, getPreviewVideoData(baseUrl, s.ScenePreviewParts))
		}

		*destination = append(*destination, scene)
	}

	return nil
}

func sectionsByTags(ctx context.Context, client graphql.Client, baseUrl string, prefix string, destination *[]Scene) error {
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

	tagMap := make(map[string]map[string]PreviewVideoData)
	for _, s := range scenesResponse.FindScenes.Scenes {
		for _, tag := range s.Tags {
			hsTagName := fmt.Sprintf("%s%s", prefix, tag.Name)
			if tagMap[hsTagName] == nil {
				tagMap[hsTagName] = make(map[string]PreviewVideoData)
			}

			if _, ok := tagMap[hsTagName][s.Id]; !ok {
				tagMap[hsTagName][s.Id] = getPreviewVideoData(baseUrl, s.ScenePreviewParts)
			}
		}
	}

	for k := range tagMap {
		if len(tagMap[k]) == 0 {
			continue
		}
		scene := Scene{
			Name: k,
		}
		for _, vd := range tagMap[k] {
			scene.List = append(scene.List, vd)
		}
		*destination = append(*destination, scene)
	}

	return nil
}

func getPreviewVideoData(baseUrl string, s gql.ScenePreviewParts) PreviewVideoData {
	return PreviewVideoData{
		Id:           s.Id,
		ThumbnailUrl: stash.ApiKeyed(s.Paths.Screenshot),
		Title:        s.Title,
		VideoLength:  int(s.File.Duration),
		VideoUrl:     videoDataUrl(baseUrl, s.Id),
	}
}

func videoDataUrl(baseUrl string, id string) string {
	return fmt.Sprintf("%s/deovr/%s", baseUrl, id)
}

func containsSavedFilterId(id string, list []Scene) bool {
	for _, v := range list {
		if id == v._savedFilterId {
			return true
		}
	}
	return false
}
