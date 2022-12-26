package di

import (
	"context"
	"fmt"
)

func Set[K comparable, T any](ctx context.Context, k K, v T) context.Context {
	return context.WithValue(ctx, k, v)
}

func Get[K comparable, T any](ctx context.Context, k K, typeName string) (v T, err error) {
	vi := ctx.Value(k)
	if vi == nil {
		err = fmt.Errorf("%s not in context", typeName)
		return
	}

	var ok bool
	v, ok = vi.(T)
	if !ok {
		err = fmt.Errorf("value in context is not a %s", typeName)
		return
	}

	return v, nil
}
