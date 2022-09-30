package heresphere

import "strings"

type Legend struct {
	Short          string
	shortLowerCase string
	Full           string
	fullLowerCase  string
}

func NewLegend(short string, full string) *Legend {
	return &Legend{
		Short:          short,
		shortLowerCase: strings.ToLower(short),
		Full:           full,
		fullLowerCase:  strings.ToLower(full),
	}
}

func (l Legend) IsMatch(s string) bool {
	s = strings.ToLower(s)
	return s == l.shortLowerCase || s == l.fullLowerCase
}
