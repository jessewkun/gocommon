// Package utils 提供一些常用的工具函数
package utils

import (
	"strings"
)

// MaskPhoneNumber 手机号码脱敏处理
// 规则：
// 1. 手机号保留前3位和后4位，中间用*代替
// 2. 非11位数字，根据长度智能脱敏
// 3. 空字符串或非法字符返回空字符串
func MaskPhoneNumber(phone string) string {
	// 去除空格和特殊字符
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return ""
	}

	length := len(phone)
	switch {
	case length < 7: // 短号码，只显示最后一位
		return strings.Repeat("*", length-1) + phone[length-1:]
	case length == 11: // 标准手机号
		return phone[:3] + strings.Repeat("*", 4) + phone[7:]
	default: // 其他长度号码，保留前3位和后4位
		if length < 8 {
			return strings.Repeat("*", 4)
		}
		return phone[:3] + strings.Repeat("*", length-7) + phone[length-4:]
	}
}

// MaskCustome 自定义脱敏
// 规则：
// 1. 根据起始位置和结束位置，用*代替中间的字符
// 2. 起始位置不能为负数
// 3. 如果end为-1，则从start开始全部替换
// 4. 如果end不为-1，则起始位置不能大于结束位置，且结束位置不能大于字符串长度
// 5. 空字符串或非法字符返回空字符串
// 6. 支持中文字符，按字符位置而非字节位置处理
func MaskCustome(str string, start int, end int) string {
	if start < 0 {
		return ""
	}

	// 将字符串转换为rune切片，以正确处理中文字符
	runes := []rune(str)

	// 如果end为-1，则从start开始全部替换
	if end == -1 {
		if start >= len(runes) {
			return ""
		}
		var result strings.Builder
		result.WriteString(string(runes[:start]))
		result.WriteString(strings.Repeat("*", len(runes)-start))
		return result.String()
	}

	// 正常的end参数处理
	if end < 0 || start > end || end > len(runes) {
		return ""
	}

	// 构建结果字符串，用对应数量的星号替换
	var result strings.Builder
	result.WriteString(string(runes[:start]))
	result.WriteString(strings.Repeat("*", end-start))
	result.WriteString(string(runes[end:]))

	return result.String()
}
