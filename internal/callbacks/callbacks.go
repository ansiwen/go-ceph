package callbacks

import (
	"sync"
)

// The logic of this file is largely adapted from:
// https://github.com/golang/go/wiki/cgo#function-variables
//
// Also helpful:
// https://eli.thegreenplace.net/2019/passing-callbacks-and-pointers-to-cgo/

// Callbacks provides a tracker for data that is to be passed between Go
// and C callback functions. The Go callback/object may not be passed
// by a pointer to C code and so instead integer indexes into an internal
// map are used.
// Typically the item being added will either be a callback function or
// a data structure containing a callback function. It is up to the caller
// to control and validate what "callbacks" get used.
type Callbacks struct {
	mutex sync.RWMutex
	cmap  map[uintptr]interface{}
	index uintptr
}

func (cb *Callbacks) nextIndex() uintptr {
	index := cb.index
	for {
		cb.index++
		if _, found := cb.cmap[cb.index]; !found {
			break
		}
	}
	return index
}

// New returns a new callbacks tracker.
func New() *Callbacks {
	return &Callbacks{cmap: make(map[uintptr]interface{})}
}

// Add a callback/object to the tracker and return a new index
// for the object.
func (cb *Callbacks) Add(v interface{}) uintptr {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	index := cb.nextIndex()
	cb.cmap[index] = v
	return index
}

// Remove a callback/object given it's index.
func (cb *Callbacks) Remove(index uintptr) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	delete(cb.cmap, index)
}

// Lookup returns a mapped callback/object given an index.
func (cb *Callbacks) Lookup(index uintptr) interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.cmap[index]
}
