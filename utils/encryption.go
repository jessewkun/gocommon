package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
)

// AesCbc 提供aes-cbc加密解密功能
type AesCbc struct {
	Key string
	Iv  string
}

// Md5X 计算md5值
func Md5X(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// Encode cbc加密
func (ac *AesCbc) Encode(data string) (string, error) {
	if data == "" {
		return "", nil
	}
	_data := []byte(data)
	_key := []byte(ac.Key)
	_iv := []byte(ac.Iv)

	_data = ac.pKCS7Padding(_data)
	block, err := aes.NewCipher(_key)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, _iv)
	mode.CryptBlocks(_data, _data)
	return base64.StdEncoding.EncodeToString(_data), nil
}

// Decode cbc解密
func (ac *AesCbc) Decode(data string) (string, error) {
	if data == "" {
		return "", nil
	}
	_data, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	_key := []byte(ac.Key)
	_iv := []byte(ac.Iv)

	block, err := aes.NewCipher(_key)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCDecrypter(block, _iv)
	mode.CryptBlocks(_data, _data)
	_data, err = ac.pKCS7UnPadding(_data)
	if err != nil {
		return "", err
	}

	return string(_data), nil
}

// SafeEncode 安全加密包装函数，防止panic
func (ac *AesCbc) SafeEncode(data string) (string, error) {
	var result string
	var encodeErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				encodeErr = errors.New("加密过程发生panic")
			}
		}()
		result, encodeErr = ac.Encode(data)
	}()
	return result, encodeErr
}

// SafeDecode 安全加密包装函数，防止panic
func (ac *AesCbc) SafeDecode(data string) (string, error) {
	var result string
	var decodeErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				decodeErr = errors.New("解密过程发生panic")
			}
		}()
		result, decodeErr = ac.Decode(data)
	}()
	return result, decodeErr
}

// pKCS7Padding 添加PKCS7填充
func (ac *AesCbc) pKCS7Padding(data []byte) []byte {
	padding := aes.BlockSize - len(data)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// pKCS7UnPadding 移除PKCS7填充
func (ac *AesCbc) pKCS7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return data, errors.New("数据为空")
	}
	unpadding := int(data[length-1])
	if unpadding == 0 || unpadding > aes.BlockSize || unpadding > length {
		return data, errors.New("padding值非法")
	}
	for i := length - unpadding; i < length; i++ {
		if data[i] != byte(unpadding) {
			return data, errors.New("padding内容非法")
		}
	}
	return data[:(length - unpadding)], nil
}

// HMACSHA1 计算HMAC-SHA1值
func HMACSHA1(message, key string) []byte {
	// 将密钥转换为字节数组
	keyBytes := []byte(key)

	// 创建 HMAC-SHA1 对象
	hmacSHA1 := hmac.New(sha1.New, keyBytes)

	// 写入消息
	hmacSHA1.Write([]byte(message))

	// 计算 HMAC，并将结果转换为十六进制字符串
	return hmacSHA1.Sum(nil)
}
