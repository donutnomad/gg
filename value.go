package gg

import "io"

type ivalue struct {
	typ       Node
	items     *Group
	multiLine bool
}

// Value creates a composite literal for a given type.
// It can be used for struct literals or custom type literals (like type aliases).
//
// For struct literals, use AddField:
//
//	Value("User").AddField("ID", "1").AddField("Name", Lit("Alice"))
//	// => User{ID: 1, Name: "Alice"}
//
// For custom types (like slice/array type aliases), use AddElement:
//
//	Value("UserList").AddElement(S("User{ID: 1}"), S("User{ID: 2}"))
//	// => UserList{User{ID: 1}, User{ID: 2}}
func Value(typ any) *ivalue {
	return &ivalue{
		typ:   parseNode(typ),
		items: newGroup("{", "}", ","),
	}
}

func (v *ivalue) render(w io.Writer) {
	v.typ.render(w)

	if v.multiLine && v.items.length() > 0 {
		// Multi-line format
		writeString(w, "{\n")
		first := true
		for _, item := range v.items.items {
			if !first {
				writeString(w, ",\n")
			}
			item.render(w)
			first = false
		}
		writeString(w, ",\n}")
	} else {
		// Single-line format
		v.items.render(w)
	}
}

func (v *ivalue) String() string {
	buf := pool.Get()
	defer buf.Free()

	v.render(buf)
	return buf.String()
}

// AddField adds a named field to the composite literal (for struct literals).
func (v *ivalue) AddField(name, value interface{}) *ivalue {
	v.items.append(field(name, value, ":"))
	return v
}

// AddElement adds an element without a name to the composite literal.
// This is useful for custom types like slice/array type aliases.
//
// Example:
//
//	Value("UserList").AddElement(S("User{ID: 1}"), S("User{ID: 2}"))
//	// => UserList{User{ID: 1}, User{ID: 2}}
func (v *ivalue) AddElement(elements ...any) *ivalue {
	v.items.append(elements...)
	return v
}

// MultiLine sets the value to use multi-line format where each element
// or field is on its own line.
//
// Example:
//
//	Value("UserList").AddElement(...).MultiLine()
//	// =>
//	// UserList{
//	//   User{ID: 1},
//	//   User{ID: 2},
//	// }
func (v *ivalue) MultiLine() *ivalue {
	v.multiLine = true
	return v
}

// islice represents a slice literal like []T{elem1, elem2, ...}
type islice struct {
	elemType  Node
	items     *Group
	multiLine bool
}

// Slice creates a slice literal with the given element type and elements.
// elemType can be a string (e.g., "int", "string") or a Node (e.g., pkg.Type("User")).
// elements are the items to include in the slice.
//
// Example:
//
//	Slice("int", Lit(1), Lit(2), Lit(3))                    // []int{1, 2, 3}
//	Slice("string", Lit("a"), Lit("b"))                     // []string{"a", "b"}
//	Slice(types.Type("User"), S("User{ID: 1}"))            // []types.User{User{ID: 1}}
//	Slice(types.Ptr("Config"), S("&Config{}"))             // []*types.Config{&Config{}}
func Slice(elemType any, elements ...any) *islice {
	s := &islice{
		elemType: parseNode(elemType),
		items:    newGroup("{", "}", ", "),
	}
	if len(elements) > 0 {
		s.items.append(elements...)
	}
	return s
}

func (s *islice) render(w io.Writer) {
	writeString(w, "[]")
	s.elemType.render(w)

	if s.multiLine && s.items.length() > 0 {
		// Multi-line format: {\n  elem1,\n  elem2,\n}
		writeString(w, "{\n")
		first := true
		for _, item := range s.items.items {
			if !first {
				writeString(w, ",\n")
			}
			item.render(w)
			first = false
		}
		writeString(w, ",\n}")
	} else {
		// Single-line format: {elem1, elem2}
		s.items.render(w)
	}
}

func (s *islice) String() string {
	buf := pool.Get()
	defer buf.Free()

	s.render(buf)
	return buf.String()
}

// AddElement adds elements to the slice literal.
func (s *islice) AddElement(elements ...any) *islice {
	if len(elements) > 0 {
		s.items.append(elements...)
	}
	return s
}

// MultiLine sets the slice to use multi-line format where each element
// is on its own line. This is useful for better readability when there
// are many elements.
//
// Example:
//
//	Slice("string", Lit("a"), Lit("b"), Lit("c")).MultiLine()
//	// =>
//	// []string{
//	//   "a",
//	//   "b",
//	//   "c",
//	// }
func (s *islice) MultiLine() *islice {
	s.multiLine = true
	return s
}

// iarray represents an array literal like [N]T{elem1, elem2, ...}
type iarray struct {
	size      int
	elemType  Node
	items     *Group
	multiLine bool
}

// Array creates an array literal with the given size, element type and elements.
// elemType can be a string (e.g., "int", "string") or a Node (e.g., pkg.Type("User")).
// elements are the items to include in the array.
//
// Example:
//
//	Array(3, "int", Lit(1), Lit(2), Lit(3))                 // [3]int{1, 2, 3}
//	Array(2, "string", Lit("a"), Lit("b"))                  // [2]string{"a", "b"}
//	Array(5, types.Type("User"), S("User{ID: 1}"))         // [5]types.User{User{ID: 1}}
func Array(size int, elemType any, elements ...any) *iarray {
	a := &iarray{
		size:     size,
		elemType: parseNode(elemType),
		items:    newGroup("{", "}", ", "),
	}
	if len(elements) > 0 {
		a.items.append(elements...)
	}
	return a
}

func (a *iarray) render(w io.Writer) {
	writeStringF(w, "[%d]", a.size)
	a.elemType.render(w)

	if a.multiLine && a.items.length() > 0 {
		// Multi-line format
		writeString(w, "{\n")
		first := true
		for _, item := range a.items.items {
			if !first {
				writeString(w, ",\n")
			}
			item.render(w)
			first = false
		}
		writeString(w, ",\n}")
	} else {
		// Single-line format
		a.items.render(w)
	}
}

func (a *iarray) String() string {
	buf := pool.Get()
	defer buf.Free()

	a.render(buf)
	return buf.String()
}

// AddElement adds elements to the array literal.
func (a *iarray) AddElement(elements ...any) *iarray {
	if len(elements) > 0 {
		a.items.append(elements...)
	}
	return a
}

// MultiLine sets the array to use multi-line format where each element
// is on its own line.
func (a *iarray) MultiLine() *iarray {
	a.multiLine = true
	return a
}
