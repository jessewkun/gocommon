package mysql

import (
	"testing"

	"gorm.io/gorm"
)

func TestManagerHealthCheckNilDB(t *testing.T) {
	mgr := &Manager{
		conns: map[string]*gorm.DB{
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
