package utils

import (
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

// asciiRandomBytes 包含生成随机字符串时可用的字母表字节
var asciiRandomBytes = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var upperCharmap [256]byte
var lowerCharmap [256]byte

// randomMu 保护 randomSource 的并发访问
var randomMu sync.Mutex

// randomSource 是共享的随机数生成器
var randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))

func init() {
	// 初始化 upperCharmap：所有字节先映射到自身
	for i := 0; i < 256; i++ {
		upperCharmap[i] = byte(i)
	}
	// 字母 a-z 转换为大写
	for i := 'a'; i <= 'z'; i++ {
		upperCharmap[i] = byte(i - 32)
	}

	// 初始化 lowerCharmap：所有字节先映射到自身
	for i := 0; i < 256; i++ {
		lowerCharmap[i] = byte(i)
	}
	// 字母 A-Z 转换为小写
	for i := 'A'; i <= 'Z'; i++ {
		lowerCharmap[i] = byte(i + 32)
	}
}

// validASCIIFieldChar 检查字节是否为合法的 ASCII 字段字符
// 仅接受字母、下划线和点号，数字被视为非法字符
func validASCIIFieldChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || b == '.'
}

// UpperASCIIFieldString 将 ASCII 字段字符串转换为大写形式
// 保留 '_' 和 '.' 不变，遇到非法字符（包括数字）时返回空字符串
func UpperASCIIFieldString(op []byte) string {
	result := make([]byte, len(op))
	for i := range op {
		b := op[i]
		if !validASCIIFieldChar(b) {
			return ""
		}
		// 字母使用映射表转换，非字母（'_' 和 '.'）保持原值
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') {
			result[i] = upperCharmap[b]
		} else {
			result[i] = b
		}
	}
	return UnsafeBytesToString(result)
}

// LowerASCIIFieldString 将 ASCII 字段字符串转换为小写形式
// 保留 '_' 和 '.' 不变，遇到非法字符（包括数字）时返回空字符串
func LowerASCIIFieldString(op []byte) string {
	result := make([]byte, len(op))
	for i := range op {
		b := op[i]
		if !validASCIIFieldChar(b) {
			return ""
		}
		// 字母使用映射表转换，非字母（'_' 和 '.'）保持原值
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') {
			result[i] = lowerCharmap[b]
		} else {
			result[i] = b
		}
	}
	return UnsafeBytesToString(result)
}

// UpperASCIIFieldByte 返回单个字节的大写映射
func UpperASCIIFieldByte(b byte) byte {
	return upperCharmap[b]
}

// FastRandomString 生成指定长度的随机字符串
// 注意：此函数仅用于非密码学用途，不应用于生成安全敏感的随机字符串
func FastRandomString(n int) string {
	if n <= 0 {
		return ""
	}
	result := make([]byte, n)
	randomMu.Lock()
	for i := 0; i < n; i++ {
		result[i] = asciiRandomBytes[randomSource.Intn(len(asciiRandomBytes))]
	}
	randomMu.Unlock()
	return UnsafeBytesToString(result)
}

// UnsafeStringToBytes 高效地将 string 转换为 []byte（零拷贝）
// 注意：返回的 []byte 与原 string 共享底层内存，调用者必须确保不会修改返回的切片
func UnsafeStringToBytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// UnsafeBytesToString 高效地将 []byte 转换为 string（零拷贝）
// 注意：返回的 string 与原 []byte 共享底层内存，调用者必须确保在使用期间原 []byte 不会被修改
func UnsafeBytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}
