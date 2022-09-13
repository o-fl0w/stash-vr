package heresphere

import "strings"

type Legend struct {
	Short         string
	Full          string
	fullLowerCase string
}

func NewLegend(short string, full string) *Legend {
	return &Legend{
		Short:         short,
		Full:          full,
		fullLowerCase: strings.ToLower(full),
	}
}

func (l Legend) IsMatch(s string) bool {
	s = strings.ToLower(s)
	return s == l.Short || s == l.fullLowerCase
}
