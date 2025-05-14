package util

func UnorderedEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	freq := make(map[T]int, len(a))
	for _, x := range a {
		freq[x]++
	}

	for _, y := range b {
		if freq[y] == 0 {
			return false
		}
		freq[y]--
	}
	return true
}
