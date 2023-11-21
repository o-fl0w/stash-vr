package web

import (
	_ "embed"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"html/template"
	"net/http"
	"stash-vr/internal/build"
	"stash-vr/internal/config"
	"stash-vr/internal/sections"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/stimhub"
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

func IndexHandler(stashClient graphql.Client, stimhubClient *stimhub.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := indexData{
			Redact:                  config.Redacted,
			Version:                 build.FullVersion(),
			LogLevel:                config.Get().LogLevel,
			ForceHTTPS:              config.Get().ForceHTTPS,
			IsSyncMarkersAllowed:    config.Get().IsSyncMarkersAllowed,
			StashGraphQLUrl:         config.Get().StashGraphQLUrl,
			IsApiKeyProvided:        config.Get().StashApiKey != "",
			StashConnectionResponse: fail,
		}

		if version, err := gql.Version(r.Context(), stashClient); err == nil {
			data.StashConnectionResponse = ok
			data.StashVersion = version.Version.Version
			ss := sections.Get(r.Context(), stashClient, stimhubClient)
			data.SectionCount = len(ss)
			count := sections.Count(ss)
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
