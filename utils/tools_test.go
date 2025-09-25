package utils

import "testing"

func TestMaskPhoneNumber(t *testing.T) {
	type args struct {
		phone string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{phone: "13812345678"},
			want: "138****5678",
		},
		{
			name: "test2",
			args: args{phone: "123456"},
			want: "*****6",
		},
		{
			name: "test3",
			args: args{phone: "12345678901"},
			want: "123****8901",
		},
		{
			name: "test4",
			args: args{phone: "1234567"},
			want: "****",
		},
		{
			name: "test5",
			args: args{phone: ""},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskPhoneNumber(tt.args.phone); got != tt.want {
				t.Errorf("MaskPhoneNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaskCustome(t *testing.T) {
	type args struct {
		str   string
		start int
		end   int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{str: "13812345678", start: 3, end: 7},
			want: "138****5678",
		},
		{
			name: "test2",
			args: args{str: "123456", start: 0, end: 2},
			want: "**3456",
		},
		{
			name: "test3",
			args: args{str: "12345678901", start: 3, end: 7},
			want: "123****8901",
		},
		{
			name: "你好",
			args: args{str: "你好", start: 1, end: 2},
			want: "你*",
		},
		{
			name: "从start开始全部替换",
			args: args{str: "13812345678", start: 3, end: -1},
			want: "138********",
		},
		{
			name: "中文从start开始全部替换",
			args: args{str: "你好世界", start: 1, end: -1},
			want: "你***",
		},
		{
			name: "从开头开始全部替换",
			args: args{str: "123456", start: 0, end: -1},
			want: "******",
		},
		{
			name: "start超出范围",
			args: args{str: "123", start: 5, end: -1},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskCustome(tt.args.str, tt.args.start, tt.args.end); got != tt.want {
				t.Errorf("MaskCustome() = %v, want %v", got, tt.want)
			}
		})
	}
}
