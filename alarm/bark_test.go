package alarm

import (
	"context"
	"testing"
)

func TestSendBark(t *testing.T) {
	originalCfg := Cfg
	Cfg = &Config{
		Bark: &Bark{
			BarkIds: []string{"jT64URJj8b6Fp9Y3nVKJiP"},
		},
		Timeout: 5,
	}
	if err := Init(); err != nil {
		t.Fatalf("InitBark() failed: %v", err)
	}

	t.Cleanup(func() {
		Cfg = originalCfg
		Init()
	})

	// --- Test Cases ---
	type args struct {
		ctx     context.Context
		title   string
		content []string
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
				content: []string{"This is a test message from a unit test."},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Cfg.Bark.Send(tt.args.ctx, tt.args.title, tt.args.content); err != nil {
				t.Errorf("Bark.Send() error = %v, wantErr %v", err, false)
			}
		})
	}
}
