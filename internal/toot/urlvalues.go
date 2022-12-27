package toot

import (
	"net/url"
	"strconv"
)

func SetNonZero(q *url.Values, k string, v any) {
	if s, ok := v.(string); ok && s != "" {
		q.Set(k, s)
		return
	}
	if i, ok := v.(int); ok && i != 0 {
		q.Set(k, strconv.Itoa(i))
		return
	}
	if b, ok := v.(bool); ok && b {
		q.Set(k, "true")
		return
	}
	if ss, ok := v.([]string); ok {
		for _, s := range ss {
			q.Add(k, s)
		}
	}
}
