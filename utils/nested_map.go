package utils

/*
 * @Author: weisu.yxd
 * @Date: 2022-11-17 12:07:09
 */

import (
	"errors"
	"reflect"
	"sync"
)

// Define error code for NestedMap.
var (
	ErrNotBranch = errors.New("keys are not branch") // A node in the path specified by keys is not branch.
	ErrNotLeaf   = errors.New("keys are not leaf")   // A node in the path specified by keys is not leaf.
	ErrNotExist  = errors.New("value is not exist")  // No value is present
)

// NestedMap is is a map that can be nested infinitely, as shown below:
//
//	               /->key1:value1
//	/->branch_map1 -->key2:value2
//
// root_map                \->key3:value3
//
//	\->branch_map2 -->branch_map3 -->key4:value4
//	                              \->key5:value5
//
// NestedMap key type is ...interface{}, it can only be in the following formats:
// 1. "a", 2, "c" or,
// 2. [1, "b", 3]
// But it cannot be in the following format:
// 1. "a", ["b", "c"] or,
// 2. 1, [2, 3] or,
// 3. "a", [2, 3] or,
// 4. 1, ["b", "c"]
type NestedMap sync.Map

// Load returns the value stored in the map for a key, or nil if no value is present.
func (nm *NestedMap) Load(keys ...interface{}) (interface{}, error) {
	keys = nm.convertKeys(keys...)
	m, err := nm.search(keys[:len(keys)-1]...)
	if nil != err {
		return nil, err
	}

	value, exist := m.Load(keys[len(keys)-1])
	if !exist {
		return nil, ErrNotExist
	} else if _, ok := value.(*NestedMap); ok {
		return nil, ErrNotLeaf
	}

	return value, nil
}

// Store sets the value for keys.
func (nm *NestedMap) Store(value interface{}, keys ...interface{}) error {
	keys = nm.convertKeys(keys...)
	m, err := nm.searchOrCreate(keys[:len(keys)-1]...)
	if nil != err {
		return err
	}

	m.Store(keys[len(keys)-1], value)
	return nil
}

// LoadOrStore returns the existing value for the keys if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (nm *NestedMap) LoadOrStore(value interface{}, keys ...interface{}) (actual interface{}, loaded bool, err error) {
	keys = nm.convertKeys(keys...)
	m, err := nm.searchOrCreate(keys[:len(keys)-1]...)
	if nil != err {
		return nil, false, err
	}

	actual, loaded = m.LoadOrStore(keys[len(keys)-1], value)
	return actual, loaded, nil
}

// Delete deletes the value for keys.
func (nm *NestedMap) Delete(keys ...interface{}) error {
	keys = nm.convertKeys(keys...)
	m, err := nm.search(keys[:len(keys)-1]...)
	if nil != err {
		return err
	}

	m.Delete(keys[len(keys)-1])
	return nil
}

// Range calls f sequentially for each key and value present in the map. If f returns false, range stops the iteration.
func (nm *NestedMap) Range(f func(keys []interface{}, value interface{}) bool, keys ...interface{}) error {
	keys = nm.convertKeys(keys...)
	m, err := nm.search(keys...)
	if nil != err {
		return err
	}

	nm.nestRange(m, f, keys)
	return nil
}

// nestRange recursive range the map.
func (nm *NestedMap) nestRange(m *sync.Map, f func(keys []interface{}, value interface{}) bool, keys []interface{}) bool {
	// Extend keys.
	keys = append(keys, nil)
	depth := len(keys)

	// Iterate branch.
	m.Range(func(key, value interface{}) bool {
		// Set key.
		keys[depth-1] = key

		// Check branch.
		b, ok := value.(*NestedMap)
		if !ok {
			return f(keys, value)
		}

		// Recursive range.
		ok = nm.nestRange((*sync.Map)(b), f, keys)
		keys = keys[:depth]
		return ok
	})

	keys = keys[:depth-1]
	return true
}

// nest nested search path specified by keys.
func (nm *NestedMap) search(keys ...interface{}) (*sync.Map, error) {
	b, m := nm, (*sync.Map)(nm)
	for i := 0; i < len(keys); i++ {
		// Search path.
		value, exist := m.Load(keys[i])
		if !exist {
			return nil, ErrNotExist
		}

		// Check branch type.
		var ok bool
		if b, ok = value.(*NestedMap); !ok {
			return nil, ErrNotBranch
		}

		// Continue search.
		m = (*sync.Map)(b)
	}
	return m, nil
}

// searchOrCreate nested search path specified by keys, create if the branch does not exist.
func (nm *NestedMap) searchOrCreate(keys ...interface{}) (*sync.Map, error) {
	b, m := nm, (*sync.Map)(nm)
	for i := 0; i < len(keys); i++ {
		// Load before store to avoid useless memory applyment.
		value, exist := m.Load(keys[i])
		if !exist {
			value, _ = m.LoadOrStore(keys[i], new(NestedMap))
		}

		// Check branch type.
		var ok bool
		if b, ok = value.(*NestedMap); !ok {
			return nil, ErrNotBranch
		}

		// Continue search.
		m = (*sync.Map)(b)
	}
	return m, nil
}

// convertKeys unify keys' type []interface{}
func (*NestedMap) convertKeys(keys ...interface{}) []interface{} {
	// Convert when there is only one key and it is a slice type
	if len(keys) != 1 {
		return keys
	} else if t := reflect.TypeOf(keys[0]); t.Kind() != reflect.Slice {
		return keys
	} else if t.Elem().Kind() == reflect.Interface {
		return keys[0].([]interface{})
	}

	// Convert one by one.
	s := reflect.ValueOf(keys[0])
	r := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		r[i] = s.Index(i).Interface()
	}
	return r
}
