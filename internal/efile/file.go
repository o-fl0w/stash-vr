package efile

import "path/filepath"

func FileNameSplitExt(fileName string) (string, string) {
	ext := filepath.Ext(fileName)
	return fileName[:len(fileName)-len(ext)], ext
}
