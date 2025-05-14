package stash

import (
	"fmt"
	"regexp"
	"sort"
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

func GetStreams(sp *gql.SceneParts) []Stream {
	sourcesByName := make(map[string][]Source)
	for _, stream := range sp.SceneStreams {
		if *stream.Label == "Direct stream" || !strings.HasPrefix(*stream.Mime_type, "video/mp4") {
			continue
		}

		resolution, err := parseResolutionFromLabel(*stream.Label)
		if err != nil {
			resolution = sp.Files[0].Height
		}
		sourcesByName[*stream.Mime_type] = append(sourcesByName[*stream.Mime_type], Source{
			Resolution: resolution,
			Url:        stream.Url,
		})
	}

	streams := make([]Stream, 0)
	for name, sources := range sourcesByName {
		sort.Slice(sources, func(i, j int) bool { return sources[i].Resolution > sources[j].Resolution })
		streams = append(streams, Stream{
			Name:    name,
			Sources: sources,
		})
	}

	directStream := Source{
		Resolution: sp.Files[0].Height,
		Url:        *sp.Paths.Stream,
	}

	streams = append(streams, Stream{
		Name:    "direct",
		Sources: []Source{directStream},
	})

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
