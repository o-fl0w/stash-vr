package library

type Stats struct {
	Links  int
	Scenes int
}

func Count(sections []Section) Stats {
	var linkCount int
	sceneIds := make(map[string]struct{})
	for _, s := range sections {
		linkCount += len(s.Ids)
		for _, id := range s.Ids {
			sceneIds[id] = struct{}{}
		}
	}
	return Stats{
		Links:  linkCount,
		Scenes: len(sceneIds),
	}
}
