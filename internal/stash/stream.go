package stash

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
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

func GetStreams(ctx context.Context, fsp gql.SceneFullParts, sortResolutionAsc bool) []Stream {
	var streams []Stream

	original := Stream{
		Name: fsp.File.Video_codec,
		Sources: []Source{{
			Resolution: fsp.File.Height,
			Url:        fsp.Paths.Stream,
		}},
	}

	mp4Sources := getMp4Sources(ctx, fsp.StreamsParts)
	sortSourcesByResolution(mp4Sources, sortResolutionAsc)

	switch fsp.File.Video_codec {
	case "h264":
		streams = append(streams, Stream{
			Name:    "tc/h264",
			Sources: mp4Sources,
		})
		streams = append(streams, original)
	case "hevc", "h265":
		streams = append(streams, Stream{
			Name:    "tc/h265",
			Sources: mp4Sources,
		})
		streams = append(streams, original)
	default:
		log.Ctx(ctx).Debug().Str("codec", fsp.File.Video_codec).Msg("Codec not supported? Adding transcoded streams only")
		streams = append(streams, Stream{
			Name:    "tc/other",
			Sources: mp4Sources,
		})
	}

	// stash adds query parameter 'apikey' for direct stream but not for transcoded streams - add it
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

func getMp4Sources(ctx context.Context, sps gql.StreamsParts) []Source {
	sourceMap := make(map[int]Source)

	for _, s := range sps.SceneStreams {
		lowerCaseLabel := strings.ToLower(s.Label)

		if strings.Contains(lowerCaseLabel, "mp4") {
			resolution, err := parseResolutionFromLabel(lowerCaseLabel)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("label", lowerCaseLabel).Msg("Failed to parse resolution from label")
				continue
			}

			if _, ok := sourceMap[resolution]; ok {
				continue
			}

			sourceMap[resolution] = Source{
				Resolution: resolution,
				Url:        s.Url,
			}
		} else if lowerCaseLabel == "direct stream" {
			sourceMap[sps.File.Height] = Source{
				Resolution: sps.File.Height,
				Url:        s.Url,
			}
		}
	}
	sources := make([]Source, 0, len(sourceMap))
	for _, v := range sourceMap {
		sources = append(sources, v)
	}
	return sources
}

func sortSourcesByResolution(sources []Source, asc bool) {
	if asc {
		sort.Slice(sources, func(i, j int) bool { return sources[i].Resolution < sources[j].Resolution })
	} else {
		sort.Slice(sources, func(i, j int) bool { return sources[i].Resolution > sources[j].Resolution })
	}
}
