package util

import (
	"slices"
	"strings"
)

func FirstNonEmpty(ss ...*string) string {
	for _, s := range ss {
		if s != nil && *s != "" {
			return *s
		}
	}
	return ""
}

func StrSliceEquals(s string, ss []string, v string) bool {
	v = strings.ToLower(v)
	return strings.ToLower(s) == v || slices.ContainsFunc(ss, func(el string) bool {
		return strings.ToLower(el) == v
	})
}
