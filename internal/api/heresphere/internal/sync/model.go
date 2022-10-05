package sync

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/api/common"
	"stash-vr/internal/api/heresphere/internal/videodata"
	"stash-vr/internal/stash"
	"strings"
)

type UpdateVideoData struct {
	Rating     *float32         `json:"rating,omitempty"`
	IsFavorite *bool            `json:"isFavorite,omitempty"`
	Tags       *[]videodata.Tag `json:"tags,omitempty"`
	DeleteFile *bool            `json:"deleteFile,omitempty"`
}

func (v UpdateVideoData) IsUpdateRequest() bool {
	return v.Rating != nil || v.IsFavorite != nil || v.Tags != nil
}

func (v UpdateVideoData) IsDeleteRequest() bool {
	return v.DeleteFile != nil && *v.DeleteFile
}

type requestDetails struct {
	tagIds          []string
	studioId        string
	performerIds    []string
	markers         []marker
	incrementO      bool
	toggleOrganized bool
}

type marker struct {
	tag   string
	title string
	start float64
}

func parseUpdateRequestTags(ctx context.Context, client graphql.Client, tags []videodata.Tag) requestDetails {
	request := requestDetails{}

	for _, tagReq := range tags {
		if strings.HasPrefix(tagReq.Name, "!") {
			cmd := tagReq.Name[1:]
			if common.LegendOCount.IsMatch(cmd) {
				request.incrementO = true
				continue
			} else if common.LegendOrganized.IsMatch(cmd) {
				request.toggleOrganized = true
				continue
			}
		}

		tagType, tagName, isCategorized := strings.Cut(tagReq.Name, ":")

		if isCategorized && common.LegendTag.IsMatch(tagType) {
			if tagName == "" {
				log.Ctx(ctx).Trace().Str("request", tagReq.Name).Msg("Empty tag name, skipping")
				continue
			}
			id, err := stash.FindOrCreateTag(ctx, client, tagName)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("request", tagReq.Name).Msg("Failed to find or create tag")
				continue
			}
			request.tagIds = append(request.tagIds, id)
		} else if isCategorized && common.LegendStudio.IsMatch(tagType) {
			if tagName == "" {
				continue
			}
			id, err := stash.FindOrCreateStudio(ctx, client, tagName)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("request", tagReq.Name).Msg("Failed to find or create studio")
				continue
			}
			request.studioId = id
		} else if isCategorized && common.LegendPerformer.IsMatch(tagType) {
			if tagName == "" {
				log.Ctx(ctx).Trace().Str("request", tagReq.Name).Msg("Empty performer name, skipping")
				continue
			}
			id, err := stash.FindOrCreatePerformer(ctx, client, tagName)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("request", tagReq.Name).Msg("Failed to find or create performer")
				continue
			}
			request.performerIds = append(request.performerIds, id)
		} else if isCategorized && (common.LegendMovie.IsMatch(tagType) || common.LegendOCount.IsMatch(tagType) || common.LegendOrganized.IsMatch(tagType)) {
			log.Ctx(ctx).Trace().Str("request", tagReq.Name).Msg("Tag type is reserved, skipping")
			continue
		} else {
			var markerTitle string
			markerPrimaryTag := tagType
			if isCategorized {
				markerTitle = tagName
			}
			request.markers = append(request.markers, marker{
				tag:   markerPrimaryTag,
				title: markerTitle,
				start: float64(tagReq.Start) / 1000,
			})
		}
	}

	return request
}
