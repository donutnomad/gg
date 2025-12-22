package main

import (
	"fmt"

	. "github.com/donutnomad/gg"
)

func main() {
	gen := New()
	gen.SetPackage("example")

	types := gen.P("github.com/example/types")

	// 示例 1: 单行格式（默认）
	gen.Body().AddLineComment("单行格式（默认）")
	gen.Body().NewVar().AddField("numbers1", Slice("int",
		Lit(1), Lit(2), Lit(3), Lit(4), Lit(5),
	))

	gen.Body().AddLine()

	// 示例 2: 多行格式
	gen.Body().AddLineComment("多行格式 - 每个元素独占一行")
	gen.Body().NewVar().AddField("numbers2", Slice("int",
		Lit(1), Lit(2), Lit(3), Lit(4), Lit(5),
	).MultiLine())

	gen.Body().AddLine()

	// 示例 3: 自定义类型的多行格式
	gen.Body().AddLineComment("自定义类型别名的多行格式")
	gen.Body().AddLineComment("假设已定义: type UserList []User")
	gen.Body().NewVar().AddField("users", Value("UserList").
		AddElement(
			S("User{ID: 1, Name: \"Alice\"}"),
			S("User{ID: 2, Name: \"Bob\"}"),
			S("User{ID: 3, Name: \"Charlie\"}"),
			S("User{ID: 4, Name: \"David\"}"),
		).MultiLine(),
	)

	gen.Body().AddLine()

	// 示例 4: 外部包类型的多行格式
	gen.Body().AddLineComment("外部包类型的多行格式")
	gen.Body().NewVar().AddField("configs", Slice(types.Ptr("Config"),
		S("&types.Config{Name: \"dev\", Port: 8080}"),
		S("&types.Config{Name: \"staging\", Port: 8081}"),
		S("&types.Config{Name: \"prod\", Port: 8082}"),
	).MultiLine())

	gen.Body().AddLine()

	// 示例 5: 数组的多行格式
	gen.Body().AddLineComment("固定大小数组的多行格式")
	gen.Body().NewVar().AddField("matrix", Array(3, "[]int",
		Slice("int", Lit(1), Lit(2), Lit(3)),
		Slice("int", Lit(4), Lit(5), Lit(6)),
		Slice("int", Lit(7), Lit(8), Lit(9)),
	).MultiLine())

	fmt.Println(gen.String())
}
