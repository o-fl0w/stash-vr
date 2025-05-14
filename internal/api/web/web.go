package web

import (
	_ "embed"
	"github.com/rs/zerolog/log"
	"html/template"
	"net/http"
	"stash-vr/internal/build"
	"stash-vr/internal/config"
	"stash-vr/internal/library"
	"strings"
)

//go:embed "index.html"
var indexTemplate string
var tmpl = template.Must(template.New("index").Parse(indexTemplate))

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
	IsApiKeyProvided        bool
	StashConnectionResponse string
	StashVersion            string
	SectionCount            int
	LinkCount               int
	SceneCount              int
}

func IndexHandler(libraryService *library.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := indexData{
			Redact:                  config.Redacted,
			Version:                 build.FullVersion(),
			LogLevel:                config.Get().LogLevel,
			ForceHTTPS:              config.Get().ForceHTTPS,
			StashGraphQLUrl:         config.Get().StashGraphQLUrl,
			IsApiKeyProvided:        config.Get().StashApiKey != "",
			StashConnectionResponse: fail,
		}

		if version, err := libraryService.GetClientVersions(r.Context()); err != nil {
			if strings.HasSuffix(err.Error(), "unauthorized") {
				data.StashConnectionResponse = unauthorized
			}
			log.Ctx(r.Context()).Warn().Err(err).Msg("Failed to retrieve stash version")
		} else {
			data.StashVersion = version["stash"]
			if sections, err := libraryService.GetSections(r.Context()); err != nil {
				log.Ctx(r.Context()).Warn().Err(err).Msg("Failed to retrieve sections")
			} else {
				data.StashConnectionResponse = ok
				data.SectionCount = len(sections)
				count := library.Count(sections)
				data.LinkCount = count.Links
				data.SceneCount = count.Scenes
			}
		}
		if err := tmpl.Execute(w, data); err != nil {
			log.Ctx(r.Context()).Err(err).Msg("index: execute template")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
