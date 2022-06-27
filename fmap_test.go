package fmap_test

import (
	"testing"

	"github.com/bloeys/fmap"
)

func TestNMap(t *testing.T) {

	fm := fmap.NewFMap[uint, string]()

	fm.Set(1, "Hi")
	fm.Set(4, "Hello")
	fm.Set(10, "There")

	AllTrue(t, fm.Get(1) == "Hi", fm.Get(4) == "Hello", fm.Get(10) == "There")

	v, ok := fm.GetWithOK(55)
	AllTrue(t, v == "", !ok)
	v, ok = fm.GetWithOK(10)
	AllTrue(t, v == "There", ok)

	AllTrue(t, fm.Contains(1), fm.Contains(4), fm.Contains(10), !fm.Contains(5000))

	for i := 0; i < 100; i++ {
		fm.Set(8*uint(i), "There")
	}
}

func AllTrue(t *testing.T, vals ...bool) {

	for _, v := range vals {
		if !v {
			t.Errorf("Expected true but got false\n")
			return
		}
	}
}
