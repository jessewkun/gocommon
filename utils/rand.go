package utils

import (
	"strings"
	"time"

	"math/rand/v2"
)

// RandomNum 返回指定范围的随机数
func RandomNum(min int, max int) int {
	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))
	return r.IntN(max-min+1) + min
}

// RandomElement 返回 map 中的随机元素
func RandomElement(m map[string]interface{}) (string, interface{}) {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	randomKey := keys[RandomNum(0, len(keys)-1)]
	return randomKey, m[randomKey]
}

// 定义字符集
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomString 返回指定长度的随机字符串，包含大小写字母和数字
func RandomString(n int) string {
	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[r.IntN(len(letterBytes))]
	}
	return string(b)
}

// RandomCode 返回指定长度的随机数字
func RandomCode(n int) string {
	const digits = "0123456789"
	var sb strings.Builder
	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))

	for i := 0; i < n; i++ {
		sb.WriteByte(digits[r.IntN(len(digits))])
	}

	return sb.String()
}
