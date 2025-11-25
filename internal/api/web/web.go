package web

import (
	"context"
	_ "embed"
	"errors"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"html/template"
	"net/http"
	"stash-vr/internal/build"
	"stash-vr/internal/config"
	"stash-vr/internal/library"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/static"
	"sync"
)

var indexTmpl = template.Must(template.ParseFS(static.Fs, "index.html"))

const (
	statusOk           = "OK"
	statusError        = "ERROR"
	statusUnauthorized = "UNAUTHORIZED"
)

type filterData struct {
	Id   string
	Name string
}

type filterOverride struct {
	ID         string
	SourceName string
	Name       string
	Disabled   bool
}

type stashData struct {
	Version             string
	FilterData          []filterData
	FilterOverrides     []filterOverride
	SampleSceneCoverUrl string
}

type indexData struct {
	Redact                  func(string) string
	Version                 string
	LogLevel                string
	ForceHTTPS              bool
	IsSyncMarkersAllowed    bool
	StashGraphQLUrl         string
	IsApiKeyProvided        bool
	StashConnectionResponse string
	StashData               *stashData
	SectionCount            int
	LinkCount               int
	SceneCount              int
}

func sampleSceneCoverUrl(ctx context.Context, stashClient graphql.Client) (string, error) {
	resp, err := gql.FindSampleSceneCover(ctx, stashClient)
	if err != nil {
		return "", err
	}
	return stash.ApiKeyed(*resp.FindScenes.Scenes[0].Paths.Screenshot), nil
}

func stashFilters(ctx context.Context, stashClient graphql.Client) ([]filterData, error) {
	resp, err := gql.FindSavedSceneFilters(ctx, stashClient)
	if err != nil {
		return nil, err
	}
	fd := make([]filterData, len(resp.FindSavedFilters))
	for i, sf := range resp.FindSavedFilters {
		fd[i] = filterData{
			Id:   sf.Id,
			Name: sf.Name,
		}
	}
	return fd, nil
}

func filterOverrideRows(ctx context.Context, stashFilters []filterData) []filterOverride {
	cfg := config.User(ctx)

	rows := make([]filterOverride, 0, len(stashFilters))
	seen := map[string]struct{}{}
	for _, cf := range cfg.Filters {
		for _, sf := range stashFilters {
			if sf.Id == cf.ID {
				name := cf.Name
				if name == "" {
					name = sf.Name
				}
				rows = append(rows, filterOverride{ID: sf.Id, SourceName: sf.Name, Name: name, Disabled: cf.Disabled})
				seen[sf.Id] = struct{}{}
				break
			}
		}
	}
	for _, s := range stashFilters {
		if _, ok := seen[s.Id]; ok {
			continue
		}
		rows = append(rows, filterOverride{ID: s.Id, SourceName: s.Name, Name: s.Name, Disabled: false})
	}

	return rows
}

func IndexHandler(libraryService *library.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var redactFunc func(string) string
		if !config.Application().IsRedactDisabled {
			redactFunc = config.Redacted
		}
		data := indexData{
			Redact:                  redactFunc,
			Version:                 build.FullVersion(),
			LogLevel:                config.Application().LogLevel,
			ForceHTTPS:              config.Application().ForceHTTPS,
			StashGraphQLUrl:         config.Application().StashGraphQLUrl,
			IsApiKeyProvided:        config.Application().StashApiKey != "",
			StashConnectionResponse: statusError,
		}

		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if version, err := stash.GetVersion(r.Context(), libraryService.StashClient); err != nil {
				var gqlErr *graphql.HTTPError
				if errors.As(err, &gqlErr) {
					if gqlErr.StatusCode == 401 {
						data.StashConnectionResponse = statusUnauthorized
					}
				}
				log.Ctx(r.Context()).Warn().Err(err).Msg("Failed to retrieve stash version")
			} else {
				data.StashConnectionResponse = statusOk
				data.StashData = &stashData{Version: version}
				data.StashData.FilterData, err = stashFilters(r.Context(), libraryService.StashClient)
				if err != nil {
					log.Ctx(r.Context()).Warn().Err(err).Msg("Failed to retrieve stash filters")
				} else {
					data.StashData.FilterOverrides = filterOverrideRows(r.Context(), data.StashData.FilterData)
				}
				data.StashData.SampleSceneCoverUrl, err = sampleSceneCoverUrl(r.Context(), libraryService.StashClient)
				if err != nil {
					log.Ctx(r.Context()).Warn().Err(err).Msg("Failed to retrieve sample scene cover url")
				}
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if sections, err := libraryService.GetSections(r.Context()); err != nil {
				log.Ctx(r.Context()).Warn().Err(err).Msg("Failed to retrieve sections")
			} else {
				data.SectionCount = len(sections)
				data.LinkCount = libraryService.Stats.Links
				data.SceneCount = libraryService.Stats.Scenes
			}
		}()

		wg.Wait()

		if err := indexTmpl.Execute(w, data); err != nil {
			log.Ctx(r.Context()).Err(err).Msg("index: execute template")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func FiltersUpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", 400)
			return
		}

		ids := r.PostForm["id"]
		if len(ids) == 0 {
			config.Save(r.Context(), config.UserConfig{})
		}

		sourceNames := r.PostForm["sourceName"]
		targetNames := r.PostForm["targetName"]
		disabled := r.PostForm["disabled"]

		disabledSet := make(map[string]struct{}, len(disabled))
		for _, id := range disabled {
			disabledSet[id] = struct{}{}
		}
		ovs := make([]config.Filter, 0, len(ids))
		for i, id := range ids {
			ov := config.Filter{ID: id}
			_, ov.Disabled = disabledSet[id]

			if targetNames[i] != "" && targetNames[i] != sourceNames[i] {
				ov.Name = targetNames[i]
			}

			ovs = append(ovs, ov)
		}

		cfg := config.User(r.Context())
		cfg.Filters = ovs

		config.Save(r.Context(), cfg)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
