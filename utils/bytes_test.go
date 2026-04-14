package utils

import (
	"bytes"
	"testing"
)

func TestIncrementByteSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "简单自增",
			input:    []byte{0x00, 0x00, 0x01},
			expected: []byte{0x00, 0x00, 0x02},
		},
		{
			name:     "中间位置进位",
			input:    []byte{0x00, 0xFF, 0xFF},
			expected: []byte{0x01, 0x00, 0x00},
		},
		{
			name:     "单字节自增",
			input:    []byte{0x42},
			expected: []byte{0x43},
		},
		{
			name:     "全零输入",
			input:    []byte{0x00, 0x00},
			expected: []byte{0x00, 0x01},
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			result := IncrementByteSlice(tc.input)
			if !bytes.Equal(result, tc.expected) {
				t.Errorf("IncrementByteSlice(%v) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestIncrementByteSliceCarry(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "低字节进位",
			input:    []byte{0x00, 0x00, 0xFF},
			expected: []byte{0x00, 0x01, 0x00},
		},
		{
			name:     "多字节连续进位",
			input:    []byte{0x00, 0xFF, 0xFF},
			expected: []byte{0x01, 0x00, 0x00},
		},
		{
			name:     "中间字节进位",
			input:    []byte{0x01, 0xFF, 0xFF},
			expected: []byte{0x02, 0x00, 0x00},
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			result := IncrementByteSlice(tc.input)
			if !bytes.Equal(result, tc.expected) {
				t.Errorf("IncrementByteSlice(%v) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestIncrementByteSliceOverflow(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "单字节溢出",
			input:    []byte{0xFF},
			expected: []byte{0x00},
		},
		{
			name:     "多字节溢出",
			input:    []byte{0xFF, 0xFF, 0xFF},
			expected: []byte{0x00, 0x00, 0x00},
		},
		{
			name:     "两字节溢出",
			input:    []byte{0xFF, 0xFF},
			expected: []byte{0x00, 0x00},
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			result := IncrementByteSlice(tc.input)
			if !bytes.Equal(result, tc.expected) {
				t.Errorf("IncrementByteSlice(%v) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestIncrementByteSliceEmpty(t *testing.T) {
	// 空输入应返回新的空切片
	input := []byte{}
	result := IncrementByteSlice(input)

	if len(result) != 0 {
		t.Errorf("IncrementByteSlice(空切片) 长度错误: got %d, want 0", len(result))
	}

	// 确保返回的是新切片（非 nil）
	if result == nil {
		t.Errorf("IncrementByteSlice(空切片) 不应返回 nil")
	}
}

func TestIncrementByteSliceNotModifySource(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "普通输入",
			input: []byte{0x01, 0x02, 0x03},
		},
		{
			name:  "进位情况",
			input: []byte{0x00, 0xFF, 0xFF},
		},
		{
			name:  "溢出情况",
			input: []byte{0xFF, 0xFF},
		},
		{
			name:  "单字节",
			input: []byte{0x42},
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			// 保存原始输入的副本
			original := make([]byte, len(tc.input))
			copy(original, tc.input)

			// 执行自增
			_ = IncrementByteSlice(tc.input)

			// 验证源切片未被修改
			if !bytes.Equal(tc.input, original) {
				t.Errorf("IncrementByteSlice 修改了源切片: got %v, want %v", tc.input, original)
			}
		})
	}
}

func TestIncrementByteSliceReturnsNewSlice(t *testing.T) {
	input := []byte{0x01, 0x02, 0x03}
	result := IncrementByteSlice(input)

	// 验证返回的是新切片，不是源切片
	if &result[0] == &input[0] {
		t.Errorf("IncrementByteSlice 应返回新的切片，而非复用源切片")
	}
}
