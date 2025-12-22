package main

import (
	"fmt"

	. "github.com/donutnomad/gg"
)

func main() {
	gen := New()
	gen.SetPackage("example")
	gen.SetHeader("Code generated. DO NOT EDIT.")

	types := gen.P("github.com/example/types")

	// 示例 1: 匿名切片类型（使用 Slice）
	gen.Body().AddLineComment("匿名切片类型 - 使用 Slice()")
	gen.Body().NewVar().AddField("users1", Slice("User",
		S("User{ID: 1, Name: \"Alice\"}"),
		S("User{ID: 2, Name: \"Bob\"}"),
	))

	gen.Body().AddLine()

	// 示例 2: 自定义类型别名（使用 Value + AddElement）
	gen.Body().AddLineComment("自定义类型别名 - 使用 Value().AddElement()")
	gen.Body().AddLineComment("假设已定义: type UserList []User")
	gen.Body().NewVar().AddField("users2",
		Value("UserList").AddElement(
			S("User{ID: 1, Name: \"Alice\"}"),
			S("User{ID: 2, Name: \"Bob\"}"),
		),
	)

	gen.Body().AddLine()

	// 示例 3: 外部包的自定义类型
	gen.Body().AddLineComment("外部包的自定义类型")
	gen.Body().AddLineComment("假设已定义: type types.UserList []types.User")
	gen.Body().NewVar().AddField("users3",
		Value(types.Type("UserList")).AddElement(
			S("types.User{ID: 1}"),
			S("types.User{ID: 2}"),
		),
	)

	gen.Body().AddLine()

	// 示例 4: 结构体字面量（使用 Value + AddField）
	gen.Body().AddLineComment("结构体字面量 - 使用 Value().AddField()")
	gen.Body().NewVar().AddField("config",
		Value("Config").
			AddField("Host", Lit("localhost")).
			AddField("Port", "8080"),
	)

	gen.Body().AddLine()

	// 示例 5: Map 类型别名
	gen.Body().AddLineComment("Map 类型别名")
	gen.Body().AddLineComment("假设已定义: type StringMap map[string]string")
	gen.Body().NewVar().AddField("settings",
		Value("StringMap").
			AddField(Lit("timeout"), Lit("30s")).
			AddField(Lit("retry"), Lit("3")),
	)

	fmt.Println(gen.String())
}
