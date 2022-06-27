package fmap

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

func (fm *FMap[T, V]) Add(key T, value V) {

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
		break
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
		Buckets: make([]Bucket[T, V], 10),
	}

	return fm
}
