package fmap

import "fmt"

const (
	//Must always be an even number
	elementsPerBucket           = 8
	elementsPerBucketBits       = 3
	elementsPerBucketBitsMask64 = 0x0000_0000_0000_0007

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

type FMap[T AllowedKeysIf, V any] struct {
	Elements []element[T, V]

	cap         uint64
	len         uint64
	bucketCount uint64
}

func (fm *FMap[T, V]) Set(key T, value V) {

	bucketIndex := fm.GetBucketIndexFromKey(key)
	inBucketIndex := fm.GetElementIndexFromKey(key)
	e := &fm.Elements[bucketIndex+inBucketIndex]

	if !e.IsSet {
		e.Key = key
		e.Value = value
		e.IsSet = true
		fm.len++
		return
	} else if e.Key == key {
		e.Value = value
		return
	}

	// fmt.Println("Collision with key", key, "at load factor of", fm.LoadFactor(), "and len of", fm.Len)
	for attempts := 0; attempts < maxConsecutiveGrows; attempts++ {

		for i := bucketIndex; i < bucketIndex+elementsPerBucket; i++ {

			e := &fm.Elements[i]
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
			fm.len++
			return
		}

		// println("Growing to", len(fm.Buckets)*2*elementsPerBucket, "with key", key, "; Len before grow:", fm.Len)
		fm.Grow()

		bucketIndex = fm.GetBucketIndexFromKey(key)
	}

	panic("Grew map " + fmt.Sprint(maxConsecutiveGrows) + " times but still couldn't add key. Something is wrong. Key: " + fmt.Sprint(key))
}

func (fm *FMap[T, V]) Grow() {

	oldElements := fm.Elements

	fm.len = 0 //Readding values to the new bucket will increase the size
	fm.cap = fm.cap * 2
	fm.bucketCount *= 2
	fm.Elements = make([]element[T, V], fm.cap)
	for i := 0; i < len(oldElements); i++ {

		e := &oldElements[i]
		if !e.IsSet {
			continue
		}

		fm.Set(e.Key, e.Value)
	}
}

func (fm *FMap[T, V]) Get(key T) (value V) {
	value, _ = fm.GetWithOK(key)
	return value
}

func (fm *FMap[T, V]) GetWithOK(key T) (value V, ok bool) {

	bucketIndex := fm.GetBucketIndexFromKey(key)

	for i := bucketIndex; i < bucketIndex+elementsPerBucket; i++ {

		e := &fm.Elements[i]
		if !e.IsSet || e.Key != key {
			continue
		}

		return e.Value, true
	}

	return value, false
}

func (fm *FMap[T, V]) Contains(key T) bool {

	bucketIndex := fm.GetBucketIndexFromKey(key)

	for i := bucketIndex; i < bucketIndex+elementsPerBucket; i++ {

		e := &fm.Elements[i]
		if !e.IsSet || e.Key != key {
			continue
		}

		return true
	}

	return false
}

func (fm *FMap[T, V]) Delete(key T) {

	bucketIndex := fm.GetBucketIndexFromKey(key)

	for i := bucketIndex; i < bucketIndex+elementsPerBucket; i++ {

		e := &fm.Elements[i]
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
	return (uint64(key>>elementsPerBucketBits) & (fm.bucketCount - 1)) * elementsPerBucket
}

func (fm *FMap[T, V]) GetElementIndexFromKey(key T) uint64 {
	x := uint64(key) & elementsPerBucketBitsMask64
	return x & (elementsPerBucket - 1)
}

func (fm *FMap[T, V]) LoadFactor() float32 {
	return float32(fm.len) / float32(fm.Cap())
}

func (fm *FMap[T, V]) Cap() uint64 {
	return fm.cap
}

func NewFMap[T AllowedKeysIf, V any]() *FMap[T, V] {

	//We need to ensure bucket count is always even so we can use & to do remainder
	fm := &FMap[T, V]{
		Elements:    make([]element[T, V], 2*elementsPerBucket),
		len:         0,
		cap:         2 * elementsPerBucket,
		bucketCount: 2,
	}

	return fm
}
