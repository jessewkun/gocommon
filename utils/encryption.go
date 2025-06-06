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
)

type AesCbc struct {
	Key string
	Iv  string
}

// md5
func Md5X(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// cbc加密
func (ac *AesCbc) Encode(data string) (string, error) {
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

// cbc解密
func (ac *AesCbc) Decode(data string) (string, error) {
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
	_data = ac.pKCS7UnPadding(_data)

	return string(_data), nil
}

func (ac *AesCbc) pKCS7Padding(data []byte) []byte {
	padding := aes.BlockSize - len(data)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func (ac *AesCbc) pKCS7UnPadding(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}

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
