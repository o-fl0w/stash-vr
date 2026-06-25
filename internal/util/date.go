package util

import (
	"time"
)

func NormalizeDate(s string) string {
	var layout string

	switch len(s) {
	case 4:
		layout = "2006"
	case 7:
		layout = "2006-01"
	case 10:

		return s
	}

	t, err := time.Parse(layout, s)
	if err != nil {
		return s
	}

	return t.Format("2006-01-02")
}
