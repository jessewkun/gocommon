package redis

import (
	"testing"

	goredis "github.com/go-redis/redis/v8"
)

func TestManagerHealthCheckNilClient(t *testing.T) {
	mgr := &Manager{
		conns: map[string]goredis.UniversalClient{
			"primary": nil,
		},
	}

	resp := mgr.HealthCheck()
	status, ok := resp["primary"]
	if !ok {
		t.Fatalf("expected status for primary")
	}
	if status.Status != "error" {
		t.Fatalf("expected error status, got %q", status.Status)
	}
	if status.Error == "" {
		t.Fatalf("expected error message")
	}
}
