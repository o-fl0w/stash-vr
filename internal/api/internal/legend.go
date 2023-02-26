package internal

import "strings"

var (
	LegendTag       = newLegend("#", "Tag")
	LegendStudio    = newLegend("$", "Studio")
	LegendPerformer = newLegend("@", "Performer")
	LegendMovie     = newLegend("/", "Movie")
	LegendOCount    = newLegend("O", "O-Count")
	LegendOrganized = newLegend("Org", "Organized")
	LegendPlayCount = newLegend("P", "PlayCount")
)

type Legend struct {
	Short          string
	shortLowerCase string
	Full           string
	fullLowerCase  string
}

func (l Legend) IsMatch(s string) bool {
	s = strings.ToLower(s)
	return s == l.shortLowerCase || s == l.fullLowerCase
}

func newLegend(short string, full string) *Legend {
	return &Legend{
		Short:          short,
		shortLowerCase: strings.ToLower(short),
		Full:           full,
		fullLowerCase:  strings.ToLower(full),
	}
}
