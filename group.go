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

	isfirst := true
	for _, node := range g.items {
		if !isfirst {
			writeString(w, g.separator)
		}
		node.render(w)
		isfirst = false
	}

	if g.close != "" && !g.shouldOmitWrap() {
		writeString(w, g.close)
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
