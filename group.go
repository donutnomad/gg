package gg

import (
	"fmt"
	"io"
	"os"
)

func NewGroup() *Group {
	return newGroup("", "", "\n")
}

// NewInlineGroup creates a group with no separators between elements.
// Use this when you need to combine multiple nodes into a single line.
//
// Example:
//
//	gg.NewInlineGroup().Append(
//	    gg.S("tn := "),
//	    gsqlPkg.Call("TableName", gg.Lit("tableName")),
//	)
//	// => tn := gsqlPkg.TableName("tableName")
func NewInlineGroup() *Group {
	return newGroup("", "", "")
}

func newGroup(open, close, sep string) *Group {
	return &Group{
		open:      open,
		close:     close,
		separator: sep,
	}
}

type Group struct {
	items     []Node
	open      string
	close     string
	separator string

	// NewIf this result is true, we will omit the wrap like `()`, `{}`.
	omitWrapIf func() bool

	// mergeFields when true, consecutive fields with the same type will be merged.
	// Example: (a string, b string, c int) => (a, b string, c int)
	mergeFields bool
}

func (g *Group) length() int {
	return len(g.items)
}

func (g *Group) shouldOmitWrap() bool {
	if g.omitWrapIf == nil {
		return false
	}
	return g.omitWrapIf()
}

func (g *Group) append(node ...interface{}) *Group {
	if len(node) == 0 {
		return g
	}
	g.items = append(g.items, parseNodes(node)...)
	return g
}

// Append adds nodes to the group
func (g *Group) Append(node ...any) *Group {
	return g.append(node...)
}

func (g *Group) render(w io.Writer) {
	if g.open != "" && !g.shouldOmitWrap() {
		writeString(w, g.open)
	}

	if g.mergeFields {
		g.renderMergedFields(w)
	} else {
		isfirst := true
		for _, node := range g.items {
			if !isfirst {
				writeString(w, g.separator)
			}
			node.render(w)
			isfirst = false
		}
	}

	if g.close != "" && !g.shouldOmitWrap() {
		writeString(w, g.close)
	}
}

// renderMergedFields renders fields with consecutive same types merged.
// Example: (a string, b string, c int) => (a, b string, c int)
func (g *Group) renderMergedFields(w io.Writer) {
	type fieldInfo struct {
		name string
		typ  string
	}

	// Extract field information
	var fields []fieldInfo
	for _, node := range g.items {
		if f, ok := node.(*ifield); ok {
			// Get name and type as strings
			nameBuf := pool.Get()
			f.name.render(nameBuf)
			name := nameBuf.String()
			nameBuf.Free()

			typBuf := pool.Get()
			f.value.render(typBuf)
			typ := typBuf.String()
			typBuf.Free()

			fields = append(fields, fieldInfo{name: name, typ: typ})
		} else if mf, ok := node.(*multiNameField); ok {
			// multiNameField already handles multiple names with same type
			typBuf := pool.Get()
			mf.typ.render(typBuf)
			typ := typBuf.String()
			typBuf.Free()

			for _, name := range mf.names {
				fields = append(fields, fieldInfo{name: name, typ: typ})
			}
		} else {
			// Unknown node type, render as-is
			node.render(w)
			continue
		}
	}

	// Check if all field names are empty (unnamed parameters/results)
	allNamesEmpty := true
	for _, f := range fields {
		if f.name != "" {
			allNamesEmpty = false
			break
		}
	}

	// If all names are empty, don't merge - just output types
	if allNamesEmpty {
		isfirst := true
		for _, f := range fields {
			if !isfirst {
				writeString(w, g.separator)
			}
			writeString(w, f.typ)
			isfirst = false
		}
		return
	}

	// Group consecutive fields with the same type
	isfirst := true
	i := 0
	for i < len(fields) {
		if !isfirst {
			writeString(w, g.separator)
		}

		// Find all consecutive fields with the same type
		j := i + 1
		for j < len(fields) && fields[j].typ == fields[i].typ {
			j++
		}

		// Write names
		for k := i; k < j; k++ {
			if k > i {
				writeString(w, ", ")
			}
			writeString(w, fields[k].name)
		}

		// Write type
		writeString(w, " ")
		writeString(w, fields[i].typ)

		isfirst = false
		i = j
	}
}

// Deprecated: use `Generator.Write(w)` instead.
func (g *Group) Write(w io.Writer) {
	g.render(w)
}

// Deprecated: use `Generator.WriteFile(w)` instead.
func (g *Group) WriteFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file %s: %s", path, err)
	}
	g.render(file)
	return nil
}

// Deprecated: use `Generator.AppendFile(w)` instead.
func (g *Group) AppendFile(path string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("create file %s: %s", path, err)
	}
	g.render(file)
	return nil
}

func (g *Group) String() string {
	buf := pool.Get()
	defer buf.Free()

	g.render(buf)
	return buf.String()
}

func (g *Group) AddLineComment(content string, args ...interface{}) *Group {
	g.append(LineComment(content, args...))
	return g
}

func (g *Group) AddPackage(name string) *Group {
	g.append(Package(name))
	return g
}

func (g *Group) AddLine() *Group {
	g.append(Line())
	return g
}

func (g *Group) AddString(content string, args ...interface{}) *Group {
	g.append(S(content, args...))
	return g
}

func (g *Group) AddType(name string, typ interface{}) *Group {
	g.append(Type(name, typ))
	return g
}

func (g *Group) AddTypeAlias(name string, typ interface{}) *Group {
	g.append(TypeAlias(name, typ))
	return g
}

func (g *Group) NewImport() *iimport {
	i := Import()
	g.append(i)
	return i
}

func (g *Group) NewIf(judge interface{}) *iif {
	i := If(judge)
	g.append(i)
	return i
}

func (g *Group) NewFor(judge interface{}) *ifor {
	i := For(judge)
	g.append(i)
	return i
}

func (g *Group) NewSwitch(judge interface{}) *iswitch {
	i := Switch(judge)
	g.append(i)
	return i
}

func (g *Group) NewVar() *ivar {
	i := Var()
	g.append(i)
	return i
}

func (g *Group) NewConst() *iconst {
	i := Const()
	g.append(i)
	return i
}

func (g *Group) NewFunction(name string) *ifunction {
	f := Function(name)
	g.append(f)
	return f
}

func (g *Group) NewStruct(name string) *istruct {
	i := Struct(name)
	g.append(i)
	return i
}

func (g *Group) NewInterface(name string) *iinterface {
	i := Interface(name)
	g.append(i)
	return i
}
