package fmap

import "fmt"

const (
	maxElementsPerBucket = 8
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
	Elements [maxElementsPerBucket]element[T, V]
}

type FMap[T AllowedKeysIf, V any] struct {
	Buckets []Bucket[T, V]
}

func (fm *FMap[T, V]) Set(key T, value V) {

	for attempts := 0; attempts < 3; attempts++ {

		i := fm.GetBucketIndexFromKey(key)
		b := &fm.Buckets[i]
		for i := 0; i < maxElementsPerBucket; i++ {

			e := &b.Elements[i]
			if e.IsSet {
				continue
			}

			e.Key = key
			e.Value = value
			e.IsSet = true
			return
		}

		println("Growing to", len(fm.Buckets)*2, "with key", key)
		fm.Grow()
	}

	panic("Grew map 3 times but still couldn't add key. Something is wrong. Key: " + fmt.Sprint(key))
}

func (fm *FMap[T, V]) Grow() {

	oldBuckets := fm.Buckets
	fm.Buckets = make([]Bucket[T, V], len(fm.Buckets)*2)
	for i := 0; i < len(oldBuckets); i++ {

		b := &oldBuckets[i]
		for i := 0; i < maxElementsPerBucket; i++ {

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

	for i := 0; i < maxElementsPerBucket; i++ {

		e := &b.Elements[i]
		if e.Key != key {
			continue
		}

		return e.Value
	}

	return value
}

func (fm *FMap[T, V]) GetWithOK(key T) (value V, ok bool) {

	i := fm.GetBucketIndexFromKey(key)
	b := &fm.Buckets[i]

	for i := 0; i < maxElementsPerBucket; i++ {

		e := &b.Elements[i]
		if e.Key != key {
			continue
		}

		return e.Value, true
	}

	return value, false
}

func (fm *FMap[T, V]) Contains(key T) bool {

	i := fm.GetBucketIndexFromKey(key)
	b := &fm.Buckets[i]

	for i := 0; i < maxElementsPerBucket; i++ {

		e := &b.Elements[i]
		if e.Key != key {
			continue
		}

		return true
	}

	return false
}

func (fm *FMap[T, V]) GetBucketIndexFromKey(key T) uint64 {
	return uint64(key) % uint64(len(fm.Buckets))
}

func NewFMap[T AllowedKeysIf, V any]() *FMap[T, V] {

	fm := &FMap[T, V]{
		Buckets: make([]Bucket[T, V], 8),
	}

	return fm
}
