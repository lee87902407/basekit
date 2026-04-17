package utils

import (
	"errors"
	"strconv"
)

const (
	minItoa = -128
	maxItoa = 32768
)

var (
	itoaOffset               [maxItoa - minItoa + 1]uint32
	itoaStrBuffer            string
	itoaByteBuffer           []byte
	invalidByteQuantityError = errors.New("字节必须为正整数，其单位类似M，MB，MiB")
)

// 通过预先计算的数字转字符串,可以快速的将数字转换成字符串
func Itoa(i int64) string {
	if i >= minItoa && i <= maxItoa {
		beg := itoaOffset[i-minItoa]
		if i == maxItoa {
			return itoaStrBuffer[beg:]
		}
		end := itoaOffset[i-minItoa+1]
		return itoaStrBuffer[beg:end]
	}
	return strconv.FormatInt(i, 10)
}

// 通过预先计算的数字转字符串,可以快速的将数字转换成字符串
func Utoa(i uint64) string {
	if i >= 0 && i <= maxItoa {
		beg := itoaOffset[i]
		if i == maxItoa {
			return itoaStrBuffer[beg:]
		}
		end := itoaOffset[i+1]
		return itoaStrBuffer[beg:end]
	}
	return strconv.FormatUint(i, 10)
}

func Bitoa(i int64) []byte {
	if i >= minItoa && i <= maxItoa {
		beg := itoaOffset[i-minItoa]
		if i == maxItoa {
			return itoaByteBuffer[beg:]
		}
		end := itoaOffset[i-minItoa+1]
		return itoaByteBuffer[beg:end]
	}
	return []byte(strconv.FormatInt(i, 10))
}

func Butoa(i uint64) []byte {
	if i >= 0 && i <= maxItoa {
		beg := itoaOffset[i+uint64(-minItoa)]
		if i == maxItoa {
			return itoaByteBuffer[beg:]
		}
		end := itoaOffset[i+uint64(-minItoa)+1]
		return itoaByteBuffer[beg:end]
	}
	return []byte(strconv.FormatUint(i, 10))
}
