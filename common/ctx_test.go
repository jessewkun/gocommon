package common

import (
	"context"
	"reflect"
	"testing"

	"github.com/jessewkun/gocommon/constant"
)

func TestCopyCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want context.Context
	}{
		// TODO: Add test cases.
		{
			name: "test",
			args: args{
				ctx: context.WithValue(context.Background(), constant.CTX_USER_ID, 1),
			},
			want: context.WithValue(context.Background(), constant.CTX_USER_ID, 1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CopyCtx(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CopyCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}
