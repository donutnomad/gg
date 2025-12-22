package gg

import (
	"strings"
	"testing"
)

func TestNewInlineGroup_SingleLine(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	fmt := gen.P("fmt")

	// 使用 NewInlineGroup 创建单行代码
	gen.Body().Append(
		NewInlineGroup().Append(
			S("result := "),
			fmt.Call("Sprintf", Lit("hello %s"), Lit("world")),
		),
	)

	output := gen.String()

	// 应该是单行，不应该有多余的换行
	if !strings.Contains(output, "result := fmt.Sprintf") {
		t.Errorf("Expected single line output with 'result := fmt.Sprintf', got:\n%s", output)
	}

	// 验证包含了参数
	if !strings.Contains(output, `"hello %s"`) || !strings.Contains(output, `"world"`) {
		t.Errorf("Expected parameters in output, got:\n%s", output)
	}

	// 不应该有 "result := \nfmt" 这样的换行
	if strings.Contains(output, "result := \nfmt") {
		t.Errorf("Found unwanted newline in output:\n%s", output)
	}
}

func TestNewInlineGroup_DynamicPackageRef(t *testing.T) {
	genA := New()
	genA.SetPackage("example")

	pkgA := genA.P("github.com/google/pkg")
	genA.Body().Append(
		NewInlineGroup().Append(
			S("x := "),
			pkgA.Call("Foo", Lit("a")),
		),
	)

	genB := New()
	genB.SetPackage("example")

	pkgB := genB.P("github.com/facebook/pkg")
	genB.Body().Append(
		NewInlineGroup().Append(
			S("y := "),
			pkgB.Call("Bar", Lit("b")),
		),
	)

	// Merge
	genA.Merge(genB)

	output := genA.String()

	// 验证包名正确重命名
	if !strings.Contains(output, "x := pkg.Foo") {
		t.Errorf("Expected 'x := pkg.Foo', got:\n%s", output)
	}

	if !strings.Contains(output, "y := pkg2.Bar") {
		t.Errorf("Expected 'y := pkg2.Bar' (dynamic rename), got:\n%s", output)
	}

	// 验证是单行
	if strings.Contains(output, "x := \npkg") || strings.Contains(output, "y := \npkg") {
		t.Errorf("Found unwanted newlines in output:\n%s", output)
	}
}

func TestNewInlineGroup_MultipleElements(t *testing.T) {
	gen := New()
	gen.SetPackage("main")

	types := gen.P("github.com/example/types")

	// 测试多个元素的组合
	gen.Body().Append(
		NewInlineGroup().Append(
			S("user := "),
			types.Type("User"),
			S("{ID: "),
			S("1"),
			S(", Name: "),
			Lit("Alice"),
			S("}"),
		),
	)

	output := gen.String()

	expected := `user := types.User{ID: 1, Name: "Alice"}`
	if !strings.Contains(output, expected) {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, output)
	}
}
