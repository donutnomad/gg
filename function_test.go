package gg

import "testing"

func TestFunction(t *testing.T) {
	t.Run("no receiver", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		expected := `func Test(a int, b string) (d uint)`

		Function("Test").
			AddParameter("a", "int").
			AddParameter("b", "string").
			AddResult("d", "uint").
			render(buf)

		compareAST(t, expected, buf.String())
	})
	t.Run("has receiver", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		expected := `func (r *Q) Test() (a int, b int64, d string) {
return "Hello, World!"
}`
		Function("Test").
			WithReceiver("r", "*Q").
			AddResult("a", "int").
			AddResult("b", "int64").
			AddResult("d", "string").
			AddBody(
				String(`return "Hello, World!"`),
			).
			render(buf)

		compareAST(t, expected, buf.String())
	})

	t.Run("node input", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		expected := `func (r *Q) Test() (a int, b int64, d string) {
return "Hello, World!"
}`
		Function("Test").
			WithReceiver("r", "*Q").
			AddResult("a", String("int")).
			AddResult("b", "int64").
			AddResult("d", "string").
			AddBody(
				String(`return "Hello, World!"`),
			).
			render(buf)

		compareAST(t, expected, buf.String())
	})

	t.Run("call", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		expected := `func(){}()`

		fn := Function("")
		fn.WithCall()
		fn.render(buf)

		compareAST(t, expected, buf.String())
	})

	t.Run("no name result - no receiver - single result", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		expected := `func Test(a int) (int)`

		Function("Test").
			AddParameter("a", "int").
			AddResult("", "int").
			render(buf)

		compareAST(t, expected, buf.String())
	})

	t.Run("no name result - no receiver - multi result", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		expected := `func Test(a int) (int, string, error)`

		Function("Test").
			AddParameter("a", "int").
			AddResult("", "int").
			AddResult("", "string").
			AddResult("", "error").
			render(buf)

		compareAST(t, expected, buf.String())
	})

	t.Run("no name result - has receiver - single result", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		expected := `func (r *Q) Test(a int) (int)`

		Function("Test").
			WithReceiver("r", "*Q").
			AddParameter("a", "int").
			AddResult("", "int").
			render(buf)

		compareAST(t, expected, buf.String())
	})

	t.Run("no name result - has receiver - multi result", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		expected := `func (r *Q) Test(a int) (int, string, error)`

		Function("Test").
			WithReceiver("r", "*Q").
			AddParameter("a", "int").
			AddResult("", "int").
			AddResult("", "string").
			AddResult("", "error").
			render(buf)

		compareAST(t, expected, buf.String())
	})

	t.Run("merge consecutive same type parameters", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		// Two consecutive string parameters should be merged
		expected := `func Test(a, b string, c int)`

		Function("Test").
			AddParameter("a", "string").
			AddParameter("b", "string").
			AddParameter("c", "int").
			render(buf)

		compareAST(t, expected, buf.String())
	})

	t.Run("merge consecutive same type results", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		// Two consecutive PubKey results should be merged
		expected := `func Test() (publicKey, rawPublicKey types.PubKey, refID RefID, err error)`

		Function("Test").
			AddResult("publicKey", "types.PubKey").
			AddResult("rawPublicKey", "types.PubKey").
			AddResult("refID", "RefID").
			AddResult("err", "error").
			render(buf)

		compareAST(t, expected, buf.String())
	})

	t.Run("merge multiple groups of same type", func(t *testing.T) {
		buf := pool.Get()
		defer buf.Free()

		// Multiple groups of same type parameters
		expected := `func Test(a, b int, c string, d, e int)`

		Function("Test").
			AddParameter("a", "int").
			AddParameter("b", "int").
			AddParameter("c", "string").
			AddParameter("d", "int").
			AddParameter("e", "int").
			render(buf)

		compareAST(t, expected, buf.String())
	})
}
