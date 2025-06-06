package utils

import (
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// 判断字符串是否只包含中文
func IsOnlyChinese(str string) bool {
	if len(str) < 1 {
		return false
	}
	var count int
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			count++
		}
	}
	return count == utf8.RuneCountInString(str)
}

// 判断字符串是否只包含数字
func IsOnlyNumber(str string) bool {
	if _, err := strconv.Atoi(str); err == nil {
		return true
	}
	return false
}

// 判断是否是零值
func IsZeroValue(x interface{}) bool {
	if x == nil {
		return true
	}
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

// 判断字符串是否是中国手机号码
func IsChinesePhoneNumber(phone string) bool {
	// 定义中国手机号码的正则表达式
	re := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return re.MatchString(phone)
}

// 判断字符串是否是邮箱
func IsEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// 编码文件名
func EncodeFileName(fileName string) string {
	// 使用 URL 编码对文件名进行编码
	encoded := url.QueryEscape(fileName)
	// 将空格编码从 + 改为 %20
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	return encoded
}

// 清理字符串中的非中文，非数字，非英文字母，非标点符号，非中文标点符号，非空格
func CleanInput(str string) string {
	re := regexp.MustCompile("[^0-9a-zA-Z\u4e00-\u9fa5[:punct:] 、，。！？；：“”‘’（）【】]+")
	return re.ReplaceAllString(str, "")
}

// 清理字符串中的换行符和回车符
func CleanNewline(str string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return -1 // 删除换行符和回车符
		}
		return r
	}, str)
}
