package reset_test

import (
	"testing"
	"time"
)

// TestFullReset проверяет, что все примитивные поля сбрасываются к нулевым значениям.
func TestFullReset(t *testing.T) {
	now := time.Now()
	s := &FullReset{
		IntVal:     42,
		Int8Val:    8,
		Int16Val:   16,
		Int32Val:   32,
		Int64Val:   64,
		UintVal:    1,
		Uint8Val:   2,
		Uint16Val:  3,
		Uint32Val:  4,
		Uint64Val:  5,
		Float32Val: 3.14,
		Float64Val: 2.718,
		StringVal:  "hello",
		BoolVal:    true,
		ByteVal:    'a',
		RuneVal:    '世',
		TimeVal:    now,
	}

	s.Reset()

	if s.IntVal != 0 {
		t.Errorf("IntVal = %d, want 0", s.IntVal)
	}
	if s.Int8Val != 0 {
		t.Errorf("Int8Val = %d, want 0", s.Int8Val)
	}
	if s.Int16Val != 0 {
		t.Errorf("Int16Val = %d, want 0", s.Int16Val)
	}
	if s.Int32Val != 0 {
		t.Errorf("Int32Val = %d, want 0", s.Int32Val)
	}
	if s.Int64Val != 0 {
		t.Errorf("Int64Val = %d, want 0", s.Int64Val)
	}
	if s.UintVal != 0 {
		t.Errorf("UintVal = %d, want 0", s.UintVal)
	}
	if s.Uint8Val != 0 {
		t.Errorf("Uint8Val = %d, want 0", s.Uint8Val)
	}
	if s.Uint16Val != 0 {
		t.Errorf("Uint16Val = %d, want 0", s.Uint16Val)
	}
	if s.Uint32Val != 0 {
		t.Errorf("Uint32Val = %d, want 0", s.Uint32Val)
	}
	if s.Uint64Val != 0 {
		t.Errorf("Uint64Val = %d, want 0", s.Uint64Val)
	}
	if s.Float32Val != 0 {
		t.Errorf("Float32Val = %f, want 0", s.Float32Val)
	}
	if s.Float64Val != 0 {
		t.Errorf("Float64Val = %f, want 0", s.Float64Val)
	}
	if s.StringVal != "" {
		t.Errorf("StringVal = %q, want empty", s.StringVal)
	}
	if s.BoolVal != false {
		t.Errorf("BoolVal = %v, want false", s.BoolVal)
	}
	if s.ByteVal != 0 {
		t.Errorf("ByteVal = %d, want 0", s.ByteVal)
	}
	if s.RuneVal != 0 {
		t.Errorf("RuneVal = %d, want 0", s.RuneVal)
	}
	if !s.TimeVal.IsZero() {
		t.Errorf("TimeVal = %v, want zero time", s.TimeVal)
	}
}

// TestPointersReset проверяет, что не-nil указатели разыменовываются и сбрасываются.
func TestPointersReset(t *testing.T) {
	intVal := 42
	strVal := "hello"
	boolVal := true
	floatVal := 3.14

	s := &PointersReset{
		IntPtr:    &intVal,
		StringPtr: &strVal,
		BoolPtr:   &boolVal,
		FloatPtr:  &floatVal,
	}

	s.Reset()

	if s.IntPtr == nil {
		t.Fatal("IntPtr is nil after Reset, expected non-nil pointer with zero value")
	}
	if *s.IntPtr != 0 {
		t.Errorf("*IntPtr = %d, want 0", *s.IntPtr)
	}
	if s.StringPtr == nil {
		t.Fatal("StringPtr is nil after Reset, expected non-nil pointer with zero value")
	}
	if *s.StringPtr != "" {
		t.Errorf("*StringPtr = %q, want empty", *s.StringPtr)
	}
	if s.BoolPtr == nil {
		t.Fatal("BoolPtr is nil after Reset, expected non-nil pointer with zero value")
	}
	if *s.BoolPtr != false {
		t.Errorf("*BoolPtr = %v, want false", *s.BoolPtr)
	}
	if s.FloatPtr == nil {
		t.Fatal("FloatPtr is nil after Reset, expected non-nil pointer with zero value")
	}
	if *s.FloatPtr != 0 {
		t.Errorf("*FloatPtr = %f, want 0", *s.FloatPtr)
	}
}

// TestPointersResetNil проверяет, что nil указатели остаются nil.
func TestPointersResetNil(t *testing.T) {
	s := &PointersReset{}

	s.Reset()

	if s.IntPtr != nil {
		t.Error("IntPtr should remain nil")
	}
	if s.StringPtr != nil {
		t.Error("StringPtr should remain nil")
	}
	if s.BoolPtr != nil {
		t.Error("BoolPtr should remain nil")
	}
	if s.FloatPtr != nil {
		t.Error("FloatPtr should remain nil")
	}
}

// TestSlicesMapReset проверяет, что срезы обрезаются, мапы очищаются.
func TestSlicesMapReset(t *testing.T) {
	s := &SlicesMapReset{
		IntSlice: []int{1, 2, 3},
		StrSlice: []string{"a", "b", "c"},
		StrMap:   map[string]string{"key": "value"},
		IntMap:   map[int]float64{1: 1.1},
	}

	s.Reset()

	if len(s.IntSlice) != 0 {
		t.Errorf("IntSlice length = %d, want 0", len(s.IntSlice))
	}
	if cap(s.IntSlice) == 0 {
		t.Error("IntSlice capacity should not be zero after reset")
	}
	if len(s.StrSlice) != 0 {
		t.Errorf("StrSlice length = %d, want 0", len(s.StrSlice))
	}
	if cap(s.StrSlice) == 0 {
		t.Error("StrSlice capacity should not be zero after reset")
	}
	if len(s.StrMap) != 0 {
		t.Errorf("StrMap length = %d, want 0", len(s.StrMap))
	}
	if len(s.IntMap) != 0 {
		t.Errorf("IntMap length = %d, want 0", len(s.IntMap))
	}
}

// TestComplexTypesReset проверяет указатели на срезы и мапы.
func TestComplexTypesReset(t *testing.T) {
	intSlice := []int{1, 2, 3}
	strSlice := []string{"a", "b", "c"}
	strMap := map[string]string{"key": "value"}
	intMap := map[int]float64{1: 1.1}

	s := &ComplexTypesReset{
		PIntSlice: &intSlice,
		PStrSlice: &strSlice,
		PStrMap:   &strMap,
		PIntMap:   &intMap,
	}

	s.Reset()

	if s.PIntSlice == nil {
		t.Fatal("PIntSlice is nil, expected non-nil pointer")
	}
	if len(*s.PIntSlice) != 0 {
		t.Errorf("*PIntSlice length = %d, want 0", len(*s.PIntSlice))
	}
	if s.PStrSlice == nil {
		t.Fatal("PStrSlice is nil, expected non-nil pointer")
	}
	if len(*s.PStrSlice) != 0 {
		t.Errorf("*PStrSlice length = %d, want 0", len(*s.PStrSlice))
	}
	if s.PStrMap == nil {
		t.Fatal("PStrMap is nil, expected non-nil pointer")
	}
	if len(*s.PStrMap) != 0 {
		t.Errorf("*PStrMap length = %d, want 0", len(*s.PStrMap))
	}
	if s.PIntMap == nil {
		t.Fatal("PIntMap is nil, expected non-nil pointer")
	}
	if len(*s.PIntMap) != 0 {
		t.Errorf("*PIntMap length = %d, want 0", len(*s.PIntMap))
	}
}

// TestComplexTypesResetNil проверяет, что nil указатели на срезы/мапы остаются nil.
func TestComplexTypesResetNil(t *testing.T) {
	s := &ComplexTypesReset{}

	s.Reset()

	if s.PIntSlice != nil {
		t.Error("PIntSlice should remain nil")
	}
	if s.PStrSlice != nil {
		t.Error("PStrSlice should remain nil")
	}
	if s.PStrMap != nil {
		t.Error("PStrMap should remain nil")
	}
	if s.PIntMap != nil {
		t.Error("PIntMap should remain nil")
	}
}

// TestNestedReset проверяет Reset вложенной структуры.
func TestNestedReset(t *testing.T) {
	s := &NestedReset{
		Simple: FullReset{
			IntVal:    42,
			StringVal: "hello",
			BoolVal:   true,
		},
		Child: &FullReset{
			IntVal:    100,
			StringVal: "world",
			BoolVal:   false,
		},
	}

	s.Reset()

	// Simple сбрасывается через Reset() (у FullReset есть Reset)
	if s.Simple.IntVal != 0 {
		t.Errorf("Simple.IntVal = %d, want 0", s.Simple.IntVal)
	}
	if s.Simple.StringVal != "" {
		t.Errorf("Simple.StringVal = %q, want empty", s.Simple.StringVal)
	}
	if s.Simple.BoolVal != false {
		t.Errorf("Simple.BoolVal = %v, want false", s.Simple.BoolVal)
	}

	// Child — указатель на FullReset с Reset()
	if s.Child == nil {
		t.Fatal("Child is nil after Reset, expected non-nil pointer")
	}
	if s.Child.IntVal != 0 {
		t.Errorf("Child.IntVal = %d, want 0", s.Child.IntVal)
	}
	if s.Child.StringVal != "" {
		t.Errorf("Child.StringVal = %q, want empty", s.Child.StringVal)
	}
}

// TestSomeStructReset проверяет простую структуру.
func TestSomeStructReset(t *testing.T) {
	s := &SomeStruct{
		X: 42,
		Y: "hello",
	}

	s.Reset()

	if s.X != 0 {
		t.Errorf("X = %d, want 0", s.X)
	}
	if s.Y != "" {
		t.Errorf("Y = %q, want empty", s.Y)
	}
}

// TestWithEmbedded проверяет Reset с встроенными структурами.
func TestWithEmbedded(t *testing.T) {
	s := &WithEmbedded{
		SomeStruct: SomeStruct{X: 42, Y: "hello"},
		FullReset: &FullReset{
			IntVal:    100,
			StringVal: "world",
		},
		StrMap: map[string]string{"key": "value"},
		Ints:   []int{1, 2, 3},
	}

	s.Reset()

	// SomeStruct — embedded, у неё есть Reset()
	if s.X != 0 {
		t.Errorf("X = %d, want 0", s.X)
	}
	if s.Y != "" {
		t.Errorf("Y = %q, want empty", s.Y)
	}

	// FullReset — embedded *FullReset с Reset()
	if s.FullReset == nil {
		t.Fatal("FullReset is nil after Reset, expected non-nil pointer")
	}
	if s.FullReset.IntVal != 0 {
		t.Errorf("FullReset.IntVal = %d, want 0", s.FullReset.IntVal)
	}
	if s.FullReset.StringVal != "" {
		t.Errorf("FullReset.StringVal = %q, want empty", s.FullReset.StringVal)
	}

	// StrMap и Ints
	if len(s.StrMap) != 0 {
		t.Errorf("StrMap length = %d, want 0", len(s.StrMap))
	}
	if len(s.Ints) != 0 {
		t.Errorf("Ints length = %d, want 0", len(s.Ints))
	}
}

// TestOuterWithInnerGeneric проверяет, что вложенная структура без Reset() с embedded
// generic полем корректно обрабатывается (inline reset через generateInlineReset,
// для Bar[int] выводится WARNING, остальные поля сбрасываются).
func TestOuterWithInnerGeneric(t *testing.T) {
	s := &OuterWithInnerGeneric{
		Inner: InnerWithGeneric{X: 42},
		Y:     "hello",
	}

	s.Reset()

	// X сбрасывается (обычное поле InnerWithGeneric)
	if s.Inner.X != 0 {
		t.Errorf("Inner.X = %d, want 0", s.Inner.X)
	}
	// Y сбрасывается
	if s.Y != "" {
		t.Errorf("Y = %q, want empty", s.Y)
	}
}
