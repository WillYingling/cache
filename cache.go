package cache

import (
	"context"
	"sync"
)

// Validator defines an interface that decides whether or not a cache needs to be refreshed.
type Validator interface {
	ShouldFetch() bool
	OnFetch()
}

// Fetcher is any function that returns a value with a potential error
type Fetcher[T any] func(context.Context) (T, error)

// Cache is an object which will store previous values returned
// by a fetcher, and refresh those values according to the provided
// validator. It is thread-safe.
type Cache[T any] struct {
	lastVal    T
	lastErr    error
	cacheError bool

	mux       sync.Mutex
	validator Validator
	fetcher   Fetcher[T]
	manual    *ManualValidator
}

// NewCache creates a cache object with the provided fetcher and validator.
func NewCache[T any](fetcher Fetcher[T], validator Validator, cacheError bool) *Cache[T] {
	manual := NewManualValidator()
	return &Cache[T]{
		validator:  AnyOf(validator, manual),
		fetcher:    fetcher,
		manual:     manual,
		cacheError: cacheError,
	}
}

// Get will call the fetcher if the cache is invalid, otherwise
// it will return the last stored values.
func (c *Cache[T]) Get(ctx context.Context) (T, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.validator.ShouldFetch() {
		c.lastVal, c.lastErr = c.fetcher(ctx)
		c.validator.OnFetch()
	}

	if c.lastErr != nil && !c.cacheError {
		c.manual.Invalidate()
	}

	return c.lastVal, c.lastErr
}

// Invalidate will force the cache to call the fetcher on the next call to Get()
func (c *Cache[T]) Invalidate() {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.manual.Invalidate()
}

// CachedFetcher returns a Fetcher func with an embedded cache, if it
// is more convenient to deal with function types.
func CachedFetcher[T any](f Fetcher[T], validator Validator, cacheError bool) Fetcher[T] {
	c := NewCache(f, validator, cacheError)

	return c.Get
}
