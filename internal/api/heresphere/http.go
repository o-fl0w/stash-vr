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
	"stash-vr/internal/stash"
	"stash-vr/internal/util"
	"strings"
)

type httpHandler struct {
	libraryService *library.Service
	ps             *playbackState
}

var minPlayFraction *float64

func (h *httpHandler) indexHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	baseUrl := internal.GetBaseUrl(req)

	mpf := stash.GetMinPlayPercent(ctx, h.libraryService.StashClient) / 100
	minPlayFraction = &mpf

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

	if vdReq, err := internal.UnmarshalBody[videoDataRequestDto](req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to parse request body")
	} else {
		if vdReq.DeleteFile != nil && *vdReq.DeleteFile {
			if err = h.libraryService.Delete(ctx, videoId); err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to delete scene")
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		go h.processUpdates(videoId, vdReq)
	}

	vd, err := h.libraryService.GetScene(ctx, videoId, false)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get scene")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dto, err := buildVideoData(ctx, vd, baseUrl)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to build video data")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := internal.WriteJson(ctx, w, dto); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
	}
}

func (h *httpHandler) processUpdates(videoId string, vdReq videoDataRequestDto) {
	ctx := context.Background()
	needsRefetch := false
	if vdReq.Rating != nil {
		if err := h.libraryService.UpdateRating(ctx, videoId, vdReq.Rating); err != nil {
			log.Ctx(ctx).Warn().Err(err).Float32("rating", *vdReq.Rating).Msg("Failed to update rating")
		}
		needsRefetch = true
	}
	if vdReq.IsFavorite != nil {
		if err := h.libraryService.UpdateFavorite(ctx, videoId, *vdReq.IsFavorite); err != nil {
			log.Ctx(ctx).Warn().Err(err).Bool("isFavorite", *vdReq.IsFavorite).Msg("Failed to update favorite")
		}
		needsRefetch = true
	}
	if vdReq.Tags != nil {
		h.processIncomingTags(ctx, videoId, vdReq)
		needsRefetch = true
	}
	if needsRefetch {
		_, err := h.libraryService.GetScene(ctx, videoId, true)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to refetch scene")
		}
	}
}

func (h *httpHandler) processIncomingTags(ctx context.Context, videoId string, vdReq videoDataRequestDto) {
	newTags := make([]string, 0)
	newMarkers := make([]library.MarkerDto, 0)

	hasPlayCount := false
	hasOrganized := false
	hasOCount := false
	hasRating := false

	for _, t := range *vdReq.Tags {
		key, arg, _ := strings.Cut(t.Name, ":")

		if key == "" {
			continue
		}

		switch key {
		case internal.LegendPerformer, internal.LegendSceneStudio, internal.LegendSceneGroup,
			internal.LegendMetaResolution, internal.LegendSummary, internal.LegendSummaryId:
			continue
		case internal.LegendMetaOCount:
			hasOCount = true
			continue
		case internal.LegendMetaOrganized:
			hasOrganized = true
			continue
		case internal.LegendMetaPlayCount:
			hasPlayCount = true
			continue
		case internal.LegendMetaRating:
			hasRating = true
			continue
		}

		if strings.HasPrefix(key, internal.LegendTag) {
			if key == internal.LegendTag && arg != "" && arg[0] != '#' {
				newTags = append(newTags, arg)
			}
			continue
		}

		if strings.EqualFold(key, internal.CommandIncrementO) {
			if err := h.libraryService.IncrementO(ctx, videoId); err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to increment O")
			}
			continue
		}
		if strings.EqualFold(key, internal.CommandSetOrganizedTrue) {
			if err := h.libraryService.SetOrganized(ctx, videoId, true); err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to set organized=true")
			}
			continue
		}

		m := library.MarkerDto{
			PrimaryTagName: key,
			StartSecond:    t.Start / 1000,
			MarkerId:       fmt.Sprintf("%.0f", *t.Rating),
		}
		if arg != "" {
			m.Title = arg
		}
		if t.End != nil {
			m.EndSecond = util.Ptr(*t.End / 1000)
		}
		newMarkers = append(newMarkers, m)
	}

	if !hasPlayCount {
		if err := h.libraryService.DecrementPlayCount(ctx, videoId); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to decrement play count")
		}
	}

	if !hasOrganized {
		if err := h.libraryService.SetOrganized(ctx, videoId, false); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to set organized=false")
		}
	}

	if !hasOCount {
		if err := h.libraryService.DecrementO(ctx, videoId); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to decrement O")
		}
	}

	if !hasRating {
		if err := h.libraryService.UpdateRating(ctx, videoId, nil); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to set zero rating")
		}
	}

	if err := h.libraryService.UpdateTags(ctx, videoId, newTags); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to update tags")
	}

	if err := h.libraryService.UpdateMarkers(ctx, videoId, newMarkers); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to update markers")
	}
}

func (h *httpHandler) eventsHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	ev, err := internal.UnmarshalBody[playbackEvent](req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to parse event body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parts := strings.Split(ev.Id, "/")
	videoId := parts[len(parts)-1]
	vd, err := h.libraryService.GetScene(ctx, videoId, false)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to get scene from event")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Ctx(ctx).Debug().Str("id", ev.Id).Str("event", ev.Event.String()).Send()

	switch ev.Event {
	case evPlay:
		if h.ps == nil {
			h.ps = newPlayback(vd)
		} else if h.ps.videoId != videoId {
			h.ps.handleStop(ctx, h.libraryService, minPlayFraction)
			h.ps = newPlayback(vd)
		} else {
			h.ps.handleResume()
		}
	case evPause, evClose:
		if h.ps != nil {
			h.ps.handleStop(ctx, h.libraryService, minPlayFraction)
		}
	default:
	}
}

type videoDataRequestDto struct {
	Rating           *float32  `json:"rating,omitempty"`
	IsFavorite       *bool     `json:"isFavorite,omitempty"`
	Tags             *[]tagDto `json:"tags,omitempty"`
	DeleteFile       *bool     `json:"deleteFile,omitempty"`
	NeedsMediaSource *bool     `json:"needsMediaSource,omitempty"`
}
