package cutil

/*
#include <stdlib.h>
#include <sys/uio.h>
*/
import "C"

import (
	"unsafe"
)

const iovecSize = C.sizeof_struct_iovec

// StructIovecPtr is an unsafe pointer wrapping C's `*struct iovec`.
type StructIovecPtr unsafe.Pointer

// Iovec helps manage struct iovec arrays needed by some C functions.
type Iovec struct {
	// cvec represents an array of struct iovec C memory
	cvec unsafe.Pointer
	// length of the array (in elements)
	length    int
	ptrGuards []*PtrGuard
}

// NewIovec creates an Iovec, and underlying C memory, of the specified size.
func NewIovec(l int) *Iovec {
	r := &Iovec{
		cvec:   C.malloc(C.size_t(l) * C.size_t(iovecSize)),
		length: l,
	}
	return r
}

// ByteSlicesToIovec takes a slice of byte slices and returns a new iovec that
// maps the slice data to struct iovec entries.
func ByteSlicesToIovec(data [][]byte) *Iovec {
	iov := NewIovec(len(data))
	for i := range data {
		iov.Set(i, data[i])
	}
	return iov
}

// Pointer returns a StructIovecPtr that represents the C memory of the
// underlying array.
func (v *Iovec) Pointer() StructIovecPtr {
	return StructIovecPtr(unsafe.Pointer(v.cvec))
}

// Len returns the number of entries in the Iovec.
func (v *Iovec) Len() int {
	return v.length
}

// Free the C memory in the Iovec.
func (v *Iovec) Free() {
	if v.cvec != nil {
		for _, pg := range v.ptrGuards {
			pg.Release()
		}
		C.free(v.cvec)
		v.cvec = nil
		v.length = 0
	}
}

// Set will map the memory of the given byte slice to the iovec at the
// specified position.
func (v *Iovec) Set(i int, buf []byte) {
	offset := uintptr(i) * iovecSize
	iov := (*C.struct_iovec)(unsafe.Pointer(
		uintptr(unsafe.Pointer(v.cvec)) + offset))
	pg := NewPtrGuard(&iov.iov_base, unsafe.Pointer(&buf[0]))
	v.ptrGuards = append(v.ptrGuards, pg)
	iov.iov_len = C.size_t(len(buf))
}
