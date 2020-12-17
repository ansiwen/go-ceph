package cutil

/*
#include <stdlib.h>

extern void pong(void*);

static inline void storeAndWait(void** c_ptr, void* go_ptr, void* rendezvouz) {
	*c_ptr = go_ptr;
	pong(rendezvouz);  // 1: pointer is stored
	pong(rendezvouz);  // 2: release request
	*c_ptr = NULL;
	pong(rendezvouz);  // 3: pointer is set to NULL
}
*/
import "C"

import (
	"sync"
	"unsafe"
)

// rendezvouz is a little helper type to synchronize two threads. ping() sends a
// ping and waits for the pong, and pong() (used in the other thread) waits for
// the ping and replies with a pong. That way two threads are guaranteed to
// "meet" at the code lines, where respective ping() and pong() are used.

type rendezvouz struct {
	m1, m2 sync.Mutex
}

func (v *rendezvouz) init() {
	v.m1.Lock()
	v.m2.Lock()
}

func (v *rendezvouz) ping() {
	v.m1.Unlock()
	v.m2.Lock()
}

func (v *rendezvouz) pong() {
	v.m1.Lock()
	v.m2.Unlock()
}

// PtrGuard respresents a guarded Go pointer (pointing to memory allocated by Go
// runtime) stored in C memory (allocated by C)
type PtrGuard struct {
	rendezvouz
	released bool
}

// NewPtrGuard writes the goPtr (pointing to Go memory) into C memory at the
// position cPtr, and returns a PtrGuard object.
func NewPtrGuard(cPtr *unsafe.Pointer, goPtr unsafe.Pointer) *PtrGuard {
	var v PtrGuard
	v.rendezvouz.init()
	go C.storeAndWait(cPtr, goPtr, unsafe.Pointer(&v.rendezvouz))
	v.ping() // 1: pointer is stored
	return &v
}

// Release removes the guarded Go pointer from the C memory by overwriting it
// with NULL.
func (v *PtrGuard) Release() {
	if !v.released {
		v.released = true
		v.ping() // 2: release request
		v.ping() // 3: pointer is set to NULL
	}
}

//export pong
func pong(p unsafe.Pointer) {
	m := (*rendezvouz)(p)
	m.pong()
}

// for tests
func cMalloc(n uintptr) unsafe.Pointer {
	return C.malloc(C.size_t(n))
}

func cFree(p unsafe.Pointer) {
	C.free(p)
}
