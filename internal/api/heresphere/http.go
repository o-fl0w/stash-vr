package heresphere

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/url"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/library"
	"stash-vr/internal/util"
	"strings"
)

type httpHandler struct {
	libraryService *library.Service
	vrTagId        string
}

func (h *httpHandler) indexHandler(w http.ResponseWriter, req *http.Request) {
	log.Ctx(req.Context()).Debug().Msg("INDEX")
	ctx := req.Context()
	baseUrl := internal.GetBaseUrl(req)

	sections, err := h.libraryService.GetSections(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get sections")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go func() {
		ctx := context.Background()
		_, err := h.libraryService.GetScenes(ctx)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to get scenes")
		}
	}()

	dto, err := buildIndex(sections, baseUrl)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to build index")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := internal.WriteJson(ctx, w, dto); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
	}
}

func (h *httpHandler) scanHandler(w http.ResponseWriter, req *http.Request) {
	log.Ctx(req.Context()).Debug().Msg("SCAN")
	ctx := req.Context()
	baseUrl := internal.GetBaseUrl(req)

	vds, err := h.libraryService.GetScenes(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get scenes")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dto, err := buildScan(ctx, vds, baseUrl)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to build scan")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := internal.WriteJson(ctx, w, dto); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
	}
}

func (h *httpHandler) videoDataHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	ctx := req.Context()
	baseUrl := internal.GetBaseUrl(req)
	videoId, err := url.QueryUnescape(chi.URLParam(req, "videoId"))
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("malformed videoId")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vdReq, err := internal.UnmarshalBody[videoDataRequestDto](req)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to parse request body")
		//w.WriteHeader(http.StatusBadRequest)
		//return
	}

	if vdReq.DeleteFile != nil && *vdReq.DeleteFile {
		if err = h.libraryService.Delete(ctx, videoId); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to delete scene")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	go func() {
		ctx := context.Background()
		if vdReq.Rating != nil {
			if err = h.libraryService.UpdateRating(ctx, videoId, *vdReq.Rating); err != nil {
				log.Ctx(ctx).Warn().Err(err).Float32("rating", *vdReq.Rating).Msg("Failed to update rating")
			}
		}
		if vdReq.IsFavorite != nil {
			if err = h.libraryService.UpdateFavorite(ctx, videoId, *vdReq.IsFavorite); err != nil {
				log.Ctx(ctx).Warn().Err(err).Bool("isFavorite", *vdReq.IsFavorite).Msg("Failed to update favorite")
			}
		}
		if vdReq.Tags != nil {
			newTags := make([]string, 0)
			newMarkers := make([]library.MarkerDto, 0)

			for _, t := range *vdReq.Tags {
				values := strings.Split(t.Name, ":")
				switch values[0] {
				case internal.LegendPerformer, internal.LegendSceneStudio, internal.LegendSceneGroup, internal.LegendMetaOCount,
					internal.LegendMetaOrganized, internal.LegendMetaPlayCount:
					continue
				case "!O", "!o":
					if err = h.libraryService.IncrementO(ctx, videoId); err != nil {
						log.Ctx(ctx).Warn().Err(err).Msg("Failed to increment O")
					}
				case internal.LegendTag:
					if len(values) > 1 && values[1] != "" {
						newTags = append(newTags, values[1])
					}
				default:
					if values[0] == "" {
						continue
					}
					m := library.MarkerDto{
						PrimaryTagName: values[0],
						StartSecond:    t.Start / 1000,
						MarkerId:       fmt.Sprintf("%.0f", *t.Rating),
					}
					if len(values) > 1 {
						m.Title = values[1]
					}
					if t.End != nil {
						m.EndSecond = util.Ptr(*t.End / 1000)
					}
					log.Ctx(ctx).Debug().Str("marker", fmt.Sprintf("%+v", m)).Msg("Incoming marker")
					newMarkers = append(newMarkers, m)
				}
			}
			log.Ctx(ctx).Debug().Bool("nilTags", newTags == nil).Int("tags", len(newTags)).Msg("About to update tags")
			if err = h.libraryService.UpdateTags(ctx, videoId, newTags); err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to update tags")
			}

			if err = h.libraryService.UpdateMarkers(ctx, videoId, newMarkers); err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to update markers")
			}
		}
		if vdReq.NeedsMediaSource != nil && *vdReq.NeedsMediaSource {
			if err = h.libraryService.IncrementPlayCount(ctx, videoId); err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to increment play count")
			}
		}

		_, err := h.libraryService.GetScene(ctx, videoId, true)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to refetch scene")
		}
	}()

	vd, err := h.libraryService.GetScene(ctx, videoId, false)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get scene")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dto, err := buildVideoData(vd, baseUrl)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to build video data")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := internal.WriteJson(ctx, w, dto); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
	}
}

func (h *httpHandler) eventsHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	event, err := internal.UnmarshalBody[playbackEvent](req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to parse request body")
		w.WriteHeader(http.StatusBadRequest)
	}

	log.Ctx(ctx).Trace().Str("event", fmt.Sprintf("%v", event)).Send()
	return
}

type videoDataRequestDto struct {
	Rating           *float32  `json:"rating,omitempty"`
	IsFavorite       *bool     `json:"isFavorite,omitempty"`
	Tags             *[]tagDto `json:"tags,omitempty"`
	DeleteFile       *bool     `json:"deleteFile,omitempty"`
	NeedsMediaSource *bool     `json:"needsMediaSource,omitempty"`
}
