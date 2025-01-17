package stash

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"path/filepath"
	"regexp"
	"sort"
	"stash-vr/internal/config"
	"stash-vr/internal/stash/gql"
	"strconv"
	"strings"
)

type Stream struct {
	Name    string
	Sources []Source
}

type Source struct {
	Resolution int
	Url        string
}

var rgxResolution = regexp.MustCompile(`\((\d+)p\)`)

func GetStreams(ctx context.Context, fsp gql.StreamsParts, sortResolutionAsc bool) []Stream {
	streams := make([]Stream, 2)

	directStream := Stream{
		Name: "direct",
		Sources: []Source{{
			Resolution: fsp.Files[0].Height,
			Url:        fsp.Paths.Stream,
		}},
	}

	switch fsp.Files[0].Video_codec {
	case "h264", "hevc", "h265", "mpeg4", "av1":
		streams[0] = Stream{
			Name:    "transcoding",
			Sources: getSources(ctx, fsp, "MP4", "Direct stream", sortResolutionAsc),
		}
		streams[1] = directStream
	case "vp8", "vp9":
		streams[0] = Stream{
			Name:    "transcoding",
			Sources: getSources(ctx, fsp, "WEBM", "Direct stream", sortResolutionAsc),
		}
		streams[1] = directStream
	default:
		log.Ctx(ctx).Warn().Str("codec", fsp.Files[0].Video_codec).Str("file ext", filepath.Ext(fsp.Files[0].Path)).Msg("Codec not supported? Selecting transcoding sources.")
		streams[0] = Stream{
			Name: "transcoding",
			//transcode unsupported codecs to webm by default - or should we do mp4?
			Sources: getSources(ctx, fsp, "WEBM", "webm", sortResolutionAsc),
		}
	}

	// stash adds query parameter 'apikey' for direct stream but not for transcoding streams - add it
	if config.Get().StashApiKey != "" {
		for i, stream := range streams {
			for j, source := range stream.Sources {
				streams[i].Sources[j].Url = ApiKeyed(source.Url)
			}
		}
	}

	return streams
}

func parseResolutionFromLabel(label string) (int, error) {
	match := rgxResolution.FindStringSubmatch(label)
	if len(match) < 2 {
		return 0, fmt.Errorf("no resolution height found in label")
	}
	res, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("atoi: %w", err)
	}
	return res, nil
}

func getSources(ctx context.Context, sps gql.StreamsParts, format string, defaultSourceLabel string, sortResolutionAsc bool) []Source {
	sourceMap := make(map[int]Source)

	for _, s := range sps.SceneStreams {
		if strings.Contains(s.Label, format) {
			resolution, err := parseResolutionFromLabel(s.Label)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("label", s.Label).Msg("Failed to parse resolution from label")
				continue
			}

			if _, ok := sourceMap[resolution]; ok {
				continue
			}

			sourceMap[resolution] = Source{
				Resolution: resolution,
				Url:        s.Url,
			}
		} else if s.Label == defaultSourceLabel {
			sourceMap[sps.Files[0].Height] = Source{
				Resolution: sps.Files[0].Height,
				Url:        s.Url,
			}
		}
	}
	sources := make([]Source, 0, len(sourceMap))
	for _, v := range sourceMap {
		sources = append(sources, v)
	}
	sortSourcesByResolution(sources, sortResolutionAsc)
	return sources
}

func sortSourcesByResolution(sources []Source, asc bool) {
	if asc {
		sort.Slice(sources, func(i, j int) bool { return sources[i].Resolution < sources[j].Resolution })
	} else {
		sort.Slice(sources, func(i, j int) bool { return sources[i].Resolution > sources[j].Resolution })
	}
}
