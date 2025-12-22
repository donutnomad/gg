package main

import (
	"fmt"

	. "github.com/donutnomad/gg"
)

func main() {
	// 生成器 a: 使用 github.com/google/a/book
	genA := New()
	genA.SetPackage("example")

	bookGoogle := genA.P("github.com/google/a/book")

	genA.Body().NewFunction("ProcessGoogleBook").
		AddParameter("b", bookGoogle.Type("Book")).
		AddBody(
			S("// Process Google book"),
			S("_ = b"),
		)

	fmt.Println("=== Generator A ===")
	fmt.Println(genA.String())
	fmt.Println()

	// 生成器 b: 使用 github.com/facebook/book
	genB := New()
	genB.SetPackage("example")

	bookFacebook := genB.P("github.com/facebook/book")

	genB.Body().NewFunction("ProcessFacebookBook").
		AddParameter("b", bookFacebook.Type("Book")).
		AddBody(
			S("// Process Facebook book"),
			S("_ = b"),
		)

	fmt.Println("=== Generator B ===")
	fmt.Println(genB.String())
	fmt.Println()

	// 合并 genB 到 genA
	genA.Merge(genB)

	fmt.Println("=== Merged Generator (A + B) ===")
	fmt.Println(genA.String())
	fmt.Println()
}
