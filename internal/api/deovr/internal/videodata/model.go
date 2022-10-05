package videodata

type VideoData struct {
	Authorized     string `json:"authorized"`
	FullAccess     bool   `json:"fullAccess"`
	Title          string `json:"title"`
	Id             string `json:"id"`
	VideoLength    int    `json:"videoLength"`
	Is3d           bool   `json:"is3d"`
	ScreenType     string `json:"screenType"`
	StereoMode     string `json:"stereoMode"`
	SkipIntro      int    `json:"skipIntro"`
	VideoThumbnail string `json:"videoThumbnail,omitempty"`
	VideoPreview   string `json:"videoPreview,omitempty"`
	ThumbnailUrl   string `json:"thumbnailUrl"`

	TimeStamps []TimeStamp `json:"timeStamps,omitempty"`

	Encodings []Encoding `json:"encodings"`
}

type TimeStamp struct {
	Ts   int    `json:"ts"`
	Name string `json:"name"`
}

type Encoding struct {
	Name         string        `json:"name"`
	VideoSources []VideoSource `json:"videoSources"`
}

type VideoSource struct {
	Resolution int    `json:"resolution"`
	Url        string `json:"url"`
}
