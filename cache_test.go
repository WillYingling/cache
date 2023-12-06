package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	calls := 0
	fetcher := func(_ context.Context) (int, error) {
		calls++
		return calls, nil
	}
	cache := NewCache(fetcher, nil, true)

	val, err := cache.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, val)

	val, err = cache.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, val)

	cache.Invalidate()
	val, err = cache.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, 2, val)
}

func TestCache_NoCacheError(t *testing.T) {
	calls := 0
	fetcher := func(_ context.Context) (int, error) {
		calls++
		return calls, errors.New("error")
	}
	cache := NewCache(fetcher, nil, false)

	val, err := cache.Get(context.Background())
	require.Error(t, err)
	require.Equal(t, 1, val)

	val, err = cache.Get(context.Background())
	require.Error(t, err)
	require.Equal(t, 2, val)

}
