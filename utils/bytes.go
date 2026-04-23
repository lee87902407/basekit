package utils

import (
	"strconv"

	"github.com/pkg/errors"
)

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

func RespDataToI64(b []byte) (int64, error) {
	if len(b) != 0 && len(b) < 10 {
		var neg, i = false, 0
		switch b[0] {
		case '-':
			neg = true
			fallthrough
		case '+':
			i++
		}
		if len(b) != i {
			var n int64
			for ; i < len(b) && b[i] >= '0' && b[i] <= '9'; i++ {
				n = int64(b[i]-'0') + n*10
			}
			if len(b) == i {
				if neg {
					n = -n
				}
				return n, nil
			}
		}
	}

	if n, err := strconv.ParseInt(string(b), 10, 64); err != nil {
		return 0, errors.WithStack(err)
	} else {
		return n, nil
	}
}

func RespDataToI32(b []byte) (int, error) {
	if len(b) != 0 && len(b) < 10 {
		var neg, i = false, 0
		switch b[0] {
		case '-':
			neg = true
			fallthrough
		case '+':
			i++
		}
		if len(b) != i {
			var n int
			for ; i < len(b) && b[i] >= '0' && b[i] <= '9'; i++ {
				n = int(b[i]-'0') + n*10
			}
			if len(b) == i {
				if neg {
					n = -n
				}
				return n, nil
			}
		}
	}

	if n, err := strconv.ParseInt(string(b), 10, 32); err != nil {
		return 0, errors.WithStack(err)
	} else {
		return int(n), nil
	}
}

func RespDataToU32(b []byte) (uint32, error) {
	if len(b) != 0 && len(b) < 10 {
		var i = 0
		switch b[0] {
		case '-':
			return 0, &strconv.NumError{Func: "RespDataToU32", Num: string(b), Err: strconv.ErrSyntax}
		case '+':
			i++
		}
		if len(b) != i {
			var n uint32
			for ; i < len(b) && b[i] >= '0' && b[i] <= '9'; i++ {
				n = uint32(b[i]-'0') + n*10
			}
			if len(b) == i {
				return n, nil
			}
		}
	}

	if n, err := strconv.ParseUint(string(b), 10, 32); err != nil {
		return 0, errors.WithStack(err)
	} else {
		return uint32(n), nil
	}
}

func RespDataToU64(b []byte) (uint64, error) {
	if len(b) != 0 && len(b) < 10 {
		var i = 0
		switch b[0] {
		case '-':
			return 0, &strconv.NumError{Func: "RespDataToU64", Num: string(b), Err: strconv.ErrSyntax}
		case '+':
			i++
		}
		if len(b) != i {
			var n uint64
			for ; i < len(b) && b[i] >= '0' && b[i] <= '9'; i++ {
				n = uint64(b[i]-'0') + n*10
			}
			if len(b) == i {
				return n, nil
			}
		}
	}

	if n, err := strconv.ParseUint(string(b), 10, 64); err != nil {
		return 0, errors.WithStack(err)
	} else {
		return n, nil
	}
}
