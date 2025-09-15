package heresphere

import (
	"hash/fnv"
	"sort"
	"stash-vr/internal/config"
	"stash-vr/internal/library"
	"stash-vr/internal/util"
	"strconv"
	"strings"
	"unicode"
)

func getSummary(vd *library.VideoData) string {
	if len(vd.SceneParts.Tags) == 0 {
		return ""
	}

	m := make(map[string]string)
	for _, t := range vd.SceneParts.Tags {
		if t.Sort_name == config.Application().ExcludeSortName {
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

func tokenizeWords(s string) [][]rune {
	words := make([][]rune, 0, 4)
	cur := make([]rune, 0, 8)
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur = append(cur, unicode.ToLower(r))
		} else if len(cur) > 0 {
			words = append(words, cur)
			cur = make([]rune, 0, 8)
		}
	}
	if len(cur) > 0 {
		words = append(words, cur)
	}
	return words
}

func joinPieces(pieces [][]rune, limit int) string {
	total := 0
	for _, p := range pieces {
		total += len(p)
		if total >= limit {
			total = limit
			break
		}
	}
	out := make([]rune, 0, total)
	for _, p := range pieces {
		if len(out) >= limit {
			break
		}
		need := limit - len(out)
		if len(p) > need {
			out = append(out, p[:need]...)
			break
		}
		out = append(out, p...)
	}
	return string(out)
}
func mnemonicPrefix(s string, mnemoLen int) string {
	if mnemoLen <= 0 {
		mnemoLen = 3
	}
	words := tokenizeWords(s)

	if len(words) == 0 {
		out := make([]rune, 0, mnemoLen)
		firstDone := false
		for _, r := range strings.ToLower(s) {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				continue
			}
			if !firstDone {
				out = append(out, unicode.ToUpper(r))
				firstDone = true
			} else {
				out = append(out, r)
			}
			if len(out) == mnemoLen {
				break
			}
		}
		return string(out)
	}

	pieces := make([][]rune, len(words))
	total := 0
	for i, w := range words {
		if total == mnemoLen {
			break
		}
		if len(w) == 0 {
			continue
		}
		pieces[i] = append(pieces[i], unicode.ToUpper(w[0]))
		words[i] = w[1:]
		total++
	}
	if total == mnemoLen {
		return joinPieces(pieces, mnemoLen)
	}

	for total < mnemoLen {
		progress := false
		for i := 0; i < len(words) && total < mnemoLen; i++ {
			if len(words[i]) == 0 {
				continue
			}
			pieces[i] = append(pieces[i], words[i][0])
			words[i] = words[i][1:]
			total++
			progress = true
		}
		if !progress {
			break
		}
	}
	return joinPieces(pieces, mnemoLen)
}

func tinyHash(s string, hashChars int) string {
	if hashChars <= 0 {
		return ""
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(strings.TrimSpace(strings.ToLower(s))))
	base := strconv.FormatUint(h.Sum64(), 36)
	if len(base) >= hashChars {
		return base[len(base)-hashChars:]
	}
	return strings.Repeat("0", hashChars-len(base)) + base
}

func MnemonicID(s string, mnemoLen, hashChars int) string {
	return mnemonicPrefix(s, mnemoLen) + "-" + tinyHash(s, hashChars)
}
