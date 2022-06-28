package fmap

import "fmt"

const (
	//Must always be an even number
	ElementsPerBucket           = 8
	elementsPerBucketBits       = 3
	elementsPerBucketBitsMask64 = 0x0000_0000_0000_0007 //0b0000...0000_0111

	maxConsecutiveGrows = 3
)

type AllowedKeysIf interface {
	uint8 | uint16 | uint32 | uint64 | uint
}

type FMap[T AllowedKeysIf, V any] struct {

	//Parallel arrays are used (instead of []struct{Key,Value,IsSet}) because they better use cache.
	//For example looping over IsSet, since cache will be full of IsSets and no useless data (i.e. Keys and Values)
	Keys             []T
	Values           []V
	ElementsInBucket []uint8

	cap uint64
	len uint64

	//A bucket is defined by a starting index in the flat array and a constant size.
	//For example, if ElementsPerBucket=8 then bucket=0 will be 0<=i<=7, bucket=1 will be 8<=i<=15 etc
	bucketCount uint64
}

func (fm *FMap[T, V]) Set(key T, value V) {

	bucketIndex := fm.GetBucketIndexFromKey(key)
	bucketElementCounter := &fm.ElementsInBucket[bucketIndex>>elementsPerBucketBits]

	//Ensure we have room in the selected bucket
	freedSpaceInBucket := false
	for attempts := 0; attempts < maxConsecutiveGrows; attempts++ {

		if *bucketElementCounter == ElementsPerBucket {
			fm.Grow()
			bucketIndex = fm.GetBucketIndexFromKey(key)
			bucketElementCounter = &fm.ElementsInBucket[bucketIndex>>elementsPerBucketBits]
			continue
		}

		freedSpaceInBucket = true
		break
	}

	if !freedSpaceInBucket {
		panic("Grew map " + fmt.Sprint(maxConsecutiveGrows) + " times but still couldn't add key. Something is wrong. Key: " + fmt.Sprint(key))
	}

	//Removes all bound checks until the loop.
	//For some reason using fm.XYZ directly doesn't remove all following bound checks
	ks := fm.Keys
	vs := fm.Values

	//Removes all bound checks inside the loop
	_ = ks[bucketIndex+ElementsPerBucket-1]
	_ = vs[bucketIndex+ElementsPerBucket-1]

	//Handle key overwriting
	uncheckedElements := *bucketElementCounter
	for i := bucketIndex; uncheckedElements > 0 && i <= bucketIndex+ElementsPerBucket-1; i++ {

		uncheckedElements--
		if ks[i] == key {
			vs[i] = value
			return
		}
	}

	//New key
	ks[bucketIndex+uint64(*bucketElementCounter)] = key
	vs[bucketIndex+uint64(*bucketElementCounter)] = value
	*bucketElementCounter++
	fm.len++
}

func (fm *FMap[T, V]) Grow() {

	oldKeys := fm.Keys
	oldValues := fm.Values

	oldBucketCount := fm.bucketCount
	oldElementsInBucket := fm.ElementsInBucket

	fm.len = 0 //Readding values to the new bucket will increase the size
	fm.cap *= 2
	fm.bucketCount *= 2

	fm.Keys = make([]T, fm.cap)
	fm.Values = make([]V, fm.cap)
	fm.ElementsInBucket = make([]uint8, fm.bucketCount)

	for i := uint64(0); i < oldBucketCount; i++ {

		bucketStartIndex := i * ElementsPerBucket
		bucketElementsCounter := uint64(oldElementsInBucket[i])
		for j := uint64(0); j < bucketElementsCounter; j++ {
			fm.Set(oldKeys[bucketStartIndex+j], oldValues[bucketStartIndex+j])
		}
	}
}

func (fm *FMap[T, V]) Get(key T) (value V) {
	value, _ = fm.GetWithOK(key)
	return value
}

func (fm *FMap[T, V]) GetWithOK(key T) (value V, ok bool) {

	bucketIndex := fm.GetBucketIndexFromKey(key)

	_ = fm.Keys[bucketIndex+ElementsPerBucket-1]
	_ = fm.Values[bucketIndex+ElementsPerBucket-1]

	bucketEleCount := fm.ElementsInBucket[bucketIndex>>elementsPerBucketBits]
	for i := bucketIndex; bucketEleCount > 0; i++ {

		bucketEleCount--
		if fm.Keys[i] != key {
			continue
		}

		return fm.Values[i], true
	}

	return value, false
}

func (fm *FMap[T, V]) Contains(key T) bool {

	bucketIndex := fm.GetBucketIndexFromKey(key)

	_ = fm.Keys[bucketIndex+ElementsPerBucket-1]
	_ = fm.Values[bucketIndex+ElementsPerBucket-1]

	bucketEleCount := fm.ElementsInBucket[bucketIndex>>elementsPerBucketBits]
	for i := bucketIndex; bucketEleCount > 0; i++ {

		bucketEleCount--
		if fm.Keys[i] == key {
			return true
		}
	}

	return false
}

func (fm *FMap[T, V]) Delete(key T) {

	bucketIndex := fm.GetBucketIndexFromKey(key)

	_ = fm.Keys[bucketIndex+ElementsPerBucket-1]
	_ = fm.Values[bucketIndex+ElementsPerBucket-1]

	elementsInBucket := &fm.ElementsInBucket[bucketIndex>>elementsPerBucketBits]
	elementsToCheck := *elementsInBucket
	for i := bucketIndex; elementsToCheck > 0; i++ {

		if fm.Keys[i] != key {
			elementsToCheck--
			continue
		}

		//Move the deleted key to the newly 'unreachable' bucket location.
		//Any location >ElementsInBucket is unreachable, and when the counter is
		//decremented a new location is created that is unreachable.
		//Put the deleted key/value there.
		unreachableIndex := bucketIndex + uint64(*elementsInBucket-1)
		fm.Keys[i], fm.Keys[unreachableIndex] = fm.Keys[unreachableIndex], fm.Keys[i]
		fm.Values[i], fm.Values[unreachableIndex] = fm.Values[unreachableIndex], fm.Values[i]
		*elementsInBucket--
		return
	}
}

func (fm *FMap[T, V]) GetBucketIndexFromKey(key T) uint64 {

	//We can get the remainder without division by: number & (evenNumber - 1).
	//The lower n bits are not used for bucket selection because they are reserved for in-bucket indexing
	return (uint64(key>>elementsPerBucketBits) & (fm.bucketCount - 1)) * ElementsPerBucket
}

func (fm *FMap[T, V]) GetElementIndexFromKey(key T) uint64 {
	x := uint64(key) & elementsPerBucketBitsMask64
	return x & (ElementsPerBucket - 1)
}

func (fm *FMap[T, V]) LoadFactor() float32 {
	return float32(fm.len) / float32(fm.Cap())
}

func (fm *FMap[T, V]) Cap() uint64 {
	return fm.cap
}

func NewFMap[T AllowedKeysIf, V any]() *FMap[T, V] {

	//We need to ensure bucket count is always even so we can use & to do remainder
	const bucketCount = 2

	fm := &FMap[T, V]{
		Keys:   make([]T, bucketCount*ElementsPerBucket),
		Values: make([]V, bucketCount*ElementsPerBucket),

		ElementsInBucket: make([]uint8, bucketCount),

		len:         0,
		cap:         bucketCount * ElementsPerBucket,
		bucketCount: bucketCount,
	}

	return fm
}
