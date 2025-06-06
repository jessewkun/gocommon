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
