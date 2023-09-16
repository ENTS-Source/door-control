package doors

import (
	"encoding/json"
	"fmt"
	"log"
)

func parseMessage(b []byte) map[string]any {
	m := make(map[string]any)
	if err := json.Unmarshal(b, &m); err != nil {
		log.Println(err)
	}
	return m
}

func readMessageVal[T any](m map[string]any, key string) (T, error) {
	var zero T
	if v, ok := m[key]; ok {
		if v2, ok := v.(T); ok {
			return v2, nil
		} else {
			return zero, fmt.Errorf("wrong type %T", v)
		}
	} else {
		return zero, nil
	}
}
