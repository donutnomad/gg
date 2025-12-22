package gg

import (
	"strings"
	"testing"
)

func TestSlice_MultiLine(t *testing.T) {
	s := Slice("string",
		Lit("apple"),
		Lit("banana"),
		Lit("cherry"),
	).MultiLine()

	output := s.String()

	// Should have newlines
	if !strings.Contains(output, "{\n") {
		t.Errorf("Expected multi-line format with opening brace on new line, got:\n%s", output)
	}

	if !strings.Contains(output, ",\n}") {
		t.Errorf("Expected multi-line format with trailing comma, got:\n%s", output)
	}

	// Each element should be on its own line
	lines := strings.Split(output, "\n")
	if len(lines) < 5 { // opening brace, 3 elements, closing brace
		t.Errorf("Expected at least 5 lines, got %d:\n%s", len(lines), output)
	}
}

func TestArray_MultiLine(t *testing.T) {
	a := Array(3, "int",
		Lit(1),
		Lit(2),
		Lit(3),
	).MultiLine()

	output := a.String()

	if !strings.Contains(output, "{\n") {
		t.Errorf("Expected multi-line format, got:\n%s", output)
	}

	if !strings.Contains(output, ",\n}") {
		t.Errorf("Expected trailing comma, got:\n%s", output)
	}
}

func TestValue_MultiLine(t *testing.T) {
	v := Value("UserList").
		AddElement(S("User{ID: 1}")).
		AddElement(S("User{ID: 2}")).
		AddElement(S("User{ID: 3}")).
		MultiLine()

	output := v.String()

	if !strings.Contains(output, "{\n") {
		t.Errorf("Expected multi-line format, got:\n%s", output)
	}

	if !strings.Contains(output, ",\n}") {
		t.Errorf("Expected trailing comma, got:\n%s", output)
	}
}

func TestValue_StructMultiLine(t *testing.T) {
	v := Value("Config").
		AddField("Host", Lit("localhost")).
		AddField("Port", "8080").
		AddField("Timeout", "30").
		MultiLine()

	output := v.String()

	if !strings.Contains(output, "{\n") {
		t.Errorf("Expected multi-line format, got:\n%s", output)
	}

	if !strings.Contains(output, "Host") {
		t.Errorf("Expected field Host, got:\n%s", output)
	}
}

func TestMultiLine_WithPackageRef(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	types := gen.P("github.com/example/types")

	gen.Body().NewVar().AddField("users",
		Slice(types.Type("User"),
			S("types.User{ID: 1, Name: \"Alice\"}"),
			S("types.User{ID: 2, Name: \"Bob\"}"),
			S("types.User{ID: 3, Name: \"Charlie\"}"),
		).MultiLine(),
	)

	output := gen.String()

	// Should have multi-line slice
	if !strings.Contains(output, "[]types.User{\n") {
		t.Errorf("Expected multi-line slice, got:\n%s", output)
	}

	// Should have all users
	if !strings.Contains(output, "Alice") || !strings.Contains(output, "Bob") || !strings.Contains(output, "Charlie") {
		t.Errorf("Expected all users, got:\n%s", output)
	}
}

func TestMultiLine_Empty(t *testing.T) {
	// Empty slice should still work with MultiLine
	s := Slice("string").MultiLine()

	output := s.String()

	// Empty slice should just be {}
	expected := "[]string{}"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestSlice_SingleLine_Default(t *testing.T) {
	// Without calling MultiLine(), should be single line
	s := Slice("int", Lit(1), Lit(2), Lit(3))

	output := s.String()

	// Should NOT have newlines
	if strings.Contains(output, "\n") {
		t.Errorf("Expected single-line format, got:\n%s", output)
	}

	expected := "[]int{1, 2, 3}"
	compareAST(t, expected, output)
}
