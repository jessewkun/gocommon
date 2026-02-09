package mongodb

import (
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

func TestManagerHealthCheckNilClient(t *testing.T) {
	mgr := &Manager{
		conns: map[string]*mongo.Client{
			"primary": nil,
		},
		configs: map[string]*Config{
			"primary": {MaxPoolSize: 10},
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
