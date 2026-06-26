// Package reset_test предоставляет тестовые структуры для проверки кодогенератора.
package reset_test

import "time"

// generate:reset
type FullReset struct {
	IntVal     int
	Int8Val    int8
	Int16Val   int16
	Int32Val   int32
	Int64Val   int64
	UintVal    uint
	Uint8Val   uint8
	Uint16Val  uint16
	Uint32Val  uint32
	Uint64Val  uint64
	Float32Val float32
	Float64Val float64
	StringVal  string
	BoolVal    bool
	ByteVal    byte
	RuneVal    rune
	TimeVal    time.Time
}

// generate:reset
type PointersReset struct {
	IntPtr    *int
	StringPtr *string
	BoolPtr   *bool
	FloatPtr  *float64
}

// generate:reset
type SlicesMapReset struct {
	IntSlice []int
	StrSlice []string
	StrMap   map[string]string
	IntMap   map[int]float64
}

// generate:reset
type ComplexTypesReset struct {
	PIntSlice *[]int
	PStrSlice *[]string
	PStrMap   *map[string]string
	PIntMap   *map[int]float64
}

// generate:reset
type NestedReset struct {
	Simple FullReset
	Child  *FullReset
}

// Тоже помечена generate:reset для проверки встраивания
// generate:reset
type SomeStruct struct {
	X int
	Y string
}

// generate:reset
type WithEmbedded struct {
	SomeStruct
	*FullReset
	StrMap map[string]string
	Ints   []int
}

// Bar — generic-структура для тестирования embedded generic полей.
// Не помечена generate:reset.
type Bar[T any] struct {
	Value T
}

// generate:reset
type WithGenericEmbedded struct {
	Bar[int]
	Name string
}

// InnerWithGeneric — структура без Reset(), содержащая embedded generic поле.
// Не помечена generate:reset, встраивается в OuterWithInnerGeneric.
type InnerWithGeneric struct {
	Bar[int]
	X int
}

// generate:reset
type OuterWithInnerGeneric struct {
	Inner InnerWithGeneric
	Y     string
}
