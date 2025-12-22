package gg

import (
	"strings"
	"testing"
)

func TestMerge_PackageAliasConflict(t *testing.T) {
	// Generator A: uses github.com/google/a/book
	genA := New()
	genA.SetPackage("example")

	bookGoogle := genA.P("github.com/google/a/book")

	genA.Body().NewFunction("ProcessGoogleBook").
		AddParameter("b", bookGoogle.Type("Book")).
		AddResult("", "error").
		AddBody(
			S("return nil"),
		)

	// Generator B: uses github.com/facebook/book (same package name, different path)
	genB := New()
	genB.SetPackage("example")

	bookFacebook := genB.P("github.com/facebook/book")

	genB.Body().NewFunction("ProcessFacebookBook").
		AddParameter("b", bookFacebook.Type("Book")).
		AddResult("", "error").
		AddBody(
			S("return nil"),
		)

	// Merge B into A
	genA.Merge(genB)

	output := genA.String()

	// Verify imports: book2 should be aliased
	if !strings.Contains(output, `book2 "github.com/facebook/book"`) {
		t.Errorf("Expected aliased import book2, got:\n%s", output)
	}

	if !strings.Contains(output, `"github.com/google/a/book"`) {
		t.Errorf("Expected google book import, got:\n%s", output)
	}

	// Verify function signatures: Google book should use "book", Facebook should use "book2"
	if !strings.Contains(output, "ProcessGoogleBook(b book.Book)") {
		t.Errorf("Expected ProcessGoogleBook to use book.Book, got:\n%s", output)
	}

	if !strings.Contains(output, "ProcessFacebookBook(b book2.Book)") {
		t.Errorf("Expected ProcessFacebookBook to use book2.Book, got:\n%s", output)
	}
}

func TestMerge_ComplexTypes(t *testing.T) {
	// Generator A
	genA := New()
	genA.SetPackage("main")

	typesA := genA.P("github.com/example/types")

	genA.Body().NewVar().AddField("users", Slice(typesA.Type("User"),
		S("types.User{ID: 1}"),
	))

	// Generator B with conflicting package name
	genB := New()
	genB.SetPackage("main")

	typesB := genB.P("github.com/other/types")

	genB.Body().NewVar().AddField("configs", Slice(typesB.Ptr("Config"),
		S("types.Config{}"),
	))

	// Merge
	genA.Merge(genB)

	output := genA.String()

	// Verify imports
	if !strings.Contains(output, `"github.com/example/types"`) {
		t.Errorf("Expected example/types import, got:\n%s", output)
	}

	if !strings.Contains(output, `types2 "github.com/other/types"`) {
		t.Errorf("Expected aliased types2 import, got:\n%s", output)
	}

	// Verify slice types are updated
	if !strings.Contains(output, "[]types.User{") {
		t.Errorf("Expected []types.User, got:\n%s", output)
	}

	if !strings.Contains(output, "[]*types2.Config{") {
		t.Errorf("Expected []*types2.Config, got:\n%s", output)
	}
}

func TestMerge_SamePackagePath(t *testing.T) {
	// When both generators use the same import path, no renaming should occur
	genA := New()
	genA.SetPackage("main")

	types := genA.P("github.com/example/types")
	genA.Body().NewVar().AddField("u1", types.Type("User"))

	genB := New()
	genB.SetPackage("main")

	types2 := genB.P("github.com/example/types")
	genB.Body().NewVar().AddField("u2", types2.Type("User"))

	// Merge
	genA.Merge(genB)

	output := genA.String()

	// Should only have one import
	importCount := strings.Count(output, `"github.com/example/types"`)
	if importCount != 1 {
		t.Errorf("Expected 1 import statement, found %d in:\n%s", importCount, output)
	}

	// Both should use "types" (no alias) - check with flexible spacing
	if !strings.Contains(output, "u1") || !strings.Contains(output, "types.User") {
		t.Errorf("Expected u1 to use types.User, got:\n%s", output)
	}

	if !strings.Contains(output, "u2") || !strings.Contains(output, "types.User") {
		t.Errorf("Expected u2 to use types.User, got:\n%s", output)
	}

	// Should NOT have types2
	if strings.Contains(output, "types2") {
		t.Errorf("Should not have types2 alias, got:\n%s", output)
	}
}

func TestMerge_ThreeWayConflict(t *testing.T) {
	// Test three different packages with the same name
	genA := New()
	genA.SetPackage("main")

	pkg1 := genA.P("github.com/a/types")
	genA.Body().NewVar().AddField("v1", pkg1.Type("T"))

	genB := New()
	genB.SetPackage("main")

	pkg2 := genB.P("github.com/b/types")
	genB.Body().NewVar().AddField("v2", pkg2.Type("T"))

	genC := New()
	genC.SetPackage("main")

	pkg3 := genC.P("github.com/c/types")
	genC.Body().NewVar().AddField("v3", pkg3.Type("T"))

	// Merge all
	genA.Merge(genB)
	genA.Merge(genC)

	output := genA.String()

	// Verify all three imports with correct aliases
	if !strings.Contains(output, `"github.com/a/types"`) {
		t.Errorf("Expected a/types import, got:\n%s", output)
	}

	if !strings.Contains(output, `types2 "github.com/b/types"`) {
		t.Errorf("Expected types2 alias for b/types, got:\n%s", output)
	}

	if !strings.Contains(output, `types3 "github.com/c/types"`) {
		t.Errorf("Expected types3 alias for c/types, got:\n%s", output)
	}

	// Verify variables use correct aliases - flexible spacing
	if !strings.Contains(output, "v1") || !strings.Contains(output, "types.T") {
		t.Errorf("Expected v1 to use types.T, got:\n%s", output)
	}

	if !strings.Contains(output, "v2") || !strings.Contains(output, "types2.T") {
		t.Errorf("Expected v2 to use types2.T, got:\n%s", output)
	}

	if !strings.Contains(output, "v3") || !strings.Contains(output, "types3.T") {
		t.Errorf("Expected v3 to use types3.T, got:\n%s", output)
	}
}
