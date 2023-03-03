package title

func GetSceneTitle(title string, basename string) string {
	if title != "" {
		return title
	}
	return basename
}
