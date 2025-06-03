package heresphere

import (
	"fmt"
	"stash-vr/internal/api/heatmap"
	"stash-vr/internal/config"
	"stash-vr/internal/library"
	"stash-vr/internal/stash"
	"stash-vr/internal/util"
	"time"
)

type videoDataDto struct {
	Access int `json:"access"`

	Title string `json:"title"`
	//Description    string      `json:"description,omitempty"`
	ThumbnailImage *string       `json:"thumbnailImage,omitempty"`
	ThumbnailVideo *string       `json:"thumbnailVideo,omitempty"`
	DateReleased   *string       `json:"dateReleased,omitempty"`
	DateAdded      string        `json:"dateAdded,omitempty"`
	Duration       float64       `json:"duration,omitempty"`
	Rating         *float32      `json:"rating,omitempty"`
	Favorites      *int          `json:"favorites,omitempty"`
	IsFavorite     *bool         `json:"isFavorite,omitempty"`
	EventServer    *string       `json:"eventServer,omitempty"`
	Scripts        []scriptDto   `json:"scripts,omitempty"`
	Tags           []tagDto      `json:"tags,omitempty"`
	Media          []mediaDto    `json:"media,omitempty"`
	Subtitles      []subtitleDto `json:"subtitles,omitempty"`

	WriteFavorite *bool `json:"writeFavorite,omitempty"`
	WriteRating   *bool `json:"writeRating,omitempty"`
	WriteTags     *bool `json:"writeTags,omitempty"`
}

type mediaDto struct {
	Name    string      `json:"name,omitempty"`
	Sources []sourceDto `json:"sources,omitempty"`
}

type sourceDto struct {
	Resolution int    `json:"resolution,omitempty"`
	Url        string `json:"url,omitempty"`
}

type scriptDto struct {
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
}

type subtitleDto struct {
	Name     string `json:"name,omitempty"`
	Language string `json:"language,omitempty"`
	Url      string `json:"url,omitempty"`
}

func buildVideoData(vd *library.VideoData, baseUrl string) (*videoDataDto, error) {
	videoId := vd.Id()
	if len(vd.SceneParts.Files) == 0 {
		return nil, fmt.Errorf("scene %s has no files", videoId)
	}

	dto := videoDataDto{
		Access:        1,
		Title:         vd.Title(),
		DateAdded:     vd.SceneParts.Created_at.Format(time.DateOnly),
		Duration:      vd.SceneParts.Files[0].Duration * 1000,
		WriteFavorite: util.Ptr(true),
		WriteRating:   util.Ptr(true),
		WriteTags:     util.Ptr(true),
		EventServer:   util.Ptr(getEventsUrl(baseUrl, videoId)),
	}

	if vd.SceneParts.Paths.Screenshot != nil {
		if vd.SceneParts.Interactive && vd.SceneParts.Paths.Interactive_heatmap != nil {
			dto.ThumbnailImage = util.Ptr(heatmap.GetCoverUrl(baseUrl, videoId))
		} else {
			dto.ThumbnailImage = util.Ptr(stash.ApiKeyed(*vd.SceneParts.Paths.Screenshot))
		}
	}

	if vd.SceneParts.Paths.Preview != nil {
		dto.ThumbnailVideo = util.Ptr(stash.ApiKeyed(*vd.SceneParts.Paths.Preview))
	}

	if vd.SceneParts.Date != nil {
		dto.DateReleased = vd.SceneParts.Date
	}

	if vd.SceneParts.Rating100 != nil {
		dto.Rating = util.Ptr(float32(*vd.SceneParts.Rating100) / 20)
	}

	if vd.SceneParts.O_counter != nil {
		dto.Favorites = vd.SceneParts.O_counter
	}

	if isFavorite(vd) {
		dto.IsFavorite = util.Ptr(true)
	}

	setMediaSources(vd, &dto)

	setScripts(vd, &dto)

	setSubtitles(vd, &dto)

	dto.Tags = getTags(vd)

	return &dto, nil
}

func setSubtitles(vd *library.VideoData, dto *videoDataDto) {
	if vd.SceneParts.Captions == nil {
		return
	}
	for _, c := range vd.SceneParts.Captions {
		dto.Subtitles = append(dto.Subtitles, subtitleDto{
			Name:     fmt.Sprintf("%s.%s", c.Language_code, c.Caption_type),
			Language: c.Language_code,
			Url:      stash.ApiKeyed(fmt.Sprintf("%s?lang=%s&type=%s", *vd.SceneParts.Paths.Caption, c.Language_code, c.Caption_type)),
		})
	}
}

func isFavorite(vd *library.VideoData) bool {
	for _, t := range vd.SceneParts.Tags {
		if t.Name == config.Get().FavoriteTag {
			return true
		}
	}
	return false
}

func setScripts(vd *library.VideoData, dto *videoDataDto) {
	if !vd.SceneParts.Interactive {
		return
	}
	dto.Scripts = append(dto.Scripts, scriptDto{
		Name: "Script-" + vd.Title(),
		Url:  stash.ApiKeyed(*vd.SceneParts.Paths.Funscript),
	})
}

func setMediaSources(vd *library.VideoData, dto *videoDataDto) {
	for _, stream := range stash.GetStreams(vd.SceneParts) {
		e := mediaDto{
			Name: stream.Name,
		}
		for _, s := range stream.Sources {
			vs := sourceDto{
				Resolution: s.Resolution,
				Url:        s.Url,
			}
			e.Sources = append(e.Sources, vs)
		}
		dto.Media = append(dto.Media, e)
	}
}
