package utils

// IncrementByteSlice 将字节切片作为大端序无符号整数进行自增
// 返回一个新的切片，源切片不会被修改
// 如果输入为空切片，返回新的空切片
// 如果发生溢出（如 0xFF...FF），返回相同长度的全零切片
func IncrementByteSlice(src []byte) []byte {
	// 处理空输入
	if len(src) == 0 {
		return []byte{}
	}

	// 创建结果切片，复制源数据
	result := make([]byte, len(src))
	copy(result, src)

	// 从最低字节（最后一个字节）开始自增
	for i := len(result) - 1; i >= 0; i-- {
		result[i]++
		// 如果没有溢出（没有回到 0），自增完成
		if result[i] != 0 {
			return result
		}
		// 如果当前字节溢出（从 0xFF 变成 0x00），继续向高位进位
	}

	// 所有字节都溢出，返回全零切片
	return result
}
