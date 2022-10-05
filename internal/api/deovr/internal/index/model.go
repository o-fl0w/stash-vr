package index

type Index struct {
	Authorized string  `json:"authorized"`
	Scenes     []Scene `json:"scenes"`
}

type Scene struct {
	Name string        `json:"name"`
	List []PreviewData `json:"list"`
}

type PreviewData struct {
	Id           string `json:"id"`
	ThumbnailUrl string `json:"thumbnailUrl"`
	Title        string `json:"title"`
	VideoLength  int    `json:"videoLength"`
	VideoUrl     string `json:"video_url"`
}
