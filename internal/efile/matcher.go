package efile

import "strings"

func FindAllMatchingEFileNames(videoFileName string, eFileNames []string) []string {
	videoFileTitle, _ := FileNameSplitExt(videoFileName)
	var matches []string
	for _, eFileName := range eFileNames {
		if strings.HasPrefix(eFileName, strings.ToLower(videoFileTitle)) {
			matches = append(matches, eFileName)
		}
	}
	return matches
}
