package utils

import (
	"testing"
)

func TestMd5X(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"TestMd5X", args{"123"}, "202cb962ac59075b964b07152d234b70"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Md5X(tt.args.str); got != tt.want {
				t.Errorf("Md5X() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAesCbc_Encode(t *testing.T) {
	t.Run("TestAesCbc_Encode1", func(t *testing.T) {
		ac := &AesCbc{
			Key: "flby5t6iJEsShfpdVpMTnUNOEZXhvgDZ",
			Iv:  "r14EWEMYO2144wK2",
		}
		got, err := ac.Encode("wk")
		if err != nil {
			t.Errorf("AesCbc.Encode() error = %v, wantErr %v", err, nil)
			return
		}
		if got != "Y5TYRyomKVaqwNfd4kQGXQ==" {
			t.Errorf("AesCbc.Encode() = %v, want %v", got, "Y5TYRyomKVaqwNfd4kQGXQ==")
		}
	})
	t.Run("TestAesCbc_Encode2", func(t *testing.T) {
		ac := &AesCbc{
			Key: "flby5t6iJEsShfpdVpMTnUNOEZXhvgDZ",
			Iv:  "r14EWEMYO2144wK2",
		}
		got, err := ac.Encode("abc")
		if err != nil {
			t.Errorf("AesCbc.Encode() error = %v, wantErr %v", err, nil)
			return
		}
		if got == "123==" {
			t.Errorf("AesCbc.Encode() = %v, want %v", got, "123==")
		}
	})
}

func TestAesCbc_Decode(t *testing.T) {
	type fields struct {
		Key string
		Iv  string
	}
	type args struct {
		data string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{"TestAesCbc_Decode1", fields{"1234567890123456", "1234567890123456"}, args{"9Hnvri1B0jIn9h5nX87ZXA=="}, "abc", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &AesCbc{
				Key: tt.fields.Key,
				Iv:  tt.fields.Iv,
			}
			got, err := ac.Decode(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("AesCbc.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AesCbc.Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHMACSHA1(t *testing.T) {
	msg := "hello"
	key := "key"
	mac := HMACSHA1(msg, key)
	if len(mac) == 0 {
		t.Error("HMACSHA1 返回值为空")
	}
}

func TestAesCbc_pKCS7UnPadding(t *testing.T) {
	ac := &AesCbc{}

	tests := []struct {
		name     string
		input    []byte
		expected []byte
		wantErr  bool
	}{
		{
			name:     "正常情况",
			input:    []byte{0x01, 0x02, 0x03, 0x01},
			expected: []byte{0x01, 0x02, 0x03},
		},
		{
			name:     "空切片",
			input:    []byte{},
			expected: []byte{},
			wantErr:  true,
		},
		{
			name:     "单个字节",
			input:    []byte{0x01},
			expected: []byte{},
		},
		{
			name:     "无效padding值大于长度",
			input:    []byte{0x10},
			expected: []byte{0x10},
			wantErr:  true,
		},
		{
			name:     "padding值为0",
			input:    []byte{0x01, 0x02, 0x00},
			expected: []byte{0x01, 0x02, 0x00},
			wantErr:  true,
		},
		{
			name:     "padding不一致",
			input:    []byte{0x01, 0x02, 0x03, 0x02},
			expected: []byte{0x01, 0x02, 0x03, 0x02},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ac.pKCS7UnPadding(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("pKCS7UnPadding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(result) != len(tt.expected) {
					t.Errorf("长度不匹配: got %v, want %v", len(result), len(tt.expected))
					return
				}
				for i, v := range result {
					if v != tt.expected[i] {
						t.Errorf("索引 %d 的值不匹配: got %v, want %v", i, v, tt.expected[i])
					}
				}
			}
		})
	}
}

func TestAesCbc_EncodeDecodeCycle(t *testing.T) {
	ac := &AesCbc{
		Key: "flby5t6iJEsShfpdVpMTnUNOEZXhvgDZ",
		Iv:  "r14EWEMYO2144wK2",
	}

	testData := "wk"

	// 加密
	encoded, err := ac.Encode(testData)
	if err != nil {
		t.Errorf("加密失败: %v", err)
		return
	}

	// 解密
	decoded, err := ac.Decode(encoded)
	if err != nil {
		t.Errorf("解密失败: %v", err)
		return
	}

	// 验证结果
	if decoded != testData {
		t.Errorf("加密解密循环失败: 原始数据=%s, 解密结果=%s", testData, decoded)
	}

	t.Logf("加密结果: %s", encoded)
}

func TestAesCbc_SafeEncodeDecodeCycle(t *testing.T) {
	ac := &AesCbc{
		Key: "flby5t6iJEsShfpdVpMTnUNOEZXhvgDZ",
		Iv:  "r14EWEMYO2144wK2",
	}
	ac2 := &AesCbc{
		Key: "flby5t6iJEsShfpdVpMTnUNOEZXhvgDQ",
		Iv:  "r14EWEMYO2144wK1",
	}
	encoded, err := ac.Encode("wk")
	if err != nil {
		t.Errorf("加密失败: %v", err)
		return
	}
	decoded, err := ac.SafeDecode(encoded)
	if err != nil {
		t.Errorf("解密失败: %v", err)
		return
	}
	if decoded != "wk" {
		t.Errorf("解密结果不正确: %s", decoded)
	}
	decoded2, err := ac2.SafeDecode(encoded)
	if err == nil {
		t.Errorf("使用错误的Key/Iv解密应该失败，但得到了结果: %s", decoded2)
		return
	}
	t.Logf("使用错误Key/Iv解密失败，符合预期: %v", err)
}
