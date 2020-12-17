package cutil

/*
extern void lock(void*);
extern void unlock(void*);

static inline void storeAndWait(void** c_ptr, void* go_ptr, void* stored_mtx, void* release_mtx) {
	*c_ptr = go_ptr;
	unlock(stored_mtx);  // send stored signal
	lock(release_mtx);   // wait for release signal
	*c_ptr = NULL;
	unlock(release_mtx); // unlock in case of of accidental double release
	unlock(stored_mtx);  // send stored signal
}
*/
import "C"

import (
	"sync"
	"unsafe"
)

// PtrGuard respresents a guarded Go pointer (pointing to memory allocated by Go
// runtime) stored in C memory (allocated by C)
type PtrGuard struct {
	stored  sync.Mutex
	release sync.Mutex
}

// WARNING: using mutexes for signalling like this is quite a delicate task in
// order to avoid deadlocks or panics. Whenever changing the code logic, please
// review at least three times that there is no unexpected state possible.
// Usually the natural choice would be to use channels instead, but these can
// not easily passed to C code because of the pointer-to-pointer cgo rule, and
// would require the use of a Go object registry.

// NewPtrGuard writes the goPtr (pointing to Go memory) into C memory at the
// position cPtr, and returns a PtrGuard object.
func NewPtrGuard(cPtr *unsafe.Pointer, goPtr unsafe.Pointer) *PtrGuard {
	var v PtrGuard
	v.release.Lock()
	v.stored.Lock()
	go C.storeAndWait(cPtr, goPtr, unsafe.Pointer(&v.stored), unsafe.Pointer(&v.release))
	v.stored.Lock() // wait for stored signal
	return &v
}

// Release removes the guarded Go pointer from the C memory by overwriting it
// with NULL.
func (v *PtrGuard) Release() {
	v.release.Unlock() // send release signal
	v.stored.Lock()    // wait for stored signal
	v.stored.Unlock()  // unlock in case of accidental double release
	v.release.Lock()   // lock in case of accidental double release
}

//export lock
func lock(p unsafe.Pointer) {
	m := (*sync.Mutex)(p)
	m.Lock()
}

//export unlock
func unlock(p unsafe.Pointer) {
	m := (*sync.Mutex)(p)
	m.Unlock()
}
