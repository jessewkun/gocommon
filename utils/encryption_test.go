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
	type fields struct {
		Key string
		Iv  string
	}
	type args struct {
		data string
	}
	t.Run("TestAesCbc_Encode1", func(t *testing.T) {
		ac := &AesCbc{
			Key: "1234567890123456",
			Iv:  "1234567890123456",
		}
		got, err := ac.Encode("abc")
		if err != nil {
			t.Errorf("AesCbc.Encode() error = %v, wantErr %v", err, nil)
			return
		}
		if got != "9Hnvri1B0jIn9h5nX87ZXA==" {
			t.Errorf("AesCbc.Encode() = %v, want %v", got, "9Hnvri1B0jIn9h5nX87ZXA==")
		}
	})
	t.Run("TestAesCbc_Encode2", func(t *testing.T) {
		ac := &AesCbc{
			Key: "1234567890123456",
			Iv:  "1234567890123456",
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
