package heresphere

import (
	"sort"
	"stash-vr/internal/config"
	"stash-vr/internal/library"
	"stash-vr/internal/prefix"
	"stash-vr/internal/util"
	"strings"
)

func getSummary(vd *library.VideoData, excludeAncestors bool) string {
	if len(vd.SceneParts.Tags) == 0 {
		return ""
	}

	m := make(map[string]string)
	for _, t := range vd.SceneParts.Tags {
		if t.Sort_name == config.Application().ExcludeSortName {
			continue
		}
		if excludeAncestors && strings.HasPrefix(t.Sort_name, prefix.SvrAncestor) {
			continue
		}
		m[t.Name] = util.FirstNonEmpty(&t.Sort_name, &t.Name)
	}

	type item struct {
		key     string
		sortKey string
	}

	items := make([]item, 0, len(m))
	for k, v := range m {
		items = append(items, item{key: k, sortKey: v})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].sortKey == items[j].sortKey {
			return items[i].key < items[j].key
		}
		return items[i].sortKey < items[j].sortKey
	})

	seen := make(map[string]struct{})
	keys := make([]string, 0, len(items))
	for _, it := range items {
		name := summaryStripper.ReplaceAllString(strings.ReplaceAll(it.key, " ", "_"), "")
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		keys = append(keys, name)
	}

	summary := strings.Join(keys, " | ")
	return summary
}
