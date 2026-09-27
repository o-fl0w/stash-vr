package multipart

import (
	"fmt"
	"slices"
	"sort"
	"stash-vr/internal/stash/gql"
	"strings"
)

func GetFilesSortedByLabel(files []*gql.ScenePartsFilesVideoFile) ([]*gql.ScenePartsFilesVideoFile, map[string]string) {
	labels := getFileLabels(files)

	sorted := make([]*gql.ScenePartsFilesVideoFile, len(files))
	copy(sorted, files)

	slices.SortFunc(sorted, func(a, b *gql.ScenePartsFilesVideoFile) int {
		return strings.Compare(labels[a.Id], labels[b.Id])
	})

	return sorted, labels
}

func getFileLabels(files []*gql.ScenePartsFilesVideoFile) map[string]string {
	labelsMap := make(map[string]string, len(files))

	if len(files) == 0 {
		return labelsMap
	}
	if len(files) == 1 {
		labelsMap[files[0].Id] = ""
		return labelsMap
	}

	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.Basename
	}

	extracted := ExtractLabels(names)
	for i, f := range files {
		labelsMap[f.Id] = extracted[i]
	}

	return labelsMap
}

func ExtractLabels(fileNames []string) []string {
	if len(fileNames) <= 1 {
		return []string{""}
	}

	minLen := len(fileNames[0])
	for _, name := range fileNames[1:] {
		if len(name) < minLen {
			minLen = len(name)
		}
	}

	prefixLen := 0
	for i := 0; i < minLen; i++ {
		char := fileNames[0][i]
		match := true
		for _, name := range fileNames[1:] {
			if name[i] != char {
				match = false
				break
			}
		}
		if !match {
			break
		}
		prefixLen++
	}

	maxSuffixLen := minLen - prefixLen
	suffixLen := 0
	for i := 0; i < maxSuffixLen; i++ {
		char := fileNames[0][len(fileNames[0])-1-i]
		match := true
		for _, name := range fileNames[1:] {
			if name[len(name)-1-i] != char {
				match = false
				break
			}
		}
		if !match {
			break
		}
		suffixLen++
	}

	labels := make([]string, len(fileNames))
	labelsValid := true

	for i, name := range fileNames {
		label := name[prefixLen : len(name)-suffixLen]
		label = strings.Trim(label, "-_ ")
		labels[i] = label

		if len(label) == 0 || len(label) > 3 {
			labelsValid = false
		}
	}

	if !labelsValid {
		indices := make([]int, len(fileNames))
		for i := range indices {
			indices[i] = i
		}
		sort.SliceStable(indices, func(i, j int) bool {
			return fileNames[indices[i]] < fileNames[indices[j]]
		})

		for i, originalIndex := range indices {
			labels[originalIndex] = fmt.Sprintf("%03d", i+1)
		}
	}

	return labels
}
