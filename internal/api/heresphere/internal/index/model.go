package index

type Index struct {
	Access  int       `json:"access"`
	Library []Library `json:"library"`
}

type VideoDataUrl string

type Library struct {
	Name string         `json:"name"`
	List []VideoDataUrl `json:"list"`
}
