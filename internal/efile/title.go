package efile

import (
	"fmt"
	"strings"
)

func MakeESceneId(sceneId string, oshash string) string {
	return fmt.Sprintf("%s %s", sceneId, oshash)
}
func GetSceneIdAndOshash(eSceneId string) (string, string, bool) {
	sceneId, oshash, isEScene := strings.Cut(eSceneId, " ")
	return sceneId, oshash, isEScene
}
