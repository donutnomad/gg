package gg

import (
	"strings"
	"testing"
)

func TestPackageRef_Basic(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	// Create package references
	types := gen.P("github.com/example/types")

	// Test type generation
	gen.Body().NewFunction("test").
		AddParameter("user", types.Type("User")).
		AddResult("", "error").
		AddBody("return nil")

	output := gen.String()

	// Verify package declaration
	if !strings.Contains(output, "package main") {
		t.Errorf("Expected package declaration, got:\n%s", output)
	}

	// Verify import
	if !strings.Contains(output, `"github.com/example/types"`) {
		t.Errorf("Expected import statement, got:\n%s", output)
	}

	// Verify qualified type
	if !strings.Contains(output, "types.User") {
		t.Errorf("Expected types.User, got:\n%s", output)
	}
}

func TestPackageRef_AliasConflict(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	// Create two packages with the same base name
	types1 := gen.P("github.com/example/types")
	types2 := gen.P("github.com/other/types")

	// Verify they get different aliases
	if types1.Alias() == types2.Alias() {
		t.Errorf("Expected different aliases, got %s and %s", types1.Alias(), types2.Alias())
	}

	// First should be "types", second should be "types2"
	if types1.Alias() != "types" {
		t.Errorf("Expected first alias to be 'types', got '%s'", types1.Alias())
	}
	if types2.Alias() != "types2" {
		t.Errorf("Expected second alias to be 'types2', got '%s'", types2.Alias())
	}

	// Use both in code
	gen.Body().NewVar().
		AddField("u1", types1.Type("User")).
		AddField("u2", types2.Type("User"))

	output := gen.String()

	// Verify both qualified types appear
	if !strings.Contains(output, "types.User") {
		t.Errorf("Expected types.User, got:\n%s", output)
	}
	if !strings.Contains(output, "types2.User") {
		t.Errorf("Expected types2.User, got:\n%s", output)
	}

	// Verify aliased import
	if !strings.Contains(output, `types2 "github.com/other/types"`) {
		t.Errorf("Expected aliased import for types2, got:\n%s", output)
	}
}

func TestPackageRef_ExplicitAlias(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	// Create package with explicit alias
	ctx := gen.PAlias("context", "ctx")

	gen.Body().NewFunction("test").
		AddParameter("c", ctx.Type("Context")).
		AddBody("return")

	output := gen.String()

	// Verify aliased import
	if !strings.Contains(output, `ctx "context"`) {
		t.Errorf("Expected ctx alias for context, got:\n%s", output)
	}

	// Verify qualified type
	if !strings.Contains(output, "ctx.Context") {
		t.Errorf("Expected ctx.Context, got:\n%s", output)
	}
}

func TestPackageRef_SliceAndPtr(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	types := gen.P("github.com/example/types")

	gen.Body().NewFunction("test").
		AddParameter("users", types.Slice("User")).
		AddParameter("config", types.Ptr("Config")).
		AddBody("return")

	output := gen.String()

	// Verify slice type
	if !strings.Contains(output, "[]types.User") {
		t.Errorf("Expected []types.User, got:\n%s", output)
	}

	// Verify pointer type
	if !strings.Contains(output, "*types.Config") {
		t.Errorf("Expected *types.Config, got:\n%s", output)
	}
}

func TestPackageRef_Call(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	types := gen.P("github.com/example/types")

	gen.Body().NewFunction("test").
		AddResult("", types.Ptr("User")).
		AddBody(S("return %s", types.Call("NewUser", Lit("name")).String()))

	output := gen.String()

	// Verify function call - note: Call generates "types.NewUser(...)"
	if !strings.Contains(output, "types.NewUser") {
		t.Errorf("Expected types.NewUser call, got:\n%s", output)
	}
}

func TestPackageRef_StdLib(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	// Standard library
	fmt := gen.P("fmt")
	time := gen.P("time")

	// Third party
	types := gen.P("github.com/example/types")

	gen.Body().NewFunction("test").
		AddParameter("t", time.Type("Time")).
		AddBody(S("%s.Println(%s)", fmt.Dot(""), types.Dot("DefaultMessage")))

	output := gen.String()

	// Verify stdlib comes before third-party (with blank line between)
	fmtIdx := strings.Index(output, `"fmt"`)
	timeIdx := strings.Index(output, `"time"`)
	typesIdx := strings.Index(output, `"github.com/example/types"`)

	if fmtIdx == -1 || timeIdx == -1 || typesIdx == -1 {
		t.Errorf("Missing imports in output:\n%s", output)
	}

	// Stdlib should come before third-party
	if fmtIdx > typesIdx || timeIdx > typesIdx {
		t.Errorf("Expected stdlib imports before third-party, got:\n%s", output)
	}
}

func TestPackageRef_ReuseExisting(t *testing.T) {
	gen := New()

	// Get the same package twice
	types1 := gen.P("github.com/example/types")
	types2 := gen.P("github.com/example/types")

	// Should be the same reference
	if types1 != types2 {
		t.Errorf("Expected same PackageRef instance for same import path")
	}

	// Should only have one import
	if len(gen.Imports()) != 1 {
		t.Errorf("Expected 1 import, got %d", len(gen.Imports()))
	}
}

func TestResolvePackageAlias(t *testing.T) {
	tests := []struct {
		importPath string
		existing   map[string]bool
		expected   string
	}{
		{"fmt", nil, "fmt"},
		{"context", nil, "context"},
		{"github.com/example/types", nil, "types"},
		{"github.com/example/pkg/v2", nil, "pkg"},
		{"github.com/example/types", map[string]bool{"types": true}, "types2"},
		{"github.com/example/types", map[string]bool{"types": true, "types2": true}, "types3"},
		{"github.com/example/some-pkg", nil, "some_pkg"},
		{"github.com/example/123pkg", nil, "_23pkg"},
	}

	for _, tt := range tests {
		t.Run(tt.importPath, func(t *testing.T) {
			result := resolvePackageAlias(tt.importPath, tt.existing)
			if result != tt.expected {
				t.Errorf("resolvePackageAlias(%q, %v) = %q, want %q",
					tt.importPath, tt.existing, result, tt.expected)
			}
		})
	}
}

func TestSanitizeIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"types", "types"},
		{"some-pkg", "some_pkg"},
		{"123pkg", "_23pkg"},
		{"my.pkg", "my_pkg"},
		{"", "pkg"},
		{"ValidName", "ValidName"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeIdentifier(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestPackageRef_Chan(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	events := gen.P("github.com/example/events")

	gen.Body().NewFunction("test").
		AddParameter("ch", events.Chan("Event")).
		AddParameter("recv", events.ChanRecv("Event")).
		AddParameter("send", events.ChanSend("Event")).
		AddBody("return")

	output := gen.String()

	// Verify channel types
	if !strings.Contains(output, "chan events.Event") {
		t.Errorf("Expected chan events.Event, got:\n%s", output)
	}
	if !strings.Contains(output, "<-chan events.Event") {
		t.Errorf("Expected <-chan events.Event, got:\n%s", output)
	}
	if !strings.Contains(output, "chan<- events.Event") {
		t.Errorf("Expected chan<- events.Event, got:\n%s", output)
	}
}

func TestPackageRef_Generic(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	types := gen.P("github.com/example/types")

	gen.Body().NewFunction("test").
		AddParameter("list", types.Generic("List", "string")).
		AddParameter("map1", types.Generic("Map", "string", "User")).
		AddResult("", types.Generic("Result", types.Type("Data"), "error")).
		AddBody("return nil")

	output := gen.String()

	// Verify generic types
	if !strings.Contains(output, "types.List[types.string]") {
		t.Errorf("Expected types.List[types.string], got:\n%s", output)
	}
	if !strings.Contains(output, "types.Map[types.string, types.User]") {
		t.Errorf("Expected types.Map[types.string, types.User], got:\n%s", output)
	}
	if !strings.Contains(output, "types.Result[types.Data, types.error]") {
		t.Errorf("Expected types.Result[types.Data, types.error], got:\n%s", output)
	}
}

func TestPackageRef_FlexibleMap(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	types1 := gen.P("github.com/example/types")
	types2 := gen.P("github.com/other/types")

	gen.Body().NewVar().
		// map[types1.Key]types1.Value
		AddField("m1", types1.Map("Key", "Value")).
		// map[types2.Key]types1.Value - cross-package
		AddField("m2", types1.Map(types2.Type("Key"), "Value"))

	output := gen.String()

	// Verify flexible map types
	if !strings.Contains(output, "map[types.Key]types.Value") {
		t.Errorf("Expected map[types.Key]types.Value, got:\n%s", output)
	}
	if !strings.Contains(output, "map[types2.Key]types.Value") {
		t.Errorf("Expected map[types2.Key]types.Value, got:\n%s", output)
	}
}

func TestResolvePackageAlias_ManyConflicts(t *testing.T) {
	// Test counter > 9 (the bug fix)
	existing := map[string]bool{
		"types":   true,
		"types2":  true,
		"types3":  true,
		"types4":  true,
		"types5":  true,
		"types6":  true,
		"types7":  true,
		"types8":  true,
		"types9":  true,
		"types10": true,
	}

	result := resolvePackageAlias("github.com/example/types", existing)
	if result != "types11" {
		t.Errorf("Expected types11, got %s", result)
	}
}
