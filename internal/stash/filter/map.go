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

func getAsString(root map[string]any, keyPath string) (string, error) {
	s, err := get[string](root, keyPath)
	if err != nil {
		a, err := get[any](root, keyPath)
		if err != nil {
			return "", err
		}
		s = fmt.Sprintf("%v", a)
	}
	return s, nil
}

func get[T any](root map[string]any, keyPath string) (T, error) {
	var t T
	keys := strings.Split(keyPath, ".")
	key := keys[0]
	if len(keys) > 1 {
		m, ok := root[key].(map[string]any)
		if !ok {
			return t, errKeyNotFound{
				keys: keys,
				root: root,
			}
		}
		restPath, _ := strings.CutPrefix(keyPath, key+".")
		return get[T](m, restPath)
	}
	v, ok := root[key]
	if !ok {
		return t, errKeyNotFound{
			keys: keys,
			root: root,
		}
	}
	r, ok := v.(T)
	if !ok {
		return t, fmt.Errorf("expected '%T' but found '%T' for path '%v' in '%v'", t, v, keyPath, root)
	}
	return r, nil
}
