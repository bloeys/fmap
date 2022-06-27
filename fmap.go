package fmap

import "fmt"

const (
	//Must always be an even number
	elementsPerBucket         = 64
	elementsPerBucketBits     = 6
	elementsPerBucketBitsMask = 0b0011_1111

	maxConsecutiveGrows = 2
)

type AllowedKeysIf interface {
	uint8 | uint16 | uint32 | uint64 | uint
}

type element[T AllowedKeysIf, V any] struct {
	Key   T
	Value V
	IsSet bool
}

type Bucket[T AllowedKeysIf, V any] struct {
	Elements [elementsPerBucket]element[T, V]
}

type FMap[T AllowedKeysIf, V any] struct {
	Buckets     []Bucket[T, V]
	BucketCount uint64
	Len         uint
}

func (fm *FMap[T, V]) Set(key T, value V) {

	inBucketIndex := fm.GetElementIndexFromKey(key)
	e := &fm.Buckets[fm.GetBucketIndexFromKey(key)].Elements[inBucketIndex]

	if !e.IsSet {
		e.Key = key
		e.Value = value
		e.IsSet = true
		fm.Len++
		return
	} else if e.Key == key {
		e.Value = value
		return
	}

	// fmt.Println("Collision with key", key, "at load factor of", fm.LoadFactor(), "and len of", fm.Len)
	for attempts := 0; attempts < maxConsecutiveGrows; attempts++ {

		bucketIndex := fm.GetBucketIndexFromKey(key)
		b := &fm.Buckets[bucketIndex]
		for i := 0; i < elementsPerBucket; i++ {

			e := &b.Elements[i]
			if e.IsSet {
				if e.Key != key {
					continue
				}

				e.Value = value
				return
			}

			e.Key = key
			e.Value = value
			e.IsSet = true
			fm.Len++
			return
		}

		// println("Growing to", len(fm.Buckets)*2*elementsPerBucket, "with key", key, "; Len before grow:", fm.Len)
		fm.Grow()
	}

	panic("Grew map " + fmt.Sprint(maxConsecutiveGrows) + " times but still couldn't add key. Something is wrong. Key: " + fmt.Sprint(key))
}

func (fm *FMap[T, V]) Grow() {

	oldBuckets := fm.Buckets

	newBucketCount := uint64(len(fm.Buckets)) * 2

	fm.Len = 0 //Readding values to the new bucket will increase the size
	fm.BucketCount = newBucketCount
	fm.Buckets = make([]Bucket[T, V], newBucketCount)
	for i := 0; i < len(oldBuckets); i++ {

		b := &oldBuckets[i]
		for i := 0; i < elementsPerBucket; i++ {

			e := &b.Elements[i]
			if !e.IsSet {
				continue
			}

			fm.Set(e.Key, e.Value)
		}
	}

}

func (fm *FMap[T, V]) Get(key T) (value V) {

	i := fm.GetBucketIndexFromKey(key)
	b := &fm.Buckets[i]

	for i := 0; i < elementsPerBucket; i++ {

		e := &b.Elements[i]
		if !e.IsSet || e.Key != key {
			continue
		}

		return e.Value
	}

	return value
}

func (fm *FMap[T, V]) GetWithOK(key T) (value V, ok bool) {

	i := fm.GetBucketIndexFromKey(key)
	b := &fm.Buckets[i]

	for i := 0; i < elementsPerBucket; i++ {

		e := &b.Elements[i]
		if !e.IsSet || e.Key != key {
			continue
		}

		return e.Value, true
	}

	return value, false
}

func (fm *FMap[T, V]) Contains(key T) bool {

	i := fm.GetBucketIndexFromKey(key)
	b := &fm.Buckets[i]

	for i := 0; i < elementsPerBucket; i++ {

		e := &b.Elements[i]
		if !e.IsSet || e.Key != key {
			continue
		}

		return true
	}

	return false
}

func (fm *FMap[T, V]) Delete(key T) {

	i := fm.GetBucketIndexFromKey(key)
	b := &fm.Buckets[i]

	for i := 0; i < elementsPerBucket; i++ {

		e := &b.Elements[i]
		if e.Key != key {
			continue
		}

		e.IsSet = false
		return
	}
}

func (fm *FMap[T, V]) GetBucketIndexFromKey(key T) uint64 {

	//We can get the remainder without division by: number & (evenNumber - 1).
	//The lower n bits are not used for bucket selection because they are reserved for in-bucket indexing
	return uint64(key>>elementsPerBucketBits) & uint64(len(fm.Buckets)-1)
}

func (fm *FMap[T, V]) GetElementIndexFromKey(key T) uint8 {
	x := uint8(key) & elementsPerBucketBitsMask
	return x & (elementsPerBucket - 1)
}

func (fm *FMap[T, V]) LoadFactor() float32 {
	return float32(fm.Len) / float32(fm.Cap())
}

func (fm *FMap[T, V]) Cap() int {
	return len(fm.Buckets) * elementsPerBucket
}

func NewFMap[T AllowedKeysIf, V any]() *FMap[T, V] {

	//We need to ensure bucket count is always even so we can use & to do remainder
	fm := &FMap[T, V]{
		Buckets:     make([]Bucket[T, V], 2),
		BucketCount: 2,
	}

	return fm
}
