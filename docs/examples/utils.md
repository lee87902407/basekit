# utils 示例文档

## 模块用途

`utils` 提供一组高频使用的通用工具函数，涵盖字符串大小写转换、随机字符串生成、零拷贝转换和字节切片运算等场景。这些函数设计轻量、无外部依赖，适合在性能敏感路径中使用。

## 典型使用场景

1. 数据库字段名、配置键名的大小写规范化。
2. 生成临时标识符、追踪 ID（非密码学安全场景）。
3. 高频字符串/字节切片互转，避免不必要的内存分配。
4. 协议序列号、版本号等大端序字节切片自增。

## 最小可运行示例

```go
package main

import (
	"fmt"

	"github.com/lee87902407/basekit/utils"
)

func main() {
	// ASCII 字段大小写转换
	upper := utils.UpperASCIIFieldString([]byte("user_name.field"))
	lower := utils.LowerASCIIFieldString([]byte("USER_NAME.FIELD"))
	fmt.Printf("upper: %s, lower: %s\n", upper, lower)

	// 快速随机字符串生成
	randomStr := utils.FastRandomString(16)
	fmt.Printf("random: %s\n", randomStr)

	// 大端序字节切片自增
	original := []byte{0x00, 0xFF}
	incremented := utils.IncrementByteSlice(original)
	fmt.Printf("original: %x, incremented: %x\n", original, incremented)

	// 零拷贝字符串转字节切片
	str := "hello"
	bytes := utils.UnsafeStringToBytes(str)
	fmt.Printf("bytes: %v\n", bytes)
}
```

## 工具函数说明

### UpperASCIIFieldString / LowerASCIIFieldString

将 ASCII 字段字符串转换为大写或小写形式，保留下划线 `_` 和点号 `.` 不变。

**约束条件：**

- 仅接受字符集 `[A-Za-z_.]`。
- 遇到非法字符（包括数字、空格、中文等）时，直接返回空字符串。

**示例：**

```go
utils.UpperASCIIFieldString([]byte("user.name"))   // 返回 "USER.NAME"
utils.LowerASCIIFieldString([]byte("USER_NAME"))   // 返回 "user_name"
utils.UpperASCIIFieldString([]byte("user123"))     // 返回 ""（数字非法）
utils.LowerASCIIFieldString([]byte("user-name"))   // 返回 ""（连字符非法）
```

### FastRandomString

生成指定长度的随机字符串，字符集为 `[0-9a-zA-Z]`。

**安全警告：**

- 此函数使用 `math/rand`，**非密码学安全**。
- **禁止**用于生成密码、令牌、密钥等安全敏感的随机字符串。
- 仅适用于临时标识符、追踪 ID 等非安全敏感场景。

**示例：**

```go
id := utils.FastRandomString(16)  // 例如 "aB3dE7fG9hJ1kL2m"
```

### UnsafeStringToBytes / BytesToString

实现字符串与字节切片之间的零拷贝互转。

**安全警告：**

- 返回值与输入共享底层内存。
- 对于 `UnsafeStringToBytes`，返回的 `[]byte` **禁止被修改**，否则会破坏原始字符串的不可变性语义。
- 对于 `BytesToString`，在使用返回的 `string` 期间，原始 `[]byte` **禁止被修改**。
- 仅在确认不会修改数据时使用。

**示例：**

```go
// 字符串转字节切片（零拷贝）
s := "hello"
b := utils.UnsafeStringToBytes(s)  // b 与 s 共享内存，禁止修改 b

// 字节切片转字符串（零拷贝）
data := []byte{'w', 'o', 'r', 'l', 'd'}
str := utils.BytesToString(data)  // str 与 data 共享内存，使用 str 期间禁止修改 data
```

### IncrementByteSlice

将字节切片作为大端序无符号整数进行自增操作。

**行为说明：**

- 按大端序解释：最高字节在前（索引 0），最低字节在后（最后一个索引）。
- 自增从最低字节（最后一个字节）开始，如有进位则向高位传播。
- 如果输入为空切片，返回新的空切片。
- 如果发生溢出（如 `0xFF...FF`），返回相同长度的全零切片。
- 源切片不会被修改，返回一个新分配的切片。

**示例：**

```go
utils.IncrementByteSlice([]byte{0x00, 0x00})  // 返回 []byte{0x00, 0x01}
utils.IncrementByteSlice([]byte{0x00, 0xFF})  // 返回 []byte{0x01, 0x00}
utils.IncrementByteSlice([]byte{0xFF, 0xFF})  // 返回 []byte{0x00, 0x00}（溢出）
utils.IncrementByteSlice([]byte{})            // 返回 []byte{}（空输入）
```

## 接入注意事项

1. **ASCII 字段转换**：仅适用于纯 ASCII 场景，不支持 Unicode 字符。遇到非法字符时返回空字符串，调用者需自行处理。
2. **随机字符串**：`FastRandomString` 不保证唯一性，不保证均匀分布，仅用于非安全敏感场景。
3. **零拷贝转换**：使用 `UnsafeStringToBytes` 和 `BytesToString` 时，必须确保不会违反 Go 的内存安全语义。
4. **大端序自增**：`IncrementByteSlice` 每次调用都会分配新切片，高频调用场景需评估 GC 压力。

## 与 README 的跳转关系

1. 统一入口位于 [`README.md`](../../README.md)。
2. 中文说明位于 [`README.zh-CN.md`](../../README.zh-CN.md)。
3. 英文说明位于 [`README.en.md`](../../README.en.md)。
4. 对应示例代码位于 [`examples/utils/`](../../examples/utils/)。
