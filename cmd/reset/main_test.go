package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// makeCommentGroup создает *ast.CommentGroup из текста.
func makeCommentGroup(text string) *ast.CommentGroup {
	return &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "// " + text},
		},
	}
}

func TestHasGenerateReset(t *testing.T) {
	tests := []struct {
		name    string
		comment *ast.CommentGroup
		want    bool
	}{
		{
			name:    "directive present",
			comment: makeCommentGroup("generate:reset"),
			want:    true,
		},
		{
			name:    "directive with args",
			comment: makeCommentGroup("generate:reset Foo Bar"),
			want:    true,
		},
		{
			name:    "directive in middle of text should not match",
			comment: makeCommentGroup("Package foo // generate:reset"),
			want:    false,
		},
		{
			name:    "comment mentioning generate:reset should not match",
			comment: makeCommentGroup("Not marked generate:reset"),
			want:    false,
		},
		{
			name:    "no directive",
			comment: makeCommentGroup("Package foo"),
			want:    false,
		},
		{
			name:    "similar but different directive",
			comment: makeCommentGroup("generate:somethingelse"),
			want:    false,
		},
		{
			name:    "nil comment group",
			comment: nil,
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasGenerateReset(tt.comment); got != tt.want {
				t.Errorf("hasGenerateReset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFields(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		want     int // expected number of fields
		embedded int // expected number of embedded fields
	}{
		{
			name: "simple fields",
			src: `package test
type T struct {
	X int
	Y string
	Z bool
}`,
			want:     3,
			embedded: 0,
		},
		{
			name: "embedded fields",
			src: `package test
type T struct {
	Bar
	*Baz
	X int
}`,
			want:     3,
			embedded: 2,
		},
		{
			name: "mixed fields",
			src: `package test
type T struct {
	A int
	B string
	SomeStruct
	C float64
	*PointersReset
}`,
			want:     5,
			embedded: 2,
		},
		{
			name: "nil field list",
			want: 0,
		},
		{
			name: "multiple fields in one line",
			src: `package test
type T struct {
	A, B, C int
}`,
			want:     3,
			embedded: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fieldList *ast.FieldList
			if tt.src != "" {
				fset := token.NewFileSet()
				f, err := parser.ParseFile(fset, "", tt.src, parser.ParseComments)
				if err != nil {
					t.Fatal(err)
				}
				ts := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
				st := ts.Type.(*ast.StructType)
				fieldList = st.Fields
			}

			got := parseFields(fieldList)
			if len(got) != tt.want {
				t.Errorf("parseFields() returned %d fields, want %d", len(got), tt.want)
			}
			embeddedCount := 0
			for _, f := range got {
				if f.embedded {
					embeddedCount++
				}
			}
			if embeddedCount != tt.embedded {
				t.Errorf("parseFields() returned %d embedded fields, want %d", embeddedCount, tt.embedded)
			}
		})
	}
}

func TestEmbeddedFieldName(t *testing.T) {
	tests := []struct {
		name     string
		exprSrc  string
		want     string
		wantFail bool
	}{
		{
			name:    "simple ident",
			exprSrc: "Bar",
			want:    "Bar",
		},
		{
			name:    "pointer to ident",
			exprSrc: "*Bar",
			want:    "Bar",
		},
		{
			name:    "selector expr",
			exprSrc: "time.Time",
			want:    "Time",
		},
		{
			name:    "pointer to selector",
			exprSrc: "*time.Time",
			want:    "Time",
		},
		{
			name:    "pointer to selector with pkg",
			exprSrc: "*somepkg.SomeType",
			want:    "SomeType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Парсим выражение типа
			src := "package p\ntype T struct {\n\t" + tt.exprSrc + "\n}"
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", src, 0)
			if err != nil {
				t.Fatal(err)
			}
			ts := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
			st := ts.Type.(*ast.StructType)
			expr := st.Fields.List[0].Type

			got := embeddedFieldName(expr)
			if got != tt.want {
				t.Errorf("embeddedFieldName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStructHasReset(t *testing.T) {
	structs := []structInfo{
		{name: "Foo"},
		{name: "Bar"},
		{name: "Baz"},
	}
	tests := []struct {
		name     string
		typeName string
		want     bool
	}{
		{name: "exists", typeName: "Bar", want: true},
		{name: "not exists", typeName: "Quux", want: false},
		{name: "empty", typeName: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := structHasReset(structs, tt.typeName); got != tt.want {
				t.Errorf("structHasReset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExprToString(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{src: "int", want: "int"},
		{src: "string", want: "string"},
		{src: "time.Time", want: "time.Time"},
		{src: "map[string]string", want: "map[string]string"},
		{src: "[]int", want: "[]int"},
		{src: "*time.Duration", want: "*time.Duration"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			// Парсим выражение
			src := "package p\nvar x " + tt.src
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", src, 0)
			if err != nil {
				t.Fatal(err)
			}
			gs := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
			expr := gs.Type

			got := exprToString(expr)
			if got != tt.want {
				t.Errorf("exprToString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateResetMethod(t *testing.T) {
	allStructs := []structInfo{
		{name: "ChildStruct"},
		{name: "ParentStruct"},
	}
	pkgTypes := map[string]ast.Expr{
		"ChildStruct": &ast.StructType{
			Fields: &ast.FieldList{
				List: []*ast.Field{
					{Names: []*ast.Ident{{Name: "A"}}, Type: &ast.Ident{Name: "int"}},
				},
			},
		},
	}

	st := structInfo{
		pkgName: "testpkg",
		name:    "ParentStruct",
		fields: []fieldInfo{
			{name: "IntField", expr: &ast.Ident{Name: "int"}},
			{name: "StrField", expr: &ast.Ident{Name: "string"}},
			{name: "BoolField", expr: &ast.Ident{Name: "bool"}},
			{name: "SliceField", expr: &ast.ArrayType{Elt: &ast.Ident{Name: "int"}}},
			{name: "MapField", expr: &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "int"}}},
			{name: "Child", expr: &ast.StarExpr{X: &ast.Ident{Name: "ChildStruct"}}, embedded: false},
		},
	}

	result := generateResetMethod(st, allStructs, pkgTypes)

	// Проверяем основные элементы
	checks := []struct {
		name string
		want string
	}{
		{"func signature", "func (s *ParentStruct) Reset()"},
		{"nil check", "if s == nil {"},
		{"int field", "s.IntField = 0"},
		{"string field", `s.StrField = ""`},
		{"bool field", "s.BoolField = false"},
		{"slice field", "s.SliceField = s.SliceField[:0]"},
		{"map field", "clear(s.MapField)"},
		{"pointer child", "s.Child.Reset()"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !contains(result, c.want) {
				t.Errorf("generateResetMethod() missing %q\nGot:\n%s", c.want, result)
			}
		})
	}

	// Проверяем, что функция не пустая
	if len(result) == 0 {
		t.Error("generateResetMethod() returned empty string")
	}
}

func TestRenderFieldExpr(t *testing.T) {
	allStructs := []structInfo{
		{name: "ResettableStruct"},
	}
	pkgTypes := map[string]ast.Expr{
		"ResettableStruct": &ast.StructType{
			Fields: &ast.FieldList{
				List: []*ast.Field{
					{Names: []*ast.Ident{{Name: "X"}}, Type: &ast.Ident{Name: "int"}},
				},
			},
		},
	}

	tests := []struct {
		name     string
		expr     ast.Expr
		fieldRef string
		want     string
	}{
		{
			name:     "int field",
			expr:     &ast.Ident{Name: "int"},
			fieldRef: "s.X",
			want:     "s.X = 0",
		},
		{
			name:     "string field",
			expr:     &ast.Ident{Name: "string"},
			fieldRef: "s.S",
			want:     `s.S = ""`,
		},
		{
			name:     "bool field",
			expr:     &ast.Ident{Name: "bool"},
			fieldRef: "s.B",
			want:     "s.B = false",
		},
		{
			name:     "slice field",
			expr:     &ast.ArrayType{Elt: &ast.Ident{Name: "int"}},
			fieldRef: "s.Sl",
			want:     "s.Sl = s.Sl[:0]",
		},
		{
			name:     "map field",
			expr:     &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "int"}},
			fieldRef: "s.M",
			want:     "clear(s.M)",
		},
		{
			name:     "pointer to int",
			expr:     &ast.StarExpr{X: &ast.Ident{Name: "int"}},
			fieldRef: "s.Ptr",
			want:     "if s.Ptr != nil {",
		},
		{
			name:     "pointer to struct with Reset",
			expr:     &ast.StarExpr{X: &ast.Ident{Name: "ResettableStruct"}},
			fieldRef: "s.Child",
			want:     "s.Child.Reset()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderFieldExpr(tt.expr, tt.fieldRef, allStructs, pkgTypes)
			if !contains(got, tt.want) {
				t.Errorf("renderFieldExpr() = %q, want to contain %q", got, tt.want)
			}
		})
	}
}

// contains проверяет, содержит ли строка подстроку
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
