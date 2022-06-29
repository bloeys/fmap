package fmap_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/bloeys/fmap"
)

const (
	mapSize = 1000_000
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

	//Ensure ElementsInBucket works properly
	fm = fmap.NewFMap[uint, string]()
	fm.Set(0, "A")
	fm.Set(1, "B")
	fm.Delete(0)
	v, ok = fm.GetWithOK(1)
	AllTrue(t, v == "B", ok)

	for i := uint(0); i < 256; i++ {
		fm.Set(i, "There"+fmt.Sprint(i))
		// fmt.Printf("i=%d, grow count=%d\n", i, fm.GrowCount)
		// fmt.Printf("i=%d, bucket=%d\n", i, fm.GetBucketIndexFromKey(i)/fmap.ElementsPerBucket)
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

	fm := fmap.NewFMap[uint, string]()

	//Fill first bucket
	for i := 0; i < fmap.ElementsPerBucket; i++ {
		fm.Set(uint(i), "There"+fmt.Sprint(i))
	}

	//With first bucket filled, everytime x=Cap is inserted a Grow is forced.
	//This simulates a DOS attack
	for i := 0; i < 10; i++ {
		fm.Set(uint(fm.Cap()), "There"+fmt.Sprint(i))
	}

	fm = fmap.NewFMap[uint, string]()
	currCap := fm.Cap()
	lf := fm.LoadFactor()
	for i := uint(0); i < 1000_000; i++ {

		x := uint(rand.Uint32())
		fm.Set(x, "There"+fmt.Sprint(i))

		if fm.Cap() != currCap {

			currCap = fm.Cap()
			fmt.Printf("i=%d, grow count=%d, cap=%d, oldLF=%f\n", x, fm.GrowCount, fm.Cap(), lf)
			lf = fm.LoadFactor()
		}
		// fmt.Printf("i=%d, bucket=%d\n", i, fm.GetBucketIndexFromKey(i)/fmap.ElementsPerBucket)
	}

}

const seed = 1092381093

func BenchmarkFMapAdd(b *testing.B) {

	rand.Seed(seed)
	fm := fmap.NewFMap[uint, string]()

	for i := uint(0); i < uint(b.N); i++ {
		fm.Set(i, "Hi")
	}
}

func BenchmarkGoMapAdd(b *testing.B) {

	rand.Seed(seed)
	m := map[uint]string{}

	for i := uint(0); i < uint(b.N); i++ {
		m[i] = "Hi"
	}
}

func BenchmarkFMapAddRand(b *testing.B) {

	rand.Seed(seed)
	fm := fmap.NewFMap[uint, string]()

	for i := 0; i < b.N; i++ {
		x := uint(rand.Uint32())
		fm.Set(x, "Hi")
	}
}

func BenchmarkGoMapAddRand(b *testing.B) {

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

	for i := 0; i < mapSize; i++ {
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

	for i := 0; i < mapSize; i++ {
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

	for i := 0; i < mapSize; i++ {
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

	for i := 0; i < mapSize; i++ {
		x := uint(rand.Uint32())
		m[x] = "Hi"
	}
	b.StartTimer()

	for i := uint(0); i < uint(b.N); i++ {
		x := uint(rand.Uint32())
		mapOut = m[x]
	}
}

var mapContains = false

func BenchmarkFMapContains(b *testing.B) {

	b.StopTimer()
	rand.Seed(seed)
	fm := fmap.NewFMap[uint, string]()

	for i := 0; i < mapSize; i++ {
		x := uint(rand.Uint32())
		fm.Set(x, "Hi")
	}
	b.StartTimer()

	for i := uint(0); i < uint(b.N); i++ {
		mapContains = fm.Contains(i)
	}

}

func BenchmarkGoMapContains(b *testing.B) {

	b.StopTimer()
	rand.Seed(seed)
	m := map[uint]string{}

	for i := 0; i < mapSize; i++ {
		x := uint(rand.Uint32())
		m[x] = "Hi"
	}
	b.StartTimer()

	for i := uint(0); i < uint(b.N); i++ {
		_, mapContains = m[i]
	}
}

func BenchmarkFMapContainsRand(b *testing.B) {

	b.StopTimer()
	rand.Seed(seed)
	fm := fmap.NewFMap[uint, string]()

	for i := 0; i < mapSize; i++ {
		x := uint(rand.Uint32())
		fm.Set(x, "Hi")
	}
	b.StartTimer()

	for i := uint(0); i < uint(b.N); i++ {
		x := uint(rand.Uint32())
		mapContains = fm.Contains(x)
	}

}

func BenchmarkGoMapContainsRand(b *testing.B) {

	b.StopTimer()
	rand.Seed(seed)
	m := map[uint]string{}

	for i := 0; i < mapSize; i++ {
		x := uint(rand.Uint32())
		m[x] = "Hi"
	}
	b.StartTimer()

	for i := uint(0); i < uint(b.N); i++ {
		x := uint(rand.Uint32())
		_, mapContains = m[x]
	}
}

func BenchmarkFMapDelete(b *testing.B) {

	b.StopTimer()
	rand.Seed(seed)
	fm := fmap.NewFMap[uint, string]()

	for i := 0; i < mapSize; i++ {
		x := uint(rand.Uint32())
		fm.Set(x, "Hi")
	}
	b.StartTimer()

	for i := uint(0); i < uint(b.N); i++ {
		fm.Delete(i)
	}

}

func BenchmarkGoMapDelete(b *testing.B) {

	b.StopTimer()
	rand.Seed(seed)
	m := map[uint]string{}

	for i := 0; i < mapSize; i++ {
		x := uint(rand.Uint32())
		m[x] = "Hi"
	}
	b.StartTimer()

	for i := uint(0); i < uint(b.N); i++ {
		delete(m, i)
	}
}

func BenchmarkFMapDeleteRand(b *testing.B) {

	b.StopTimer()
	rand.Seed(seed)
	fm := fmap.NewFMap[uint, string]()

	for i := 0; i < mapSize; i++ {
		x := uint(rand.Uint32())
		fm.Set(x, "Hi")
	}
	b.StartTimer()

	for i := uint(0); i < uint(b.N); i++ {
		x := uint(rand.Uint32())
		fm.Delete(x)
	}

}

func BenchmarkGoMapDeleteRand(b *testing.B) {

	b.StopTimer()
	rand.Seed(seed)
	m := map[uint]string{}

	for i := 0; i < mapSize; i++ {
		x := uint(rand.Uint32())
		m[x] = "Hi"
	}
	b.StartTimer()

	for i := uint(0); i < uint(b.N); i++ {
		x := uint(rand.Uint32())
		delete(m, x)
	}
}
