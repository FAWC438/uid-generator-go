package util

import "sync/atomic"

type PaddedAtomicLong struct {
	Value *atomic.Int64
	_     [7]uint64 // cache line padding
}

// TODO: Should GC be considered?

func NewPaddedAtomicLong(value int64) *PaddedAtomicLong {
	var valueToAdd atomic.Int64
	valueToAdd.Store(value)
	return &PaddedAtomicLong{Value: &valueToAdd}
}
