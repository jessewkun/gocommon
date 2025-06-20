package alarm

import (
	"context"
	"testing"
)

func TestSendBark(t *testing.T) {
	originalCfg := Cfg
	Cfg = &Config{
		BarkIds: []string{"jT64URJj8b6Fp9Y3nVKJiP"},
		Timeout: 5,
	}
	if err := InitBark(); err != nil {
		t.Fatalf("InitBark() failed: %v", err)
	}

	t.Cleanup(func() {
		Cfg = originalCfg
		InitBark()
	})

	// --- Test Cases ---
	type args struct {
		ctx     context.Context
		title   string
		content string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "valid bark message",
			args: args{
				ctx:     context.Background(),
				title:   "Go Test Alarm",
				content: "This is a test message from a unit test.",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendBark(tt.args.ctx, tt.args.title, tt.args.content); err != nil {
				t.Errorf("SendBark() error = %v, wantErr %v", err, false)
			}
		})
	}
}
