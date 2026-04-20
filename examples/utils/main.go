package main

import (
	"fmt"

	"github.com/lee87902407/basekit/utils"
)

func main() {
	fmt.Println("=== utils 示例 ===")
	fmt.Println()

	// 1. UpperASCIIFieldString 示例
	fmt.Println("--- UpperASCIIFieldString ---")
	fmt.Printf("user_name.field -> %s\n", utils.UpperASCIIFieldString([]byte("user_name.field")))
	fmt.Printf("USER -> %s\n", utils.UpperASCIIFieldString([]byte("USER")))
	fmt.Printf("config.key -> %s\n", utils.UpperASCIIFieldString([]byte("config.key")))
	// 非法字符示例（包含数字）
	fmt.Printf("user123 -> %q (包含非法数字)\n", utils.UpperASCIIFieldString([]byte("user123")))
	fmt.Println()

	// 2. LowerASCIIFieldString 示例
	fmt.Println("--- LowerASCIIFieldString ---")
	fmt.Printf("USER_NAME.FIELD -> %s\n", utils.LowerASCIIFieldString([]byte("USER_NAME.FIELD")))
	fmt.Printf("user -> %s\n", utils.LowerASCIIFieldString([]byte("user")))
	fmt.Printf("CONFIG.KEY -> %s\n", utils.LowerASCIIFieldString([]byte("CONFIG.KEY")))
	// 非法字符示例（包含连字符）
	fmt.Printf("user-name -> %q (包含非法连字符)\n", utils.LowerASCIIFieldString([]byte("user-name")))
	fmt.Println()

	// 3. FastRandomString 示例
	fmt.Println("--- FastRandomString ---")
	fmt.Printf("随机字符串 (长度 16): %s\n", utils.FastRandomString(16))
	fmt.Printf("随机字符串 (长度 8):  %s\n", utils.FastRandomString(8))
	fmt.Printf("随机字符串 (长度 32): %s\n", utils.FastRandomString(32))
	fmt.Println("注意: FastRandomString 非密码学安全，仅用于非安全敏感场景")
	fmt.Println()

	// 4. IncrementByteSlice 示例
	fmt.Println("--- IncrementByteSlice ---")
	// 正常自增
	original := []byte{0x00, 0x00}
	incremented := utils.IncrementByteSlice(original)
	fmt.Printf("0x0000 -> 0x%04x\n", incremented)

	// 低字节溢出进位
	original = []byte{0x00, 0xFF}
	incremented = utils.IncrementByteSlice(original)
	fmt.Printf("0x00FF -> 0x%04x\n", incremented)

	// 全溢出返回全零
	original = []byte{0xFF, 0xFF}
	incremented = utils.IncrementByteSlice(original)
	fmt.Printf("0xFFFF -> 0x%04x (溢出)\n", incremented)

	// 源切片不被修改
	original = []byte{0x12, 0x34}
	incremented = utils.IncrementByteSlice(original)
	fmt.Printf("源切片: 0x%04x, 结果: 0x%04x (源切片不被修改)\n", original, incremented)
	fmt.Println()

	// 5. UnsafeStringToBytes 示例
	fmt.Println("--- UnsafeStringToBytes ---")
	str := "hello world"
	bytes := utils.UnsafeStringToBytes(str)
	fmt.Printf("字符串: %q\n", str)
	fmt.Printf("转换后: %v\n", bytes)
	fmt.Println("警告: 返回的 []byte 与原字符串共享内存，禁止修改")
	fmt.Println()

	// 6. UnsafeBytesToString 示例
	fmt.Println("--- UnsafeBytesToString ---")
	data := []byte{'G', 'o', 'l', 'a', 'n', 'g'}
	result := utils.UnsafeBytesToString(data)
	fmt.Printf("字节切片: %v\n", data)
	fmt.Printf("转换后: %q\n", result)
	fmt.Println("警告: 返回的 string 与原 []byte 共享内存，使用期间禁止修改原切片")
}
