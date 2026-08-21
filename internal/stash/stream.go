package stash

import (
	"fmt"
	"regexp"
	"slices"
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
	Label      string
}

var rgxResolution = regexp.MustCompile(`\((\d+)p\)`)
var rgxContainer = regexp.MustCompile(`\((MP4|WebM|HLS|DASH)\)`)

func GetDirectStream(sp *gql.SceneParts) Stream {
	res := sp.Files[0].Height
	directStream := Source{
		Resolution: res,
		Url:        *sp.Paths.Stream,
		Label:      fmt.Sprintf("Direct-%dp", res),
	}

	return Stream{
		Name:    "direct",
		Sources: []Source{directStream},
	}
}

func GetTranscodingStreams(sp *gql.SceneParts) []Stream {
	streamsByType := make(map[string][]Source)

	for _, stream := range sp.SceneStreams {
		label := ""
		if stream.Label != nil {
			label = *stream.Label
		}
		if label == "Direct stream" {
			continue
		}

		resolution, err := parseResolutionFromLabel(label)
		if err != nil {
			resolution = sp.Files[0].Height
		}

		mimeType := ""
		if stream.Mime_type != nil {
			mimeType = *stream.Mime_type
		}

		container := parseContainerFromLabel(label)
		isDash := mimeType == "application/dash+xml" || strings.Contains(strings.ToUpper(label), "DASH")
		isMP4 := strings.Contains(strings.ToUpper(label), "MP4")

		if !isDash && !isMP4 {
			continue
		}

		// If matched by characteristics but not label regex, set container
		if container == "" {
			if isDash {
				container = "DASH"
			} else if isMP4 {
				container = "MP4"
			}
		}

		sourceLabel := fmt.Sprintf("%dp", resolution)
		if container != "" {
			sourceLabel = fmt.Sprintf("%s-%dp", container, resolution)
		}

		streamsByType[container] = append(streamsByType[container], Source{
			Resolution: resolution,
			Url:        stream.Url,
			Label:      sourceLabel,
		})
	}

	res := make([]Stream, 0, len(streamsByType))
	for container, sources := range streamsByType {
		slices.SortFunc(sources, func(a, b Source) int { return b.Resolution - a.Resolution })
		name := "transcoding"
		if container != "" {
			name = fmt.Sprintf("transcoding (%s)", container)
		}
		res = append(res, Stream{
			Name:    name,
			Sources: sources,
		})
	}

	slices.SortFunc(res, func(a, b Stream) int {
		if a.Name < b.Name {
			return -1
		}
		return 1
	})

	return res
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

func parseContainerFromLabel(label string) string {
	match := rgxContainer.FindStringSubmatch(label)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}
