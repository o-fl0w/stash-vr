package web

import (
	_ "embed"
	"errors"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"html/template"
	"net/http"
	"stash-vr/internal/build"
	"stash-vr/internal/config"
	"stash-vr/internal/library"
	"stash-vr/internal/stash/gql"
)

//go:embed "index.html"
var indexTemplate string
var tmpl = template.Must(template.New("index").Parse(indexTemplate))

const (
	ok           = "OK"
	fail         = "FAIL"
	unauthorized = "UNAUTHORIZED"
)

type filterData struct {
	Id   string
	Name string
}

type stashData struct {
	Version    string
	FilterData []filterData
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

func IndexHandler(libraryService *library.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var redactFunc func(string) string
		if !config.Get().IsRedactDisabled {
			redactFunc = config.Redacted
		}
		data := indexData{
			Redact:                  redactFunc,
			Version:                 build.FullVersion(),
			LogLevel:                config.Get().LogLevel,
			ForceHTTPS:              config.Get().ForceHTTPS,
			StashGraphQLUrl:         config.Get().StashGraphQLUrl,
			IsApiKeyProvided:        config.Get().StashApiKey != "",
			StashConnectionResponse: fail,
		}

		if version, err := libraryService.GetClientVersions(r.Context()); err != nil {
			var gqlErr *graphql.HTTPError
			if errors.As(err, &gqlErr) {
				if gqlErr.StatusCode == 401 {
					data.StashConnectionResponse = unauthorized
				}
			}
			log.Ctx(r.Context()).Warn().Err(err).Msg("Failed to retrieve stash version")
		} else {
			data.StashConnectionResponse = ok
			data.StashData = &stashData{Version: version["stash"]}
			resp, err := gql.FindSavedSceneFilters(r.Context(), libraryService.StashClient)
			if err == nil {
				for _, sf := range resp.FindSavedFilters {
					data.StashData.FilterData = append(data.StashData.FilterData, filterData{
						Id:   sf.Id,
						Name: sf.Name,
					})
				}
			}
		}
		if sections, err := libraryService.GetSections(r.Context()); err != nil {
			log.Ctx(r.Context()).Warn().Err(err).Msg("Failed to retrieve sections")
		} else {
			data.SectionCount = len(sections)
			count := library.Count(sections)
			data.LinkCount = count.Links
			data.SceneCount = count.Scenes
		}

		if err := tmpl.Execute(w, data); err != nil {
			log.Ctx(r.Context()).Err(err).Msg("index: execute template")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
