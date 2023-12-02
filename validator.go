package cache

import "time"

func removeNil(all ...Validator) []Validator {
	ret := []Validator{}
	for _, v := range all {
		if v != nil {
			ret = append(ret, v)
		}
	}

	return ret
}

type anyValidator []Validator

func (a anyValidator) ShouldFetch() bool {
	for _, v := range a {
		if v.ShouldFetch() {
			return true
		}
	}

	return false
}

func (a anyValidator) OnFetch() {
	for _, v := range a {
		v.OnFetch()
	}
}

// AnyOf returns a validator that will refresh the cache if any of the provided validators
// think the cache needs to be refreshed.
func AnyOf(v ...Validator) Validator {
	return anyValidator(removeNil(v...))
}

type allValidator []Validator

func (a allValidator) ShouldFetch() bool {
	for _, v := range a {
		if !v.ShouldFetch() {
			return false
		}
	}

	return true
}

func (a allValidator) OnFetch() {
	for _, v := range a {
		v.OnFetch()
	}
}

// AllOf returns a validator that will refresh the cache if all of the provided validators
// think the cache needs to be refreshed.
func AllOf(v ...Validator) Validator {
	return allValidator(removeNil(v...))
}

// NoCacheValidator returns a validator that always attempts to refresh the cache.
// Using this functionally removes the caching feature.
type NoCacheValidator struct{}

// ShouldFetch satisfies the Validator interface
func (n NoCacheValidator) ShouldFetch() bool { return true }

// OnFetch satisfies the Validator interface
func (n NoCacheValidator) OnFetch() {}

// ManualValidator is a validator that only refreshes the cache when explicitly
// told to do so via the Invalidate method.
type ManualValidator struct {
	shouldFetch bool
}

// NewManualValidator creates a new ManualValidator that will return true on the first
// call to ShouldFetch()
func NewManualValidator() *ManualValidator {
	return &ManualValidator{
		shouldFetch: true,
	}
}

// ShouldFetch satisfies the Validator interface
func (m *ManualValidator) ShouldFetch() bool {
	return m.shouldFetch
}

// OnFetch satisfies the Validator interface
func (m *ManualValidator) OnFetch() {
	m.shouldFetch = false
}

// Invalidate will force the next call to ShouldFetch to return true
func (m *ManualValidator) Invalidate() {
	m.shouldFetch = true
}

// TimedCacheValidator is a validator that will refresh the cache if a provided
// amount of time has passed since the last call to Get()
type TimedCacheValidator struct {
	lastFetch     time.Time
	cacheDuration time.Duration
}

// NewTimedCacheValidator returns a TimedCacheValidator with the provided
// cache duration.
func NewTimedCacheValidator(dur time.Duration) *TimedCacheValidator {
	return &TimedCacheValidator{
		cacheDuration: dur,
	}
}

// ShouldFetch satisfies the Validator interface
func (tc *TimedCacheValidator) ShouldFetch() bool {
	return time.Since(tc.lastFetch) > tc.cacheDuration
}

// OnFetch satisfies the Validator interface
func (tc *TimedCacheValidator) OnFetch() {
	tc.lastFetch = time.Now()
}

// CallCountCacheValidator is a validator that will refresh the cache if
// ShouldFetch has been called at least
type CallCountCacheValidator struct {
	callCt int
}
