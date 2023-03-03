package sections

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"stash-vr/internal/cache"
	"stash-vr/internal/config"
	"stash-vr/internal/efile"
	"stash-vr/internal/sections/internal"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/title"
	"strings"
	"sync"
)

var c cache.Cache[[]section.Section]

func Get(ctx context.Context, client graphql.Client) []section.Section {
	return c.Get(ctx, func(ctx context.Context) []section.Section {

		g, _ := errgroup.WithContext(ctx)

		eFileNamesChan := make(chan []string, 1)
		var sections []section.Section

		g.Go(func() error {
			eFileNames, err := efile.GetEFileNames(config.Get().EventServerUrl + "/efiles")
			if err != nil {
				return err
			}
			eFileNamesChan <- eFileNames
			close(eFileNamesChan)
			return nil
		})

		g.Go(func() error {
			sections = build(ctx, client, config.Get().Filters)
			return nil
		})

		err := g.Wait()
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to get EFiles")
		} else {
			eFileNames := <-eFileNamesChan

			for i := 0; i < len(sections); i++ {
				sect := &sections[i]
				for j := 0; j < len(sect.Scene); j++ {
					scene := sect.Scene[j]

					matchingEFileNames := efile.FindAllMatchingEFileNames(scene.GetFiles()[0].Basename, eFileNames)
					if len(matchingEFileNames) == 0 {
						continue
					}

					log.Ctx(ctx).Trace().Str("section", sect.Name).Int("count", len(matchingEFileNames)).Str("videoFileName", scene.GetFiles()[0].Basename).Strs("EFileNames", matchingEFileNames).Msg("Found matching EFiles")

					for _, matchingEFileName := range matchingEFileNames {
						title := title.GetSceneTitle(scene.Title, scene.GetFiles()[0].Basename)
						ePreview := gql.ScenePreviewParts{
							Id:    efile.MakeESceneIdWithEFileName(scene.Id, scene.GetFiles()[0].Basename, matchingEFileName),
							Title: efile.MakeESceneTitleWithEFileName(title, scene.GetFiles()[0].Basename, matchingEFileName),
							Paths: scene.Paths,
						}

						sect.Scene = append(sect.Scene, gql.ScenePreviewParts{})
						copy(sect.Scene[j+2:], sect.Scene[j+1:])
						sect.Scene[j+1] = ePreview
						j = j + 1
					}
				}
			}
		}

		count := Count(sections)
		if count.Links > 10000 {
			log.Ctx(ctx).Warn().Int("links", count.Links).Msg("More than 10.000 links generated. Known to cause issues with video players.")
		}

		log.Ctx(ctx).Info().Int("sections", len(sections)).Int("links", count.Links).Int("scenes", count.Scenes).Msg("Sections build complete")

		return sections
	})
}

func build(ctx context.Context, client graphql.Client, filters string) []section.Section {
	sss := make([][]section.Section, 3)

	wg := sync.WaitGroup{}

	if filters == "frontpage" || filters == "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ss, err := internal.SectionsByFrontpage(ctx, client, "")
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by front page")
				return
			}
			sss[0] = ss
			log.Ctx(ctx).Debug().Int("count", len(ss)).Msg("Sections built from front page")
		}()
	}

	if filters == "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ss, err := internal.SectionsBySavedFilters(ctx, client, "?:")
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by saved filters")
				return
			}
			sss[1] = ss
			log.Ctx(ctx).Debug().Int("count", len(ss)).Msg("Sections built from saved filters")
		}()
	}

	if filters != "frontpage" && filters != "" {
		filterIds := strings.Split(filters, ",")
		wg.Add(1)
		go func() {
			defer wg.Done()
			ss, err := internal.SectionsByFilterIds(ctx, client, "?:", filterIds)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by filter ids")
				return
			}
			sss[2] = ss
			log.Ctx(ctx).Debug().Int("count", len(ss)).Msg("Sections built from filter list")
		}()
	}

	wg.Wait()

	var sections []section.Section

	for _, ss := range sss {
		for _, s := range ss {
			if s.FilterId != "" && section.ContainsFilterId(s.FilterId, sections) {
				log.Ctx(ctx).Trace().Str("filterId", s.FilterId).Str("section", s.Name).Msg("Filter already added, skipping")
				continue
			}

			sections = append(sections, s)
		}
	}

	if len(sections) == 0 {
		log.Ctx(ctx).Info().Msg("No scenes found using current filters. Adding a default section with all scenes.")
		s, err := internal.SectionWithAllScenes(ctx, client)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to build custom section with all scenes")
		} else {
			if len(s.Scene) == 0 {
				log.Ctx(ctx).Info().Msg("No scenes found in Stash.")
			} else {
				sections = append(sections, s)
			}
		}
	}

	return sections
}
