package section

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
	"strconv"
)

const sourceFrontPage = "Front Page"

func SectionsByFrontPage(ctx context.Context, client graphql.Client, prefix string) ([]Section, error) {
	filterIds, err := findSavedFilterIdsByFrontPage(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("FindSavedFilterIdsByFrontPage: %w", err)
	}

	savedFilters := findFiltersById(ctx, client, filterIds)

	sections := util.Transform[gql.SavedFilterParts, Section](func(savedFilter gql.SavedFilterParts) (Section, error) {
		section, err := sectionFromSavedFilter(ctx, client, prefix, savedFilter)
		if err != nil {
			filterLogger(ctx, savedFilter, sourceFrontPage).Warn().Err(err).Msg("Skipped filter: sectionsByFrontPage: sectionFromSavedFilter")
			return Section{}, err
		}
		sectionLogger(ctx, savedFilter, sourceFrontPage, section).Debug().Msg("Section built")
		return section, nil
	}).Ordered(savedFilters)

	return sections, nil
}

func findSavedFilterIdsByFrontPage(ctx context.Context, client graphql.Client) ([]string, error) {
	configurationResponse, err := gql.UIConfiguration(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("UIConfiguration: %w", err)
	}

	var filterIds []string
	frontPageContent := configurationResponse.Configuration.Ui["frontPageContent"]
	if frontPageContent == nil {
		log.Ctx(ctx).Info().Msg("No frontpage content found")
		return nil, nil
	}
	frontPageFilters := configurationResponse.Configuration.Ui["frontPageContent"].([]interface{})
	for _, _filter := range frontPageFilters {
		filter := _filter.(map[string]interface{})
		typeName := filter["__typename"].(string)
		if typeName != "SavedFilter" {
			log.Ctx(ctx).Info().Str("type", typeName).Str("source", sourceFrontPage).Msg("Skipped filter of unsupported type. Only user created saved scene filters are supported")
			continue
		}

		filterId := strconv.Itoa(int(filter["savedFilterId"].(float64)))
		filterIds = append(filterIds, filterId)
	}

	return filterIds, nil
}
