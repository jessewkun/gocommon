package utils

import (
	"testing"
)

func TestIsOnlyChinese(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"TestIsOnlyChinese", args{"中文"}, true},
		{"TestIsNotOnlyChinese", args{"中文1"}, false},
		{"TestIsNotOnlyChinese", args{"中文a"}, false},
		{"TestIsNotOnlyChinese", args{""}, false},
		{"TestIsNotOnlyChinese", args{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsOnlyChinese(tt.args.str); got != tt.want {
				t.Errorf("IsOnlyChinese() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsOnlyNumber(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"TestIsOnlyNumber", args{"123"}, true},
		{"TestIsOnlyNumber", args{"a123"}, false},
		{"TestIsOnlyNumber", args{""}, false},
		{"TestIsOnlyNumber", args{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsOnlyNumber(tt.args.str); got != tt.want {
				t.Errorf("IsOnlyNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsZeroValue(t *testing.T) {
	type args struct {
		x interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"TestIsZeroValue", args{0}, true},
		{"TestIsZeroValue1", args{1}, false},
		{"TestIsZeroValuefalse", args{false}, true},
		{"TestIsZeroValuetrue", args{true}, false},
		{"TestIsZeroValueemptystring", args{""}, true},
		{"TestIsZeroValuestringa", args{"a"}, false},
		{"TestIsZeroValueemptyargs", args{}, true},
		{"TestIsZeroValueargsnil", args{nil}, true},
		{"TestIsZeroValueemptystruct", args{struct{}{}}, true},
		{"TestIsZeroValueemptyintslice", args{[]int{}}, false},
		{"TestIsZeroValueemptystringslice", args{[]string{}}, false},
		{"TestIsZeroValueemptymap", args{map[string]interface{}{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsZeroValue(tt.args.x); got != tt.want {
				t.Errorf("IsZeroValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
