package web

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"html/template"
	"net/http"
	"stash-vr/internal/api/common"
	"stash-vr/internal/api/common/section"
	"stash-vr/internal/application"
	"stash-vr/internal/config"
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
			StashGraphQLUrl:         config.Get().StashGraphQLUrl,
			IsApiKeyProvided:        config.Get().StashApiKey != "",
			StashConnectionResponse: FAIL,
		}

		if version, err := gql.Version(r.Context(), client); err == nil {
			data.StashConnectionResponse = OK
			data.StashVersion = version.Version.Version
			sections := common.GetIndex(r.Context(), client)
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
