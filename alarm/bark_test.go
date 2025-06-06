package alarm

import (
	"context"
	"testing"
)

func TestSendBark(t *testing.T) {
	InitBark(Config{BarkIds: []string{"jT64URJj8b6Fp9Y3nVKJiP"}})
	type args struct {
		ctx     context.Context
		title   string
		content string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "test",
			args: args{
				ctx:     context.Background(),
				title:   "线上报警",
				content: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SendBark(tt.args.ctx, tt.args.title, tt.args.content)
		})
	}
}
