package cows_test

import (
	"github.com/DiogoJunqueiraGeraldo/cows"
	"sync"
	"testing"
)

func TestNewCows(t *testing.T) {
	c := cows.NewCows[int](2, 5)

	if c.Len() != 2 {
		t.Errorf("expected len 2, got %d", c.Len())
	}
	if c.Cap() != 5 {
		t.Errorf("expected cap 5, got %d", c.Cap())
	}
}

func TestAppendUnique(t *testing.T) {
	c := cows.NewCows[int](0, 0)
	c = c.Append(10)
	if c.Len() != 1 {
		t.Errorf("expected len 1, got %d", c.Len())
	}
	if got := c.Get(0); got != 10 {
		t.Errorf("expected 10, got %d", got)
	}
}

func TestSetUnique(t *testing.T) {
	c := cows.NewCows[int](1, 1)
	c = c.Set(0, 42)
	if got := c.Get(0); got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
}

func TestAppendForksOnShared(t *testing.T) {
	c := cows.NewCows[int](0, 0)
	child := c.Slice(0, 0) // increment refcount

	// Append should create a new fork
	newC := c.Append(99)
	if newC == c {
		t.Errorf("expected a new cow instance after Append on shared")
	}

	if newC.Len() != 1 {
		t.Errorf("expected len 1, got %d", newC.Len())
	}

	// original child slice still valid
	if child.Len() != 0 {
		t.Errorf("child slice modified unexpectedly")
	}
}

func TestSetForksOnShared(t *testing.T) {
	c := cows.NewCows[int](1, 1)
	child := c.Slice(0, 1)
	newC := c.Set(0, 100)

	if newC == c {
		t.Errorf("expected a new cow instance after Set on shared")
	}

	if child.Get(0) != 0 {
		t.Errorf("child slice modified unexpectedly")
	}
	if newC.Get(0) != 100 {
		t.Errorf("newC did not set value correctly")
	}
}

func TestSliceAndRelease(t *testing.T) {
	c := cows.NewCows[int](3, 3)
	for i := 0; i < 3; i++ {
		c = c.Set(i, i+1)
	}

	child := c.Slice(0, 2)
	if child.Len() != 2 {
		t.Errorf("expected child len 2, got %d", child.Len())
	}

	c.Release()
	// child slice still intact
	if child.Get(0) != 1 || child.Get(1) != 2 {
		t.Errorf("child slice modified after parent release")
	}
}

func TestConcurrentAppend(t *testing.T) {
	c := cows.NewCows[int](0, 0)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Append(i)
		}(i)
	}
	wg.Wait()
	// Just ensure no panic, len may be nondeterministic if sharing was involved
}

func TestLenCap(t *testing.T) {
	c := cows.NewCows[int](2, 5)
	if l := c.Len(); l != 2 {
		t.Errorf("expected len 2, got %d", l)
	}
	if cp := c.Cap(); cp != 5 {
		t.Errorf("expected cap 5, got %d", cp)
	}
}

func TestGetOutOfBounds(t *testing.T) {
	c := cows.NewCows[int](1, 1)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on Get out of bounds")
		}
	}()
	_ = c.Get(2)
}

func TestSetOutOfBounds(t *testing.T) {
	c := cows.NewCows[int](1, 1)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on Set out of bounds")
		}
	}()
	c = c.Set(2, 10)
}
