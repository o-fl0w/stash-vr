package efile

import (
	"fmt"
	"strings"
)

func getEFileSuffix(videoFileName string, eFileName string) string {
	videoFileTitle, _ := FileNameSplitExt(videoFileName)
	return strings.Replace(eFileName, strings.ToLower(videoFileTitle), "", 1)
}

func MakeESceneIdWithEFileName(sceneId string, videoFileName string, eFileName string) string {
	eFileSuffix := getEFileSuffix(videoFileName, eFileName)
	return MakeESceneIdWithEFileSuffix(sceneId, eFileSuffix)
}

func MakeESceneIdWithEFileSuffix(sceneId string, eFileSuffix string) string {
	return sceneId + " " + eFileSuffix
}

func GetSceneIdAndEFileSuffix(eSceneId string) (string, string, bool) {
	sceneId, eFileSuffix, isEScene := strings.Cut(eSceneId, " ")
	return sceneId, eFileSuffix, isEScene
}

func MakeESceneTitleWithEFileName(sceneTitle string, videoFileName string, eFileName string) string {
	eFileSuffix := getEFileSuffix(videoFileName, eFileName)
	return MakeESceneTitleWithEFileSuffix(sceneTitle, eFileSuffix)
}

func MakeESceneTitleWithEFileSuffix(sceneTitle string, eFileSuffix string) string {
	return fmt.Sprintf("*%s (%s)", sceneTitle, strings.TrimSpace(strings.TrimLeft(eFileSuffix, ".")))
}
