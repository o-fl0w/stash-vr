package config

import (
	"fmt"
)

func Redacted(s string) string {
	if Get().IsRedactDisabled {
		return s
	} else {
		return fmt.Sprintf("REDACTED(%d)", len(s))
	}
}
