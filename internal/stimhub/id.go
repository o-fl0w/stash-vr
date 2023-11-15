package stimhub

import (
	"fmt"
	"strings"
)

func MakeStimSceneId(sceneId string, audioCrc32 string) string {
	return fmt.Sprintf("%s %s", sceneId, audioCrc32)
}
func SplitStimSceneId(stimSceneId string) (sceneId string, audioCrc32 string, isStimScene bool) {
	sceneId, audioCrc32, isStimScene = strings.Cut(stimSceneId, " ")
	return
}
