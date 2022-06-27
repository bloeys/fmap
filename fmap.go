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

type FMap[T AllowedKeysIf, V any] struct {

	//This setup (instead of []struct{Key,Value,IsSet}) makes better use of cache when
	//for example trying to find a set/unset element, since cache is full of one type of array
	//not useless data
	Keys   []T
	Values []V
	IsSet  []bool

	cap         uint64
	len         uint64
	bucketCount uint64
}

func (fm *FMap[T, V]) Set(key T, value V) {

	bucketIndex := fm.GetBucketIndexFromKey(key)
	elementIndex := bucketIndex + fm.GetElementIndexFromKey(key)

	//Removes all bound checks until the loop.
	//For some reason using fm.XYZ directly doesn't do it properly
	ks := fm.Keys
	vs := fm.Values
	is := fm.IsSet
	_ = ks[elementIndex]
	_ = vs[elementIndex]
	_ = is[elementIndex]

	if !is[elementIndex] {

		ks[elementIndex] = key
		vs[elementIndex] = value
		is[elementIndex] = true
		fm.len++
		return
	}

	if ks[elementIndex] == key {
		vs[elementIndex] = value
		return
	}
	// fmt.Println("Collision with key", key, "at load factor of", fm.LoadFactor(), "and len of", fm.Len)

	for attempts := 0; attempts < maxConsecutiveGrows; attempts++ {

		//Removes all bound checks inside the loop
		_ = ks[bucketIndex+elementsPerBucket-1]
		_ = vs[bucketIndex+elementsPerBucket-1]
		_ = is[bucketIndex+elementsPerBucket-1]

		for i := bucketIndex; i <= bucketIndex+elementsPerBucket-1; i++ {

			if is[i] {
				if ks[i] != key {
					continue
				}

				vs[i] = value
				return
			}

			ks[i] = key
			vs[i] = value
			is[i] = true
			fm.len++
			return
		}

		// println("Growing to", len(fm.Buckets)*2*elementsPerBucket, "with key", key, "; Len before grow:", fm.Len)
		fm.Grow()

		bucketIndex = fm.GetBucketIndexFromKey(key)

		ks = fm.Keys
		vs = fm.Values
		is = fm.IsSet
	}

	panic("Grew map " + fmt.Sprint(maxConsecutiveGrows) + " times but still couldn't add key. Something is wrong. Key: " + fmt.Sprint(key))
}

func (fm *FMap[T, V]) Grow() {

	oldKeys := fm.Keys
	oldValues := fm.Values
	oldIsSet := fm.IsSet

	fm.len = 0 //Readding values to the new bucket will increase the size
	fm.cap *= 2
	fm.bucketCount *= 2

	fm.Keys = make([]T, fm.cap)
	fm.Values = make([]V, fm.cap)
	fm.IsSet = make([]bool, fm.cap)

	_ = oldKeys[len(oldIsSet)-1]
	_ = oldValues[len(oldIsSet)-1]

	for i := 0; i < len(oldIsSet); i++ {

		if !oldIsSet[i] {
			continue
		}

		fm.Set(oldKeys[i], oldValues[i])
	}
}

func (fm *FMap[T, V]) Get(key T) (value V) {
	value, _ = fm.GetWithOK(key)
	return value
}

func (fm *FMap[T, V]) GetWithOK(key T) (value V, ok bool) {

	bucketIndex := fm.GetBucketIndexFromKey(key)

	for i := bucketIndex; i < bucketIndex+elementsPerBucket; i++ {

		if !fm.IsSet[i] || fm.Keys[i] != key {
			continue
		}

		return fm.Values[i], true
	}

	return value, false
}

func (fm *FMap[T, V]) Contains(key T) bool {

	bucketIndex := fm.GetBucketIndexFromKey(key)

	for i := bucketIndex; i < bucketIndex+elementsPerBucket; i++ {

		if !fm.IsSet[i] || fm.Keys[i] != key {
			continue
		}

		return true
	}

	return false
}

func (fm *FMap[T, V]) Delete(key T) {

	bucketIndex := fm.GetBucketIndexFromKey(key)

	for i := bucketIndex; i < bucketIndex+elementsPerBucket; i++ {

		if !fm.IsSet[i] || fm.Keys[i] != key {
			continue
		}

		fm.IsSet[i] = false
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
		Keys:   make([]T, 2*elementsPerBucket),
		Values: make([]V, 2*elementsPerBucket),
		IsSet:  make([]bool, 2*elementsPerBucket),

		len:         0,
		cap:         2 * elementsPerBucket,
		bucketCount: 2,
	}

	return fm
}
