package heresphere

type event int

const (
	evOpen event = iota
	evPlay
	evPause
	evClose
)

func (e event) String() string {
	switch e {
	case evOpen:
		return "open"
	case evPlay:
		return "play"
	case evPause:
		return "pause"
	case evClose:
		return "close"
	}
	return "unknown event"
}

type playbackEvent struct {
	Username      string  `json:"username,omitempty"`
	Id            string  `json:"id,omitempty"`
	Title         string  `json:"title,omitempty"`
	Event         event   `json:"event,omitempty"`
	Time          float32 `json:"time,omitempty"`
	Speed         float32 `json:"speed,omitempty"`
	Utc           float64 `json:"utc,omitempty"`
	ConnectionKey string  `json:"connectionKey,omitempty"`
}
