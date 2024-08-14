// Package cache provides partial implementations of Guava Cache,
// including support for LRU, Segmented LRU and TinyLFU.
package cache

// Key is any value which is comparable.
// See http://golang.org/ref/spec#Comparison_operators for details.
type Key interface{}

// Value is any value.
type Value interface{}

// Cache is a key-value cache which entries are added and stayed in the
// cache until either are evicted or manually invalidated.
type Cache interface {
	// GetIfPresent returns value associated with Key or (nil, false)
	// if there is no cached value for Key.
	GetIfPresent(Key) (Value, bool)

	// Put associates value with Key. If a value is already associated
	// with Key, the old one will be replaced with Value.
	Put(Key, Value)

	// Invalidate discards cached value of the given Key.
	Invalidate(Key)

	// InvalidateAll discards all entries.
	InvalidateAll()

	// Stats copies cache statistics to given Stats pointer.
	Stats(*Stats)

	// Close implements io.Closer for cleaning up all resources.
	// Users must ensure the cache is not being used before closing or
	// after closed.
	Close() error
}

// Func is a generic callback for entry events in the cache.
type Func func(Key, Value)

// LoadingCache is a cache with values are loaded automatically and stored
// in the cache until either evicted or manually invalidated.
type LoadingCache interface {
	Cache

	// Get returns value associated with Key or call underlying LoaderFunc
	// to load value if it is not present.
	Get(Key) (Value, error)

	// Refresh loads new value for Key. If the Key already existed, the previous value
	// will continue to be returned by Get while the new value is loading.
	// If Key does not exist, this function will block until the value is loaded.
	Refresh(Key)
}

// LoaderFunc retrieves the value corresponding to given Key.
type LoaderFunc func(Key) (Value, error)

// Executor specifies how cache loader is run to refresh value for the Key.
// By default, it is run in a new go routine.
type Executor interface {
	// Execute runs the fn asynchronously.
	Execute(fn func())
	// Close shuts down all running tasks. Currently, the error returned is not being used.
	Close() error
}
