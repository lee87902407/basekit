package utils

import (
	"testing"
)

func TestUpperASCIIFieldString(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "简单小写字母",
			input:    []byte("abc"),
			expected: "ABC",
		},
		{
			name:     "混合大小写",
			input:    []byte("AbC"),
			expected: "ABC",
		},
		{
			name:     "保留下划线和点",
			input:    []byte("a_b.c"),
			expected: "A_B.C",
		},
		{
			name:     "空输入",
			input:    []byte{},
			expected: "",
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			result := UpperASCIIFieldString(tc.input)
			if result != tc.expected {
				t.Errorf("UpperASCIIFieldString(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestLowerASCIIFieldString(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "简单大写字母",
			input:    []byte("ABC"),
			expected: "abc",
		},
		{
			name:     "混合大小写",
			input:    []byte("AbC"),
			expected: "abc",
		},
		{
			name:     "保留下划线和点",
			input:    []byte("A_B.C"),
			expected: "a_b.c",
		},
		{
			name:     "空输入",
			input:    []byte{},
			expected: "",
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			result := LowerASCIIFieldString(tc.input)
			if result != tc.expected {
				t.Errorf("LowerASCIIFieldString(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestASCIIFieldStringRejectsInvalidChar(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "包含空格符应返回空",
			input:    []byte("a b"),
			expected: "",
		},
		{
			name:     "包含连字符应返回空",
			input:    []byte("a-b"),
			expected: "",
		},
		{
			name:     "包含感叹号应返回空",
			input:    []byte("a!b"),
			expected: "",
		},
		{
			name:     "包含@符号应返回空",
			input:    []byte("a@b"),
			expected: "",
		},
		{
			name:     "包含中文应返回空",
			input:    []byte("a中b"),
			expected: "",
		},
		{
			name:     "Upper: 包含空格符应返回空",
			input:    []byte("A B"),
			expected: "",
		},
		{
			name:     "包含数字应返回空-小写",
			input:    []byte("a1b"),
			expected: "",
		},
		{
			name:     "包含数字应返回空-大写",
			input:    []byte("A1B"),
			expected: "",
		},
		{
			name:     "纯数字应返回空",
			input:    []byte("123"),
			expected: "",
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			// 测试 UpperASCIIFieldString
			result := UpperASCIIFieldString(tc.input)
			if result != tc.expected {
				t.Errorf("UpperASCIIFieldString(%q) = %q, want %q", tc.input, result, tc.expected)
			}

			// 测试 LowerASCIIFieldString
			result = LowerASCIIFieldString(tc.input)
			if result != tc.expected {
				t.Errorf("LowerASCIIFieldString(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestUpperASCIIFieldStringSupportsLongInput(t *testing.T) {
	// 构造一个较长的输入字符串（不包含数字）
	longInput := make([]byte, 1000)
	for i := 0; i < 1000; i++ {
		// 循环使用 'a' 到 'z', '_', '.'
		switch i % 30 {
		case 0:
			longInput[i] = '_'
		case 1:
			longInput[i] = '.'
		default:
			longInput[i] = byte('a' + (i % 26))
		}
	}

	result := UpperASCIIFieldString(longInput)

	// 验证结果长度正确
	if len(result) != 1000 {
		t.Errorf("UpperASCIIFieldString 长度错误: got %d, want 1000", len(result))
	}

	// 验证所有字符都已大写且合法（仅字母、下划线和点）
	for i := 0; i < len(result); i++ {
		c := result[i]
		if c == '_' || c == '.' {
			continue
		}
		if c < 'A' || c > 'Z' {
			t.Errorf("UpperASCIIFieldString 结果包含非法字符 %q 于位置 %d", c, i)
		}
	}
}

func TestUpperASCIIFieldByte(t *testing.T) {
	tests := []struct {
		name     string
		input    byte
		expected byte
	}{
		{
			name:     "小写字母 a",
			input:    'a',
			expected: 'A',
		},
		{
			name:     "小写字母 z",
			input:    'z',
			expected: 'Z',
		},
		{
			name:     "大写字母 A 保持不变",
			input:    'A',
			expected: 'A',
		},
		{
			name:     "大写字母 Z 保持不变",
			input:    'Z',
			expected: 'Z',
		},
		{
			name:     "下划线保持不变",
			input:    '_',
			expected: '_',
		},
		{
			name:     "点号保持不变",
			input:    '.',
			expected: '.',
		},
		{
			name:     "数字保持不变",
			input:    '1',
			expected: '1',
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			result := UpperASCIIFieldByte(tc.input)
			if result != tc.expected {
				t.Errorf("UpperASCIIFieldByte(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// indexByte 辅助函数，用于测试中查找字节位置
func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func TestFastRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "长度1",
			length: 1,
		},
		{
			name:   "长度10",
			length: 10,
		},
		{
			name:   "长度100",
			length: 100,
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			result := FastRandomString(tc.length)

			// 验证长度正确
			if len(result) != tc.length {
				t.Errorf("FastRandomString(%d) 长度错误: got %d, want %d", tc.length, len(result), tc.length)
			}

			// 验证所有字符都来自合法字符集
			for j := 0; j < len(result); j++ {
				if indexByte(string(asciiRandomBytes), result[j]) == -1 {
					t.Errorf("FastRandomString 结果包含非法字符 %q 于位置 %d", result[j], j)
				}
			}
		})
	}
}

func TestFastRandomStringReturnsEmptyForNonPositiveLength(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "零长度",
			length: 0,
		},
		{
			name:   "负数长度",
			length: -1,
		},
		{
			name:   "大负数长度",
			length: -100,
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			result := FastRandomString(tc.length)
			if result != "" {
				t.Errorf("FastRandomString(%d) = %q, want 空字符串", tc.length, result)
			}
		})
	}
}

func TestUnsafeStringToBytesAndBytesToString(t *testing.T) {
	// 测试基本转换
	original := "hello world"

	// String -> Bytes
	bytes := UnsafeStringToBytes(original)
	if len(bytes) != len(original) {
		t.Errorf("UnsafeStringToBytes 长度错误: got %d, want %d", len(bytes), len(original))
	}

	// Bytes -> String
	result := BytesToString(bytes)
	if result != original {
		t.Errorf("BytesToString 结果错误: got %q, want %q", result, original)
	}

	// 测试空字符串
	emptyBytes := UnsafeStringToBytes("")
	if len(emptyBytes) != 0 {
		t.Errorf("UnsafeStringToBytes(空字符串) 长度错误: got %d, want 0", len(emptyBytes))
	}

	emptyResult := BytesToString(emptyBytes)
	if emptyResult != "" {
		t.Errorf("BytesToString(空切片) 结果错误: got %q, want 空字符串", emptyResult)
	}

	// 测试特殊字符
	special := "a_b.cXYZ"
	specialBytes := UnsafeStringToBytes(special)
	specialResult := BytesToString(specialBytes)
	if specialResult != special {
		t.Errorf("特殊字符转换错误: got %q, want %q", specialResult, special)
	}
}
