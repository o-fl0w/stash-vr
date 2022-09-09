package util

import "fmt"

func Redacted(s string) string {
	return fmt.Sprintf("REDACTED(%d)", len(s))
}
