package utils

import (
	"reflect"
	"testing"
)

func Test_RandomNum(t *testing.T) {
	type args struct {
		min int
		max int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"test1", args{1, 10}, 5},
		{"test1", args{0, 0}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandomNum(tt.args.min, tt.args.max)
			t.Logf("got: %d", got)
		})
	}
}

func TestRandomElement(t *testing.T) {
	type args struct {
		m map[string]interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 interface{}
	}{
		{"test1", args{map[string]interface{}{"a": 1}}, "a", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := RandomElement(tt.args.m)
			if got != tt.want {
				t.Errorf("RandomElement() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("RandomElement() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestRandomString(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test1", args{11}, "1234567890"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(RandomString(tt.args.n))
		})
	}
}
