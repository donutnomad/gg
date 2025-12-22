package gg

import (
	"strings"
	"testing"
)

func TestValue_StructLiteral(t *testing.T) {
	v := Value("User")
	v.AddField("ID", "1")
	v.AddField("Name", Lit("Alice"))

	expected := `User{ID:1, Name:"Alice"}`
	got := v.String()

	compareAST(t, expected, got)
}

func TestValue_CustomTypeAlias(t *testing.T) {
	// Test for custom type like: type UserList []User
	v := Value("UserList")
	v.AddElement(S("User{ID: 1}"), S("User{ID: 2}"), S("User{ID: 3}"))

	expected := `UserList{User{ID: 1}, User{ID: 2}, User{ID: 3}}`
	got := v.String()

	compareAST(t, expected, got)
}

func TestValue_WithPackageRef(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	types := gen.P("github.com/example/types")

	// Custom type alias from external package
	gen.Body().NewVar().AddField("users",
		Value(types.Type("UserList")).AddElement(
			S("types.User{ID: 1}"),
			S("types.User{ID: 2}"),
		),
	)

	output := gen.String()

	// Should generate: var users = types.UserList{types.User{ID: 1}, types.User{ID: 2}}
	if !strings.Contains(output, "types.UserList{") {
		t.Errorf("Expected types.UserList literal, got:\n%s", output)
	}

	if !strings.Contains(output, "types.User{ID: 1}") {
		t.Errorf("Expected user elements, got:\n%s", output)
	}
}

func TestValue_EmptyLiteral(t *testing.T) {
	v := Value("Config")

	expected := `Config{}`
	got := v.String()

	compareAST(t, expected, got)
}

func TestValue_MapTypeLiteral(t *testing.T) {
	// Custom map type: type StringMap map[string]string
	v := Value("StringMap")
	v.AddField(Lit("key1"), Lit("value1"))
	v.AddField(Lit("key2"), Lit("value2"))

	expected := `StringMap{"key1":"value1", "key2":"value2"}`
	got := v.String()

	compareAST(t, expected, got)
}

func TestValue_ChainedAddElement(t *testing.T) {
	v := Value("IntList").
		AddElement(Lit(1)).
		AddElement(Lit(2), Lit(3)).
		AddElement(Lit(4))

	expected := `IntList{1, 2, 3, 4}`
	got := v.String()

	compareAST(t, expected, got)
}
