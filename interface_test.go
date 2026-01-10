package gg

import "testing"

func TestInterface(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		expected := `type Tester interface {
// TestA is a test
TestA(a int64, b int)

TestB() (err error)
}
`
		in := Interface("Tester")
		in.AddLineComment("TestA is a test")
		in.NewFunction("TestA").
			AddParameter("a", "int64").
			AddParameter("b", "int")
		in.AddLine()
		in.NewFunction("TestB").
			AddResult("err", "error")

		in.render(buf)

		compareAST(t, expected, buf.String())
	})

	t.Run("merge consecutive same type parameters", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		expected := `type Reader interface {
Read(a, b string, c int)
}`

		in := Interface("Reader")
		in.NewFunction("Read").
			AddParameter("a", "string").
			AddParameter("b", "string").
			AddParameter("c", "int")

		in.render(buf)

		compareAST(t, expected, buf.String())
	})

	t.Run("merge consecutive same type results", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		expected := `type Getter interface {
Get() (key, value string, err error)
}`

		in := Interface("Getter")
		in.NewFunction("Get").
			AddResult("key", "string").
			AddResult("value", "string").
			AddResult("err", "error")

		in.render(buf)

		compareAST(t, expected, buf.String())
	})

	t.Run("merge multiple groups of same type", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		expected := `type Handler interface {
Handle(a, b int, c string, d, e int) (x, y bool, err error)
}`

		in := Interface("Handler")
		in.NewFunction("Handle").
			AddParameter("a", "int").
			AddParameter("b", "int").
			AddParameter("c", "string").
			AddParameter("d", "int").
			AddParameter("e", "int").
			AddResult("x", "bool").
			AddResult("y", "bool").
			AddResult("err", "error")

		in.render(buf)

		compareAST(t, expected, buf.String())
	})
}
