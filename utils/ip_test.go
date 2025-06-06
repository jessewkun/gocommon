package utils

import (
	"net"
	"testing"
)

func TestIsBan(t *testing.T) {
	tests := []struct {
		name      string
		clientIP  string
		whiteList []string
		want      bool
	}{
		{
			name:      "empty whitelist",
			clientIP:  "192.168.1.1",
			whiteList: []string{},
			want:      false,
		},
		{
			name:      "wildcard rule",
			clientIP:  "192.168.1.1",
			whiteList: []string{"*"},
			want:      false,
		},
		{
			name:      "CIDR match",
			clientIP:  "192.168.1.100",
			whiteList: []string{"192.168.1.0/24"},
			want:      false,
		},
		{
			name:      "exact IP match",
			clientIP:  "10.0.0.1",
			whiteList: []string{"10.0.0.1"},
			want:      false,
		},
		{
			name:      "invalid client IP",
			clientIP:  "invalid-ip",
			whiteList: []string{"192.168.1.0/24"},
			want:      true,
		},
		{
			name:      "no match",
			clientIP:  "172.16.0.1",
			whiteList: []string{"192.168.1.0/24", "10.0.0.1"},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBan(tt.clientIP, tt.whiteList); got != tt.want {
				t.Errorf("IsBan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLocalIP(t *testing.T) {
	tests := []struct {
		name      string
		wantIPv6  bool
		wantError bool
	}{
		{
			name:      "get IPv4",
			wantIPv6:  false,
			wantError: false,
		},
		{
			name:      "get IPv6",
			wantIPv6:  true,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLocalIP(tt.wantIPv6)
			if (err != nil) != tt.wantError {
				t.Errorf("GetLocalIP() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if err == nil {
				ip := net.ParseIP(got)
				if ip == nil {
					t.Errorf("GetLocalIP() returned invalid IP: %v", got)
				}
				if tt.wantIPv6 && ip.To4() != nil {
					t.Errorf("GetLocalIP() returned IPv4 when IPv6 was requested: %v", got)
				}
				if !tt.wantIPv6 && ip.To4() == nil {
					t.Errorf("GetLocalIP() returned IPv6 when IPv4 was requested: %v", got)
				}
			}
		})
	}
}
