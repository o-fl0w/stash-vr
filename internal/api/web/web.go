package web

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"html/template"
	"net/http"
	"stash-vr/internal/application"
	"stash-vr/internal/config"
	"stash-vr/internal/section"
	"stash-vr/internal/stash/gql"
	"strings"
)

var tmpl = template.Must(template.ParseFiles("web/template/index.html"))

const (
	OK           = "OK"
	FAIL         = "FAIL"
	UNAUTHORIZED = "UNAUTHORIZED"
)

type IndexData struct {
	Version                 string
	LogLevel                string
	ForceHTTPS              bool
	IsSyncMarkersAllowed    bool
	StashGraphQLUrl         string
	IsApiKeyProvided        bool
	StashConnectionResponse string
	StashVersion            string
	SectionCount            int
	LinkCount               int
	SceneCount              int
}

func ServeIndex(client graphql.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := IndexData{
			Version:                 application.BuildVersion,
			LogLevel:                config.Get().LogLevel,
			ForceHTTPS:              config.Get().ForceHTTPS,
			IsSyncMarkersAllowed:    config.Get().IsSyncMarkersAllowed,
			StashGraphQLUrl:         config.Get().StashGraphQLUrl,
			IsApiKeyProvided:        config.Get().StashApiKey != "",
			StashConnectionResponse: FAIL,
		}

		if version, err := gql.Version(r.Context(), client); err == nil {
			data.StashConnectionResponse = OK
			data.StashVersion = version.Version.Version
			sections := section.Get(r.Context(), client)
			data.SectionCount = len(sections)
			count := section.Count(sections)
			data.LinkCount = count.Links
			data.SceneCount = count.Scenes
		} else {
			if strings.HasSuffix(err.Error(), "unauthorized") {
				data.StashConnectionResponse = UNAUTHORIZED
			}
			log.Ctx(r.Context()).Warn().Err(err).Msg("Failed to retrieve stash version")
		}
		if err := tmpl.Execute(w, data); err != nil {
			log.Ctx(r.Context()).Err(err).Msg("index: execute template")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func ServeStatic() http.HandlerFunc {
	filesDir := http.Dir("./web/static")
	return func(w http.ResponseWriter, r *http.Request) {
		rCtx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rCtx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(filesDir))
		fs.ServeHTTP(w, r)
	}
}
