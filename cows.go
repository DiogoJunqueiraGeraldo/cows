package cows

import (
	"sync"
	"sync/atomic"
)

// Cows - Copy On Write Slice
type Cows[T any] struct {
	// Atomic reference Counter
	arc *atomic.Int32

	// Read / Write Lock
	m *sync.RWMutex

	// Backing Array
	bArr []T
}

func NewCows[T any](size, cap int) *Cows[T] {
	arc := new(atomic.Int32)
	arc.Store(1)

	return &Cows[T]{
		arc:  arc,
		m:    new(sync.RWMutex),
		bArr: make([]T, size, cap),
	}
}

func (c *Cows[T]) copy() *Cows[T] {
	nArr := make([]T, len(c.bArr), cap(c.bArr))
	copy(nArr, c.bArr)

	arc := new(atomic.Int32)
	arc.Store(1)
	return &Cows[T]{
		arc:  arc,
		m:    new(sync.RWMutex),
		bArr: nArr,
	}
}

func (c *Cows[T]) Append(v T) *Cows[T] {
	c.m.Lock()
	defer c.m.Unlock()

	if c.arc.Load() > 1 {
		c.arc.Add(-1)
		nCow := c.copy()
		nCow.bArr = append(nCow.bArr, v)
		return nCow
	}

	c.bArr = append(c.bArr, v)
	return c
}

func (c *Cows[T]) Set(i int, v T) *Cows[T] {
	c.m.Lock()
	defer c.m.Unlock()

	if c.arc.Load() > 1 {
		nCow := c.copy()
		nCow.bArr[i] = v
		return nCow
	}

	c.bArr[i] = v
	return c
}

func (c *Cows[T]) Slice(start, end int) *Cows[T] {
	c.m.RLock()
	defer c.m.RUnlock()
	c.arc.Add(1)
	return &Cows[T]{
		arc:  c.arc,
		bArr: c.bArr[start:end],
		m:    c.m,
	}
}

func (c *Cows[T]) Release() {
	c.arc.Add(-1)
	c.bArr = c.bArr[:0]
}

func (c *Cows[T]) Get(i int) T {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.bArr[i]
}

func (c *Cows[T]) Len() int {
	return len(c.bArr)
}

func (c *Cows[T]) Cap() int {
	return cap(c.bArr)
}
