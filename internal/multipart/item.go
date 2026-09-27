package multipart

import (
	"fmt"
	"stash-vr/internal/config"
	"stash-vr/internal/stash/gql"
)

type PlaybackItem struct {
	VideoId string
	FileId  string
	Label   string
	File    *gql.ScenePartsFilesVideoFile
}

// GetPlaybackItems handles the "expand feature" toggle natively.
func GetPlaybackItems(sceneId string, files []*gql.ScenePartsFilesVideoFile) []PlaybackItem {
	if len(files) == 0 {
		return []PlaybackItem{}
	}

	shouldExpand := config.Application().EnablePartsExpansion && len(files) > 1

	if shouldExpand {
		minDur, maxDur := files[0].Duration, files[0].Duration
		for _, f := range files[1:] {
			if f.Duration < minDur {
				minDur = f.Duration
			}
			if f.Duration > maxDur {
				maxDur = f.Duration
			}
		}
		if maxDur-minDur <= 1.0 {
			shouldExpand = false
		}
	}

	if !shouldExpand {
		return []PlaybackItem{{
			VideoId: GetVideoIdFromScene(sceneId),
			FileId:  "",
			Label:   "",
			File:    files[0],
		}}
	}

	sortedFiles, labels := GetFilesSortedByLabel(files)
	items := make([]PlaybackItem, len(sortedFiles))
	for i, f := range sortedFiles {
		items[i] = PlaybackItem{
			VideoId: GetVideoIdFromFile(sceneId, f.Id),
			FileId:  f.Id,
			Label:   labels[f.Id],
			File:    f,
		}
	}
	return items
}

// TargetItem safely retrieves the specific PlaybackItem requested.
func TargetItem(sceneId string, files []*gql.ScenePartsFilesVideoFile, fileId string) PlaybackItem {
	items := GetPlaybackItems(sceneId, files)
	if fileId != "" {
		for _, item := range items {
			if item.FileId == fileId {
				return item
			}
		}
	}
	if len(items) > 0 {
		return items[0]
	}
	return PlaybackItem{VideoId: GetVideoIdFromScene(sceneId), File: files[0]}
}

// FormatTitle standardizes how labels are appended to titles
func FormatTitle(title string, label string) string {
	if label != "" {
		return fmt.Sprintf("%s [%s]", title, label)
	}
	return title
}
