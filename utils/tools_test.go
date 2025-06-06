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
