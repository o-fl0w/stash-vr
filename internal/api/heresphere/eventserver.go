package heresphere

type event int

const (
	open event = iota
	play
	pause
	close
)

func (e event) String() string {
	switch e {
	case open:
		return "open"
	case play:
		return "play"
	case pause:
		return "pause"
	case close:
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
