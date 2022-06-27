package nmap

type AllowedKeysIf interface {
	uint8 | uint16 | uint32 | uint64 | uint
}

type element[T AllowedKeysIf, V any] struct {
	Key   T
	Value V
}

type NMap[T AllowedKeysIf, V any] struct {
	Elements []element[T, V]
}
