package gg

import (
	"fmt"
	"io"
	"path"
	"strings"
)

// PackageRef represents a reference to an external package.
// It provides methods to generate qualified identifiers and automatically
// manages import aliases to avoid conflicts.
type PackageRef struct {
	importPath string     // full import path: "github.com/example/types"
	alias      string     // resolved alias: "types" or "types2" if conflict
	gen        *Generator // back-reference to generator for import management
}

// ImportPath returns the full import path of the package.
func (p *PackageRef) ImportPath() string {
	return p.importPath
}

// Alias returns the resolved alias for this package.
func (p *PackageRef) Alias() string {
	return p.alias
}

// Type returns a qualified type name node.
// Example: types.User
func (p *PackageRef) Type(name string) Node {
	return &qualifiedIdent{pkg: p, name: name}
}

// Dot returns a qualified identifier node (for constants, variables, etc.).
// Example: types.DefaultTimeout
func (p *PackageRef) Dot(name string) Node {
	return &qualifiedIdent{pkg: p, name: name}
}

// Call returns a qualified function call.
// Example: types.NewUser("name")
func (p *PackageRef) Call(name string, args ...any) *icall {
	c := Call(name)
	c.owner = String(p.alias)
	c.AddParameter(args...)
	return c
}

// Func returns a qualified function reference (without calling it).
// Example: types.ParseConfig
func (p *PackageRef) Func(name string) Node {
	return &qualifiedIdent{pkg: p, name: name}
}

// Slice returns a slice type of the given type name.
// Example: []types.User
func (p *PackageRef) Slice(name string) Node {
	return &sliceType{elem: &qualifiedIdent{pkg: p, name: name}}
}

// Ptr returns a pointer type of the given type name.
// Example: *types.User
func (p *PackageRef) Ptr(name string) Node {
	return &ptrType{elem: &qualifiedIdent{pkg: p, name: name}}
}

// Map returns a map type with the given key and value types.
// Both key and value can be: string (type name), Node, or *PackageRef method result.
// Example:
//
//	types.Map("string", "User")           // map[string]types.User
//	types.Map(other.Type("Key"), "Value") // map[other.Key]types.Value
//	Map(types.Type("Key"), types.Type("Value")) // map[types.Key]types.Value (use package-level Map)
func (p *PackageRef) Map(keyType any, valueType any) Node {
	return &mapType{
		key:   p.resolveType(keyType),
		value: p.resolveType(valueType),
	}
}

// Chan returns a channel type of the given type name.
// Example: chan types.Event
func (p *PackageRef) Chan(name string) Node {
	return &chanType{elem: &qualifiedIdent{pkg: p, name: name}, dir: chanBoth}
}

// ChanRecv returns a receive-only channel type.
// Example: <-chan types.Event
func (p *PackageRef) ChanRecv(name string) Node {
	return &chanType{elem: &qualifiedIdent{pkg: p, name: name}, dir: chanRecv}
}

// ChanSend returns a send-only channel type.
// Example: chan<- types.Event
func (p *PackageRef) ChanSend(name string) Node {
	return &chanType{elem: &qualifiedIdent{pkg: p, name: name}, dir: chanSend}
}

// Generic returns a generic type instantiation.
// Example:
//
//	types.Generic("List", "string")              // types.List[string]
//	types.Generic("Map", "string", types.Type("User")) // types.Map[string, types.User]
func (p *PackageRef) Generic(name string, typeArgs ...any) Node {
	args := make([]Node, len(typeArgs))
	for i, arg := range typeArgs {
		args[i] = p.resolveType(arg)
	}
	return &genericType{
		base: &qualifiedIdent{pkg: p, name: name},
		args: args,
	}
}

// resolveType converts various input types to Node.
// Accepts: string (becomes qualified type), Node (passed through), or any (uses parseNode).
func (p *PackageRef) resolveType(t any) Node {
	switch v := t.(type) {
	case string:
		return &qualifiedIdent{pkg: p, name: v}
	case Node:
		return v
	default:
		return parseNode(t)
	}
}

// qualifiedIdent represents a qualified identifier like "pkg.Name"
type qualifiedIdent struct {
	pkg  *PackageRef
	name string
}

func (q *qualifiedIdent) render(w io.Writer) {
	if q.pkg.alias != "" {
		writeString(w, q.pkg.alias)
		if q.name != "" {
			writeString(w, ".")
		}
	}
	writeString(w, q.name)
}

func (q *qualifiedIdent) String() string {
	buf := pool.Get()
	defer buf.Free()
	q.render(buf)
	return buf.String()
}

// sliceType represents a slice type like []T
type sliceType struct {
	elem Node
}

func (s *sliceType) render(w io.Writer) {
	writeString(w, "[]")
	s.elem.render(w)
}

// ptrType represents a pointer type like *T
type ptrType struct {
	elem Node
}

func (p *ptrType) render(w io.Writer) {
	writeString(w, "*")
	p.elem.render(w)
}

// mapType represents a map type like map[K]V
type mapType struct {
	key   Node
	value Node
}

func (m *mapType) render(w io.Writer) {
	writeString(w, "map[")
	m.key.render(w)
	writeString(w, "]")
	m.value.render(w)
}

// chanDir represents channel direction
type chanDir int

const (
	chanBoth chanDir = iota // chan T
	chanRecv                // <-chan T
	chanSend                // chan<- T
)

// chanType represents a channel type like chan T
type chanType struct {
	elem Node
	dir  chanDir
}

func (c *chanType) render(w io.Writer) {
	switch c.dir {
	case chanRecv:
		writeString(w, "<-chan ")
	case chanSend:
		writeString(w, "chan<- ")
	default:
		writeString(w, "chan ")
	}
	c.elem.render(w)
}

// genericType represents a generic type instantiation like T[A, B]
type genericType struct {
	base Node
	args []Node
}

func (g *genericType) render(w io.Writer) {
	g.base.render(w)
	writeString(w, "[")
	for i, arg := range g.args {
		if i > 0 {
			writeString(w, ", ")
		}
		arg.render(w)
	}
	writeString(w, "]")
}

// resolvePackageAlias extracts the base package name from an import path
// and resolves conflicts by appending a number suffix.
func resolvePackageAlias(importPath string, existingAliases map[string]bool) string {
	// Extract base name from import path
	baseName := path.Base(importPath)

	// Handle special cases like "v2", "v3" suffixes
	if strings.HasPrefix(baseName, "v") && len(baseName) <= 3 {
		// Get parent directory name instead
		parent := path.Dir(importPath)
		if parent != "." && parent != "/" {
			baseName = path.Base(parent)
		}
	}

	// Sanitize the base name (replace invalid chars)
	baseName = sanitizeIdentifier(baseName)

	// Check for conflicts and resolve
	alias := baseName
	counter := 2
	for existingAliases[alias] {
		alias = fmt.Sprintf("%s%d", baseName, counter)
		counter++
	}

	return alias
}

// sanitizeIdentifier ensures the string is a valid Go identifier
func sanitizeIdentifier(s string) string {
	if s == "" {
		return "pkg"
	}

	var result strings.Builder
	for i, r := range s {
		if i == 0 {
			// First character must be letter or underscore
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' {
				result.WriteRune(r)
			} else {
				result.WriteRune('_')
			}
		} else {
			// Subsequent characters can be letters, digits, or underscore
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
				result.WriteRune(r)
			} else if r == '-' || r == '.' {
				result.WriteRune('_')
			}
			// Skip other invalid characters
		}
	}

	if result.Len() == 0 {
		return "pkg"
	}
	return result.String()
}
