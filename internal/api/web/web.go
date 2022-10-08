package web

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"html/template"
	"net/http"
	"stash-vr/internal/application"
	"stash-vr/internal/cache"
	"stash-vr/internal/config"
	"stash-vr/internal/section"
	"stash-vr/internal/stash/gql"
	"strings"
)

var tmpl = template.Must(template.ParseFiles("web/template/index.html"))

const (
	ok           = "OK"
	fail         = "FAIL"
	unauthorized = "UNAUTHORIZED"
)

type indexData struct {
	Redact                  func(string) string
	Version                 string
	LogLevel                string
	ForceHTTPS              bool
	IsSyncMarkersAllowed    bool
	StashGraphQLUrl         string
	ApiKey                  string
	StashConnectionResponse string
	StashVersion            string
	SectionCount            int
	LinkCount               int
	SceneCount              int
}

func IndexHandler(client graphql.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := indexData{
			Redact:                  config.Redacted,
			Version:                 application.BuildVersion,
			LogLevel:                config.Get().LogLevel,
			ForceHTTPS:              config.Get().ForceHTTPS,
			IsSyncMarkersAllowed:    config.Get().IsSyncMarkersAllowed,
			StashGraphQLUrl:         config.Get().StashGraphQLUrl,
			ApiKey:                  config.Get().StashApiKey,
			StashConnectionResponse: fail,
		}

		if version, err := gql.Version(r.Context(), client); err == nil {
			data.StashConnectionResponse = ok
			data.StashVersion = version.Version.Version
			sections := cache.GetSections(r.Context(), client)
			data.SectionCount = len(sections)
			count := section.Count(sections)
			data.LinkCount = count.Links
			data.SceneCount = count.Scenes
		} else {
			if strings.HasSuffix(err.Error(), "unauthorized") {
				data.StashConnectionResponse = unauthorized
			}
			log.Ctx(r.Context()).Warn().Err(err).Msg("Failed to retrieve stash version")
		}
		if err := tmpl.Execute(w, data); err != nil {
			log.Ctx(r.Context()).Err(err).Msg("index: execute template")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
