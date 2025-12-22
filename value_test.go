package gg

import (
	"strings"
	"testing"
)

func TestSlice_BasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		slice    *islice
		expected string
	}{
		{
			name:     "int slice",
			slice:    Slice("int", Lit(1), Lit(2), Lit(3)),
			expected: "[]int{1, 2, 3}",
		},
		{
			name:     "string slice",
			slice:    Slice("string", Lit("a"), Lit("b"), Lit("c")),
			expected: `[]string{"a", "b", "c"}`,
		},
		{
			name:     "empty slice",
			slice:    Slice("bool"),
			expected: "[]bool{}",
		},
		{
			name:     "float64 slice",
			slice:    Slice("float64", Lit(1.5), Lit(2.5), Lit(3.5)),
			expected: "[]float64{1.5, 2.5, 3.5}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := pool.Get()
			defer buf.Free()

			tt.slice.render(buf)
			got := buf.String()

			compareAST(t, tt.expected, got)
		})
	}
}

func TestSlice_WithPackageRef(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	types := gen.P("github.com/example/types")
	db := gen.P("github.com/example/db")

	// Test slice with PackageRef Type
	gen.Body().NewVar().AddField("users", Slice(types.Type("User"),
		S("types.User{ID: 1}"),
		S("types.User{ID: 2}"),
	))

	// Test slice with PackageRef Ptr
	gen.Body().NewVar().AddField("configs", Slice(types.Ptr("Config"),
		S("&types.Config{Name: \"dev\"}"),
		S("&types.Config{Name: \"prod\"}"),
	))

	// Test slice with multiple packages
	gen.Body().NewVar().AddField("mixed", Slice(db.Type("Connection"),
		S("db.Connection{}"),
	))

	output := gen.String()

	// Verify slice with Type
	if !strings.Contains(output, "[]types.User{") {
		t.Errorf("Expected []types.User slice, got:\n%s", output)
	}

	// Verify slice with Ptr
	if !strings.Contains(output, "[]*types.Config{") {
		t.Errorf("Expected []*types.Config slice, got:\n%s", output)
	}

	// Verify multiple packages
	if !strings.Contains(output, "[]db.Connection{") {
		t.Errorf("Expected []db.Connection slice, got:\n%s", output)
	}
}

func TestSlice_AddElement(t *testing.T) {
	s := Slice("int", Lit(1))
	s.AddElement(Lit(2), Lit(3))

	expected := "[]int{1, 2, 3}"
	got := s.String()

	compareAST(t, expected, got)
}

func TestArray_BasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		array    *iarray
		expected string
	}{
		{
			name:     "int array",
			array:    Array(3, "int", Lit(1), Lit(2), Lit(3)),
			expected: "[3]int{1, 2, 3}",
		},
		{
			name:     "string array",
			array:    Array(2, "string", Lit("hello"), Lit("world")),
			expected: `[2]string{"hello", "world"}`,
		},
		{
			name:     "empty array",
			array:    Array(5, "bool"),
			expected: "[5]bool{}",
		},
		{
			name:     "single element",
			array:    Array(1, "int", Lit(42)),
			expected: "[1]int{42}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := pool.Get()
			defer buf.Free()

			tt.array.render(buf)
			got := buf.String()

			compareAST(t, tt.expected, got)
		})
	}
}

func TestArray_WithPackageRef(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	types := gen.P("github.com/example/types")

	// Test array with PackageRef Type
	gen.Body().NewVar().AddField("users", Array(3, types.Type("User"),
		S("types.User{ID: 1}"),
		S("types.User{ID: 2}"),
		S("types.User{ID: 3}"),
	))

	// Test array with PackageRef Ptr
	gen.Body().NewVar().AddField("configs", Array(2, types.Ptr("Config"),
		S("&types.Config{}"),
		S("nil"),
	))

	output := gen.String()

	// Verify array with Type
	if !strings.Contains(output, "[3]types.User{") {
		t.Errorf("Expected [3]types.User array, got:\n%s", output)
	}

	// Verify array with Ptr
	if !strings.Contains(output, "[2]*types.Config{") {
		t.Errorf("Expected [2]*types.Config array, got:\n%s", output)
	}
}

func TestArray_AddElement(t *testing.T) {
	a := Array(5, "int", Lit(1), Lit(2))
	a.AddElement(Lit(3), Lit(4), Lit(5))

	expected := "[5]int{1, 2, 3, 4, 5}"
	got := a.String()

	compareAST(t, expected, got)
}

func TestSlice_ComplexTypes(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	types := gen.P("github.com/example/types")

	// Test slice of slice
	gen.Body().NewVar().AddField("matrix", Slice(
		S("[]int"),
		Slice("int", Lit(1), Lit(2)),
		Slice("int", Lit(3), Lit(4)),
	))

	// Test slice of map
	gen.Body().NewVar().AddField("maps", Slice(
		S("map[string]int"),
		S("map[string]int{\"a\": 1}"),
	))

	// Test slice with generic type - note: string becomes types.string when passed through types.Generic
	gen.Body().NewVar().AddField("generic", Slice(
		types.Generic("List", S("string")), // Use S("string") to get plain string
		S("types.List[string]{}"),
	))

	output := gen.String()

	if !strings.Contains(output, "[][]int{") {
		t.Errorf("Expected [][]int, got:\n%s", output)
	}

	if !strings.Contains(output, "[]map[string]int{") {
		t.Errorf("Expected []map[string]int, got:\n%s", output)
	}

	// Fixed expectation - Generic with S("string") gives us plain string
	if !strings.Contains(output, "[]types.List[string]{") {
		t.Errorf("Expected []types.List[string], got:\n%s", output)
	}
}

func TestSlice_InFunction(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	types := gen.P("github.com/example/types")

	gen.Body().NewFunction("getUsers").
		Return(types.Slice("User")).
		AddBody(
			Return(Slice(types.Type("User"),
				S("types.User{ID: 1, Name: \"Alice\"}"),
				S("types.User{ID: 2, Name: \"Bob\"}"),
			)),
		)

	output := gen.String()

	// Check function signature - with flexible spacing
	if !strings.Contains(output, "getUsers()") || !strings.Contains(output, "[]types.User") {
		t.Errorf("Expected function signature with []types.User, got:\n%s", output)
	}

	if !strings.Contains(output, "return []types.User{") {
		t.Errorf("Expected return with slice literal, got:\n%s", output)
	}

	// Verify both user elements are present
	if !strings.Contains(output, "ID: 1") || !strings.Contains(output, "ID: 2") {
		t.Errorf("Expected user data in slice, got:\n%s", output)
	}
}
