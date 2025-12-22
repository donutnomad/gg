package gg

import (
	"strings"
	"testing"
)

// TestMerge_InterfaceWithPackageRef tests that interface method signatures
// with PackageRef types are correctly updated during merge
func TestMerge_InterfaceWithPackageRef(t *testing.T) {
	// Generator A
	genA := New()
	genA.SetPackage("example")

	typesA := genA.P("github.com/google/types")

	// Create an interface with methods using PackageRef types
	iface := genA.Body().NewInterface("Service")
	iface.NewFunction("GetUser").
		AddParameter("id", "int64").
		AddResult("", typesA.Ptr("User")).
		AddResult("", "error")

	// Generator B
	genB := New()
	genB.SetPackage("example")

	typesB := genB.P("github.com/facebook/types")

	ifaceB := genB.Body().NewInterface("Repository")
	ifaceB.NewFunction("FindUser").
		AddParameter("id", "int64").
		AddResult("", typesB.Ptr("User")).
		AddResult("", "error")

	// Merge
	genA.Merge(genB)

	output := genA.String()

	// Verify that genA's interface still uses "types.User"
	if !strings.Contains(output, "*types.User") {
		t.Errorf("Expected '*types.User' in output, got:\n%s", output)
	}

	// Verify that genB's interface now uses "types2.User" after merge
	if !strings.Contains(output, "*types2.User") {
		t.Errorf("Expected '*types2.User' (dynamic rename) in output, got:\n%s", output)
	}

	// Verify function names
	if !strings.Contains(output, "GetUser") {
		t.Errorf("Expected 'GetUser' in output, got:\n%s", output)
	}

	if !strings.Contains(output, "FindUser") {
		t.Errorf("Expected 'FindUser' in output, got:\n%s", output)
	}

	// Verify imports
	if !strings.Contains(output, `"github.com/google/types"`) {
		t.Errorf("Expected google/types import, got:\n%s", output)
	}

	if !strings.Contains(output, `types2 "github.com/facebook/types"`) {
		t.Errorf("Expected facebook/types with alias types2, got:\n%s", output)
	}
}
