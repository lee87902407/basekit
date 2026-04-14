# utils 模块设计说明

## 1. 背景

`basekit` 当前已包含 `mempool` 与 `log` 两个主模块，但缺少一组可在多个项目中复用的轻量字符串与字节切片辅助能力。现有来源文件 `/Users/yanjie/data/code/work/storage/flight-storage-tools/utils/string.go` 提供了若干可复用基础能力，但其实现依赖原项目上下文，直接照搬到 `basekit` 会带入命名不清晰、长度限制隐式存在、字符串与字节能力混杂等问题。

本次新增 `utils` 主模块，以字符串相关能力为主，并保留少量与字符串紧邻的字节切片辅助能力。

## 2. 目标

本次设计目标如下：

1. 在 `basekit` 中新增正式主模块 `utils`。
2. 将来源文件中的可复用能力整理后落到 `utils/string.go`。
3. 将非字符串主能力的字节切片自增逻辑拆分到 `utils/bytes.go`。
4. 在保留原有业务语义的前提下，修复公共模块中不适合保留的实现细节。
5. 按仓库规范同步补齐 README、示例文档、示例代码与测试。

## 3. 非目标

本次不包含以下内容：

1. 不把 `utils` 扩展成无边界的杂项工具集合。
2. 不引入 Unicode 通用大小写转换能力，本次仅处理 ASCII 范围内的特定字段转换。
3. 不提供密码学安全随机字符串能力。
4. 不对 `unsafe` 转换封装为完全安全 API，而是通过命名和文档显式暴露风险。

## 4. 模块定位与范围

`utils` 作为新的主模块，定位为：

- 以字符串处理为主；
- 允许保留少量紧邻字符串场景的字节切片辅助能力；
- 模块首批能力必须边界清晰、语义明确、文档完整。

首批文件布局：

- `utils/string.go`
- `utils/bytes.go`
- `utils/string_test.go`
- `utils/bytes_test.go`

## 5. API 设计

### 5.1 utils/string.go

计划提供以下 API：

1. `FastRandomString(n int) string`
2. `UnsafeStringToBytes(s string) []byte`
3. `BytesToString(b []byte) string`
4. `UpperASCIIFieldString(op []byte) string`
5. `LowerASCIIFieldString(op []byte) string`
6. `UpperASCIIFieldByte(b byte) byte`

#### 5.1.1 FastRandomString

语义：

- 生成长度为 `n` 的随机字符串；
- 字符集沿用来源实现：`0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`；
- 该函数只面向非加密场景，优先考虑高频调用下的速度，不承诺密码学随机质量。

约束：

- `n <= 0` 时返回空字符串；
- 文档与示例中必须明确标注“不可用于安全令牌、密钥、验证码等敏感用途”。

实现要求：

- 不再沿用每次调用都新建 `rand.New(rand.NewSource(time.Now().UnixNano()))` 的方式；
- 改为包级复用随机源或等效的轻量实现，减少频繁创建开销；
- 若使用共享随机源，必须保证并发场景下行为正确。

#### 5.1.2 UnsafeStringToBytes

语义：

- 零拷贝地将 `string` 转为 `[]byte`；
- 保留来源实现的性能特点。

约束：

- 函数名必须显式包含 `Unsafe`；
- 注释中必须明确说明返回的切片引用原字符串底层数据，不得写入，否则行为未定义；
- 示例文档需要给出使用警告。

#### 5.1.3 BytesToString

语义：

- 零拷贝地将 `[]byte` 转为 `string`；
- 空切片返回空字符串。

约束：

- 注释中需说明：若调用方后续继续修改底层切片，生成的字符串语义将依赖该共享底层内存，不适合作为长期不可变快照使用。

#### 5.1.4 UpperASCIIFieldString / LowerASCIIFieldString

语义：

- 仅用于特定字段的 ASCII 大小写转换；
- 合法字符集合仅为：`A-Z`、`a-z`、`_`、`.`；
- `_` 与 `.` 原样保留；
- 遇到任意非法字符时返回空字符串。

实现要求：

- 保留预计算字符映射表思路；
- 不再使用固定 `[64]byte` 缓冲区；
- 统一改为 `make([]byte, len(op))`，消除输入长度超过 64 时的越界风险；
- 不引入额外 Unicode 处理逻辑。

命名理由：

- `UpperASCIIFieldString` / `LowerASCIIFieldString` 比原始命名更能体现“ASCII 范围 + 特定字段”的用途与限制。

#### 5.1.5 UpperASCIIFieldByte

语义：

- 对单个字节执行与字段字符串转换一致的 ASCII 大写映射；
- 返回映射表中的结果值。

约束：

- 保留轻量快速路径；
- 行为需与 `UpperASCIIFieldString` 的字符规则保持一致。

### 5.2 utils/bytes.go

计划提供以下 API：

1. `IncrementByteSlice(src []byte) []byte`

语义：

- 将字节切片视为一个大端无符号整数并执行加一；
- 返回新切片，不修改输入切片。

约束：

- 空切片返回新的空切片；
- 全 `0xff` 输入时，返回等长全 `0x00` 结果，保持与来源实现一致。

命名理由：

- `IncrementByteSlice` 比 `IncreaseOne` 更清晰表达操作对象与行为。

## 6. 关键实现调整

相对来源文件，本次会做以下明确调整：

1. **重命名 API**：提高公共模块中的可读性与可发现性。
2. **拆分职责**：将字节切片自增从字符串文件拆出到 `bytes.go`。
3. **去除固定长度限制**：把 `[64]byte` 临时缓冲区改为按输入长度分配。
4. **保留字段字符约束**：不放宽为通用字符串转换，而是继续面向特定字段场景。
5. **显式暴露 unsafe 风险**：通过命名与注释降低误用概率。
6. **优化随机字符串热路径**：减少重复创建随机源的额外成本。

## 7. 并发与性能考虑

### 7.1 随机字符串

由于 `FastRandomString` 明确定位为高频快速、非加密用途，实现应优先减少对象创建与初始化成本。设计要求如下：

- 尽量复用随机源；
- 保证并发场景下不会产生数据竞争；
- 避免为了低价值“高随机质量”引入更重实现。

### 7.2 映射表转换

ASCII 字段转换继续采用映射表，以维持简单、可预测、低分支的转换路径。输入长度动态分配的开销可接受，因为该方案消除了固定 64 长度假设带来的公共库风险。

## 8. 错误处理与边界行为

### 8.1 FastRandomString

- `n <= 0`：返回空字符串。

### 8.2 UpperASCIIFieldString / LowerASCIIFieldString

- 输入为空：返回空字符串；
- 包含非法字符：返回空字符串；
- 输入很长：按长度正常处理，不应 panic。

### 8.3 UnsafeStringToBytes / BytesToString

- 输入为空：返回空结果；
- 对共享底层内存的风险通过注释和示例明确说明，不通过运行时防护兜底。

### 8.4 IncrementByteSlice

- 空输入：返回新空切片；
- 普通进位：返回加一后的新切片；
- 全量进位：返回等长全零切片。

## 9. 测试设计

### 9.1 string 测试

覆盖以下场景：

1. `UpperASCIIFieldString` 正常转换。
2. `LowerASCIIFieldString` 正常转换。
3. `_`、`.` 原样保留。
4. 非法字符返回空字符串。
5. 长输入（超过 64）不会 panic，且结果正确。
6. `UpperASCIIFieldByte` 的单字节映射正确。
7. `FastRandomString` 返回长度正确。
8. `FastRandomString` 的字符均落在允许字符集。
9. `UnsafeStringToBytes` 与 `BytesToString` 基本内容正确。

### 9.2 bytes 测试

覆盖以下场景：

1. 空切片输入。
2. 无进位加一。
3. 局部连续进位。
4. 全 `0xff` 溢出。
5. 输入切片未被原地修改。

## 10. 文档与示例更新

根据仓库规范，本次新增 `utils` 主模块时需同步完成：

1. 更新 `README.md`。
2. 更新 `README.zh-CN.md`。
3. 更新 `README.en.md`。
4. 新增 `docs/examples/utils.md`。
5. 新增 `examples/utils/main.go`。

文档中需要明确写入：

- `utils` 模块用途与适用场景；
- `FastRandomString` 仅用于非加密用途；
- `UnsafeStringToBytes` 的不安全语义；
- ASCII 字段转换函数只接受 `[A-Za-z_.]`；
- `IncrementByteSlice` 的大端字节切片加一语义。

## 11. 验证要求

实现完成后至少需要执行：

1. `lsp_diagnostics` 检查新增文件是否有静态诊断问题；
2. `go test ./...` 验证测试通过；
3. 对 README、示例文档、示例代码跳转关系做一次人工检查，确保 `README.md` 可访问新增模块入口。

## 12. 风险与取舍

### 12.1 unsafe API 风险

保留 `unsafe` 是性能取向下的显式取舍。风险不通过封装隐藏，而通过命名、注释、文档主动暴露。

### 12.2 随机实现取舍

`FastRandomString` 不追求密码学安全，避免把高安全成本错误地引入追求速度的基础工具函数。

### 12.3 模块边界取舍

本次接受 `IncrementByteSlice` 进入 `utils`，但仅作为紧邻能力保留，不意味着后续可无限制向 `utils` 塞入任意杂项函数。后续新增能力仍需根据主题聚合与文档规范判断是否值得成为 `utils` 的组成部分。
