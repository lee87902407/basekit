# Utils Module Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增 `utils` 主模块，引入字符串与字节切片辅助能力，完成 API 整理、测试、示例与 README 入口更新。

**Architecture:** `utils` 模块按职责拆分为 `string.go` 与 `bytes.go`。字符串文件负责 ASCII 字段大小写转换、快速随机字符串与字符串/字节切片零拷贝转换；字节切片文件负责大端字节数组自增。实现遵循 TDD：先补失败测试，再写最小实现，最后补齐 README 与示例文档。

**Tech Stack:** Go 1.25、标准库 `math/rand/v2` 或 `math/rand`、`unsafe`、Go testing

---

### Task 1: 建立 `utils` 字符串 API 与测试骨架

**Files:**
- Create: `utils/string.go`
- Create: `utils/string_test.go`

- [ ] **Step 1: Write the failing test**

```go
package utils

import (
	"strings"
	"testing"
)

func TestUpperASCIIFieldString(t *testing.T) {
	got := UpperASCIIFieldString([]byte("Abc_def.xyz"))
	if got != "ABC_DEF.XYZ" {
		t.Fatalf("UpperASCIIFieldString() = %q, want %q", got, "ABC_DEF.XYZ")
	}
}

func TestLowerASCIIFieldString(t *testing.T) {
	got := LowerASCIIFieldString([]byte("AbC_DeF.XyZ"))
	if got != "abc_def.xyz" {
		t.Fatalf("LowerASCIIFieldString() = %q, want %q", got, "abc_def.xyz")
	}
}

func TestASCIIFieldStringRejectsInvalidChar(t *testing.T) {
	if got := UpperASCIIFieldString([]byte("abc-xyz")); got != "" {
		t.Fatalf("UpperASCIIFieldString() = %q, want empty string", got)
	}
	if got := LowerASCIIFieldString([]byte("abc/xyz")); got != "" {
		t.Fatalf("LowerASCIIFieldString() = %q, want empty string", got)
	}
}

func TestUpperASCIIFieldStringSupportsLongInput(t *testing.T) {
	src := strings.Repeat("ab.", 40)
	got := UpperASCIIFieldString([]byte(src))
	want := strings.ToUpper(src)
	if got != want {
		t.Fatalf("UpperASCIIFieldString() len=%d mismatch", len(src))
	}
}

func TestUpperASCIIFieldByte(t *testing.T) {
	if got := UpperASCIIFieldByte('a'); got != 'A' {
		t.Fatalf("UpperASCIIFieldByte('a') = %q, want %q", got, 'A')
	}
	if got := UpperASCIIFieldByte('_'); got != 0 {
		t.Fatalf("UpperASCIIFieldByte('_') = %q, want %q", got, byte(0))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./utils -run 'TestUpperASCIIFieldString|TestLowerASCIIFieldString|TestASCIIFieldStringRejectsInvalidChar|TestUpperASCIIFieldStringSupportsLongInput|TestUpperASCIIFieldByte'`
Expected: FAIL with undefined `UpperASCIIFieldString` / `LowerASCIIFieldString` / `UpperASCIIFieldByte`

- [ ] **Step 3: Write minimal implementation**

```go
package utils

var asciiRandomBytes = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var upperCharmap [256]byte
var lowerCharmap [256]byte

func init() {
	for i := range upperCharmap {
		c := byte(i)
		switch {
		case c >= 'A' && c <= 'Z':
			upperCharmap[i] = c
		case c >= 'a' && c <= 'z':
			upperCharmap[i] = c - 'a' + 'A'
		}
	}

	for i := range lowerCharmap {
		c := byte(i)
		switch {
		case c >= 'A' && c <= 'Z':
			lowerCharmap[i] = c + 32
		case c >= 'a' && c <= 'z':
			lowerCharmap[i] = c
		}
	}
}

func UpperASCIIFieldString(op []byte) string {
	result := make([]byte, len(op))
	for i := range op {
		if op[i] == '_' || op[i] == '.' {
			result[i] = op[i]
			continue
		}
		c := upperCharmap[op[i]]
		if c < 'A' || c > 'Z' {
			return ""
		}
		result[i] = c
	}
	return BytesToString(result)
}

func LowerASCIIFieldString(op []byte) string {
	result := make([]byte, len(op))
	for i := range op {
		if op[i] == '_' || op[i] == '.' {
			result[i] = op[i]
			continue
		}
		c := lowerCharmap[op[i]]
		if c < 'a' || c > 'z' {
			return ""
		}
		result[i] = c
	}
	return BytesToString(result)
}

func UpperASCIIFieldByte(b byte) byte {
	return upperCharmap[b]
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./utils -run 'TestUpperASCIIFieldString|TestLowerASCIIFieldString|TestASCIIFieldStringRejectsInvalidChar|TestUpperASCIIFieldStringSupportsLongInput|TestUpperASCIIFieldByte'`
Expected: PASS

### Task 2: 补齐随机字符串与 unsafe 转换能力

**Files:**
- Modify: `utils/string.go`
- Modify: `utils/string_test.go`

- [ ] **Step 1: Write the failing test**

```go
package utils

import "testing"

func TestFastRandomString(t *testing.T) {
	got := FastRandomString(32)
	if len(got) != 32 {
		t.Fatalf("len(FastRandomString(32)) = %d, want 32", len(got))
	}
	for i := range got {
		if indexByte(asciiRandomBytes, got[i]) == -1 {
			t.Fatalf("FastRandomString() contains invalid char %q", got[i])
		}
	}
}

func TestFastRandomStringReturnsEmptyForNonPositiveLength(t *testing.T) {
	if got := FastRandomString(0); got != "" {
		t.Fatalf("FastRandomString(0) = %q, want empty string", got)
	}
}

func TestUnsafeStringToBytesAndBytesToString(t *testing.T) {
	src := "hello"
	b := UnsafeStringToBytes(src)
	if len(b) != len(src) {
		t.Fatalf("len(UnsafeStringToBytes()) = %d, want %d", len(b), len(src))
	}
	if got := BytesToString([]byte("world")); got != "world" {
		t.Fatalf("BytesToString() = %q, want %q", got, "world")
	}
	if string(b) != src {
		t.Fatalf("string(UnsafeStringToBytes()) = %q, want %q", string(b), src)
	}
}

func indexByte(src []byte, target byte) int {
	for i := range src {
		if src[i] == target {
			return i
		}
	}
	return -1
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./utils -run 'TestFastRandomString|TestUnsafeStringToBytesAndBytesToString'`
Expected: FAIL with undefined `FastRandomString` / `UnsafeStringToBytes` / `BytesToString`

- [ ] **Step 3: Write minimal implementation**

```go
package utils

import (
	"math/rand"
	"sync"
	"unsafe"
)

var randomMu sync.Mutex
var randomSource = rand.New(rand.NewSource(1))

func FastRandomString(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	randomMu.Lock()
	for i := range b {
		b[i] = asciiRandomBytes[randomSource.Intn(len(asciiRandomBytes))]
	}
	randomMu.Unlock()
	return BytesToString(b)
}

// UnsafeStringToBytes 零拷贝返回字符串底层字节视图，调用方不得写入返回切片。
func UnsafeStringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToString 零拷贝返回字节切片对应字符串，后续修改底层切片会影响结果语义。
func BytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./utils -run 'TestFastRandomString|TestUnsafeStringToBytesAndBytesToString'`
Expected: PASS

### Task 3: 拆分并实现字节切片自增能力

**Files:**
- Create: `utils/bytes.go`
- Create: `utils/bytes_test.go`

- [ ] **Step 1: Write the failing test**

```go
package utils

import (
	"bytes"
	"testing"
)

func TestIncrementByteSlice(t *testing.T) {
	src := []byte{0x01, 0x02, 0x03}
	got := IncrementByteSlice(src)
	want := []byte{0x01, 0x02, 0x04}
	if !bytes.Equal(got, want) {
		t.Fatalf("IncrementByteSlice() = %v, want %v", got, want)
	}
	if !bytes.Equal(src, []byte{0x01, 0x02, 0x03}) {
		t.Fatalf("IncrementByteSlice() modified source: %v", src)
	}
}

func TestIncrementByteSliceCarry(t *testing.T) {
	got := IncrementByteSlice([]byte{0x00, 0x00, 0xff})
	want := []byte{0x00, 0x01, 0x00}
	if !bytes.Equal(got, want) {
		t.Fatalf("IncrementByteSlice() = %v, want %v", got, want)
	}
}

func TestIncrementByteSliceOverflow(t *testing.T) {
	got := IncrementByteSlice([]byte{0xff, 0xff})
	want := []byte{0x00, 0x00}
	if !bytes.Equal(got, want) {
		t.Fatalf("IncrementByteSlice() = %v, want %v", got, want)
	}
}

func TestIncrementByteSliceEmpty(t *testing.T) {
	got := IncrementByteSlice(nil)
	if len(got) != 0 {
		t.Fatalf("len(IncrementByteSlice(nil)) = %d, want 0", len(got))
	}
	if got == nil {
		t.Fatal("IncrementByteSlice(nil) should return new empty slice")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./utils -run 'TestIncrementByteSlice'`
Expected: FAIL with undefined `IncrementByteSlice`

- [ ] **Step 3: Write minimal implementation**

```go
package utils

func IncrementByteSlice(src []byte) []byte {
	if len(src) == 0 {
		return []byte{}
	}

	dst := make([]byte, len(src))
	copy(dst, src)

	for i := len(dst) - 1; i >= 0; i-- {
		dst[i]++
		if dst[i] != 0 {
			return dst
		}
	}

	return dst
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./utils -run 'TestIncrementByteSlice'`
Expected: PASS

### Task 4: 补齐模块文档、示例与 README 入口

**Files:**
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `README.en.md`
- Modify: `docs/examples/README.md`
- Modify: `examples/README.md`
- Create: `docs/examples/utils.md`
- Create: `examples/utils/main.go`

- [ ] **Step 1: Write the failing documentation expectation**

```text
需要新增以下可见入口，否则视为未完成：
1. README 三份文件中都出现 utils 模块简介与跳转。
2. docs/examples/README.md 中出现 utils 示例文档入口。
3. examples/README.md 中出现 examples/utils/ 入口。
4. docs/examples/utils.md 说明 FastRandomString 非加密、UnsafeStringToBytes 不安全、ASCII 字段转换字符集限制、IncrementByteSlice 语义。
5. examples/utils/main.go 给出最小可运行示例。
```

- [ ] **Step 2: Verify the expectation currently fails**

Run: `rg -n "utils" README.md README.zh-CN.md README.en.md docs/examples/README.md examples/README.md`
Expected: `README*` 与示例入口文件中尚未完整出现 `utils` 模块说明

- [ ] **Step 3: Write minimal documentation and example implementation**

```markdown
# docs/examples/utils.md

# utils 模块示例

## 模块用途

`utils` 提供一组轻量字符串与字节切片辅助能力，当前包含：

- 面向特定字段的 ASCII 大小写转换
- 非加密快速随机字符串生成
- 字符串与字节切片的零拷贝转换
- 大端字节切片加一

## 适用场景

1. 需要对受限字段名做快速 ASCII 归一化。
2. 需要在非安全场景快速生成随机文本。
3. 需要显式接受 `unsafe` 风险来减少字符串与字节切片转换开销。
4. 需要对字节切片形式的游标或键做顺序递增。

## 最小示例

```go
package main

import (
	"fmt"

	"github.com/lee87902407/basekit/utils"
)

func main() {
	fmt.Println(utils.UpperASCIIFieldString([]byte("event.name")))
	fmt.Println(utils.FastRandomString(8))
	fmt.Println(utils.IncrementByteSlice([]byte{0x00, 0x00, 0xff}))
}
```

## 注意事项

1. `UpperASCIIFieldString` / `LowerASCIIFieldString` 仅接受 `[A-Za-z_.]`，非法字符返回空字符串。
2. `FastRandomString` 仅用于非加密用途。
3. `UnsafeStringToBytes` 返回的切片不得写入。
4. `BytesToString` 与原切片共享底层内存语义，不适合作为长期不可变快照。

## 相关入口

- 统一入口：[`README.md`](../../README.md)
- 中文说明：[`README.zh-CN.md`](../../README.zh-CN.md)
```

```go
package main

import (
	"fmt"

	"github.com/lee87902407/basekit/utils"
)

func main() {
	fmt.Println("upper:", utils.UpperASCIIFieldString([]byte("event.name")))
	fmt.Println("lower:", utils.LowerASCIIFieldString([]byte("EVENT.NAME")))
	fmt.Println("random:", utils.FastRandomString(8))
	fmt.Println("next:", utils.IncrementByteSlice([]byte{0x00, 0x00, 0xff}))
}
```

```markdown
# README.md 增补内容

- `utils`：以字符串处理为主的轻量工具模块，提供特定字段 ASCII 大小写转换、非加密快速随机字符串、零拷贝字符串/字节切片转换与大端字节切片自增能力。
- `utils` 示例文档：[`docs/examples/utils.md`](./docs/examples/utils.md)
- `utils` 示例代码：[`examples/utils/`](./examples/utils/)
```

- [ ] **Step 4: Verify docs and example are discoverable**

Run: `go test ./... && rg -n "utils" README.md README.zh-CN.md README.en.md docs/examples/README.md examples/README.md docs/examples/utils.md`
Expected: PASS, and all entry files contain `utils` links

### Task 5: 全量验证与收尾检查

**Files:**
- Modify: `utils/string.go`
- Modify: `utils/string_test.go`
- Modify: `utils/bytes.go`
- Modify: `utils/bytes_test.go`
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `README.en.md`
- Modify: `docs/examples/README.md`
- Modify: `examples/README.md`
- Modify: `docs/examples/utils.md`
- Modify: `examples/utils/main.go`

- [ ] **Step 1: Run LSP diagnostics on new code files**

Run: `lsp_diagnostics` for `utils/string.go`, `utils/string_test.go`, `utils/bytes.go`, `utils/bytes_test.go`
Expected: no errors

- [ ] **Step 2: Run full test suite**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 3: Manually verify README navigation**

```text
检查以下跳转关系：
1. README.md -> docs/examples/utils.md
2. README.md -> examples/utils/
3. README.zh-CN.md -> docs/examples/utils.md
4. README.en.md -> docs/examples/utils.md
5. docs/examples/README.md -> docs/examples/utils.md
```

- [ ] **Step 4: Review comments for unsafe and non-crypto warnings**

```text
确认以下文本已经出现在代码注释或文档中：
1. FastRandomString 仅用于非加密用途
2. UnsafeStringToBytes 返回切片不得写入
3. BytesToString 共享底层内存语义
4. Upper/Lower ASCII Field String 仅接受 [A-Za-z_.]
```
