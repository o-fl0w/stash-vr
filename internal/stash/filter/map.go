package filter

import (
	"fmt"
	"strings"
)

type errKeyNotFound struct {
	keys []string
	root map[string]any
}

func (e errKeyNotFound) Error() string {
	return fmt.Sprintf("required key '%s' not found in '%v'", e.keys, e.root)
}

func get[T any](root map[string]any, keyPath string) (T, error) {
	keys := strings.Split(keyPath, ".")
	var t T
	m := root
	for i, key := range keys {
		v, ok := m[key]
		if !ok {
			break
		}
		if i == len(keys)-1 {
			//reached last key, return value
			if v == nil {
				return t, nil
			}
			r, ok := v.(T)
			if !ok {
				return t, fmt.Errorf("expected '%T' but found '%T' for path '%v' in '%v'", t, v, keyPath, root)
			}
			return r, nil
		}
		m, ok = v.(map[string]any)
		if !ok {
			return t, fmt.Errorf("expected 'any' but found '%T' for '%v' in path '%v' in '%v'", v, key, keyPath, m)
		}
	}
	return t, errKeyNotFound{
		keys: keys,
		root: root,
	}
}
