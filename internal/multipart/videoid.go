package multipart

import (
	"fmt"
	"hash/fnv"
	"strings"
)

// GetVideoIdFromScene returns the fuzzy videoId for a standard scene
func GetVideoIdFromScene(sceneId string) string {
	return sceneId
}

// GetVideoIdFromFile returns the fuzzy videoId for a specific file within a scene
func GetVideoIdFromFile(sceneId, fileId string) string {
	if fileId == "" {
		return sceneId
	}
	return sceneId + "_" + fileId
}

// ParseVideoId safely splits the fuzzy videoId back into its database parts
func ParseVideoId(videoId string) (sceneId, fileId string) {
	parts := strings.Split(videoId, "_")
	if len(parts) == 2 {
		return parts[0], parts[1] // Returns SceneID, FileID
	}
	return videoId, "" // Fallback for single-part scenes
}

// HashVideoId generates a safe numeric ID specifically required by DeoVR
func HashVideoId(videoId string) string {
	h := fnv.New32a()
	h.Write([]byte(videoId))
	return fmt.Sprintf("%d", h.Sum32()&0x7FFFFFFF)
}

// AppendQueryParam safely appends a dummy parameter for VR player cache busting
func AppendQueryParam(urlStr string, key, value string) string {
	if strings.Contains(urlStr, "?") {
		return urlStr + "&" + key + "=" + value
	}
	return urlStr + "?" + key + "=" + value
}
