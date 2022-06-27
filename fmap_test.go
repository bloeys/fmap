package fmap_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/bloeys/fmap"
)

func TestFMap(t *testing.T) {

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

	fm.Delete(10)
	v, ok = fm.GetWithOK(10)
	AllTrue(t, !fm.Contains(10), fm.Get(10) == "", v == "", !ok)

	for i := uint(0); i < 256; i++ {
		fm.Set(i, "There"+fmt.Sprint(i))
	}

	for i := uint(0); i < 256; i++ {
		if fm.Get(i) != "There"+fmt.Sprint(i) {
			t.Errorf("Expected %d to exist in map\n", i)
		}
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

func TestPlayground(t *testing.T) {

}

const seed = 1092381093

func BenchmarkFMapAdd(b *testing.B) {

	rand.Seed(seed)
	fm := fmap.NewFMap[uint, string]()

	for i := 0; i < b.N; i++ {
		x := uint(rand.Uint32())
		fm.Set(x, "Hi")
	}
}

func BenchmarkGoMapAdd(b *testing.B) {

	rand.Seed(seed)
	m := map[uint]string{}

	for i := 0; i < b.N; i++ {
		x := uint(rand.Uint32())
		m[x] = "Hi"
	}
}

var mapOut string = ""

func BenchmarkFMapGet(b *testing.B) {

	b.StopTimer()
	rand.Seed(seed)
	fm := fmap.NewFMap[uint, string]()

	for i := 0; i < 1000_000; i++ {
		x := uint(rand.Uint32())
		fm.Set(x, "Hi")
	}
	b.StartTimer()

	for i := uint(0); i < uint(b.N); i++ {
		mapOut = fm.Get(i)
	}

}

func BenchmarkGoMapGet(b *testing.B) {

	b.StopTimer()
	rand.Seed(seed)
	m := map[uint]string{}

	for i := 0; i < 1000_000; i++ {
		x := uint(rand.Uint32())
		m[x] = "Hi"
	}
	b.StartTimer()

	for i := uint(0); i < uint(b.N); i++ {
		mapOut = m[i]
	}
}

func BenchmarkFMapGetRand(b *testing.B) {

	b.StopTimer()
	rand.Seed(seed)
	fm := fmap.NewFMap[uint, string]()

	for i := 0; i < 1000_000; i++ {
		x := uint(rand.Uint32())
		fm.Set(x, "Hi")
	}
	b.StartTimer()

	for i := uint(0); i < uint(b.N); i++ {
		x := uint(rand.Uint32())
		mapOut = fm.Get(x)
	}

}

func BenchmarkGoMapGetRand(b *testing.B) {

	b.StopTimer()
	rand.Seed(seed)
	m := map[uint]string{}

	for i := 0; i < 1000_000; i++ {
		x := uint(rand.Uint32())
		m[x] = "Hi"
	}
	b.StartTimer()

	for i := uint(0); i < uint(b.N); i++ {
		x := uint(rand.Uint32())
		mapOut = m[x]
	}
}
