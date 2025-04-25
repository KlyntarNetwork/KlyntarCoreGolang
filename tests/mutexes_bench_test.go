package tests

import (
	"sync"
	"sync/atomic"
	"testing"
)

type DataX struct {
	Value int
}

var (
	rwData    = &DataX{Value: 42}
	rwMu      sync.RWMutex
	atomicPtr atomic.Pointer[DataX]
)

func init() {
	atomicPtr.Store(&DataX{Value: 42})
}

// Benchmark: Read with RWMutex
func BenchmarkRWMutexRead(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rwMu.RLock()
			_ = rwData.Value
			rwMu.RUnlock()
		}
	})
}

// Benchmark: Read with atomic.Pointer
func BenchmarkAtomicPointerRead(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d := atomicPtr.Load()
			_ = d.Value
		}
	})
}
