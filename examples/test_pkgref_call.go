package main

import (
	"fmt"
	"strings"

	. "github.com/donutnomad/gg"
)

func main() {
	// Generator A
	genA := New()
	genA.SetPackage("example")

	gsqlPkg := genA.P("github.com/google/gsql")

	// ❌ 错误方式：使用 S() + Call()，会立即固化为字符串
	// tnDecl := S("tn := %s", gsqlPkg.Call("TableName", Lit("tableName")))

	// ✅ 正确方式1：使用 NewInlineGroup() 组合 Node
	genA.Body().Append(
		NewInlineGroup().Append(
			S("tn := "),
			gsqlPkg.Call("TableName", Lit("tableName")),
		),
	)

	fmt.Println("=== Generator A ===")
	fmt.Println(genA.String())
	fmt.Println()

	// Generator B - 使用同名但不同路径的包
	genB := New()
	genB.SetPackage("example")

	gsqlPkg2 := genB.P("github.com/facebook/gsql")
	genB.Body().Append(
		NewInlineGroup().Append(
			S("tn2 := "),
			gsqlPkg2.Call("TableName", Lit("tableName2")),
		),
	)

	fmt.Println("=== Generator B ===")
	fmt.Println(genB.String())
	fmt.Println()

	// Merge A + B
	fmt.Println("=== Before Merge ===")
	fmt.Printf("genB gsqlPkg2 alias: %s\n", gsqlPkg2.Alias())

	genA.Merge(genB)

	fmt.Println("\n=== After Merge ===")
	fmt.Printf("genB gsqlPkg2 alias: %s\n", gsqlPkg2.Alias())

	fmt.Println("=== Merged (A + B) ===")
	fmt.Println(genA.String())
	fmt.Println()

	// 检查是否正确重命名
	output := genA.String()
	if !strings.Contains(output, "gsql.TableName") {
		fmt.Println("❌ 错误：找不到 gsql.TableName")
	} else {
		fmt.Println("✅ 正确：找到 gsql.TableName")
	}

	if !strings.Contains(output, "gsql2.TableName") {
		fmt.Println("❌ 错误：找不到 gsql2.TableName（说明包名被固化了）")
	} else {
		fmt.Println("✅ 正确：找到 gsql2.TableName（包名动态更新）")
	}
}
