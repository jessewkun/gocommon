package oss

import (
	"testing"
	"time"
)

func TestNewOss(t *testing.T) {
	tests := []struct {
		name    string
		config  *OssConfig
		wantErr bool
	}{
		{
			name: "有效配置",
			config: &OssConfig{
				Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
				AccessKeyID:     "test-key-id",
				AccessKeySecret: "test-key-secret",
			},
			wantErr: false,
		},
		{
			name:    "空配置",
			config:  nil,
			wantErr: true,
		},
		{
			name: "缺少Endpoint",
			config: &OssConfig{
				AccessKeyID:     "test-key-id",
				AccessKeySecret: "test-key-secret",
			},
			wantErr: true,
		},
		{
			name: "缺少AccessKeyID",
			config: &OssConfig{
				Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
				AccessKeySecret: "test-key-secret",
			},
			wantErr: true,
		},
		{
			name: "缺少AccessKeySecret",
			config: &OssConfig{
				Endpoint:    "https://oss-cn-hangzhou.aliyuncs.com",
				AccessKeyID: "test-key-id",
			},
			wantErr: true,
		},
		{
			name: "缺少Bucket",
			config: &OssConfig{
				Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
				AccessKeyID:     "test-key-id",
				AccessKeySecret: "test-key-secret",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOss(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOss() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewOssSimple(t *testing.T) {
	endpoint := "https://oss-cn-hangzhou.aliyuncs.com"
	accessKeyID := "test-key-id"
	accessKeySecret := "test-key-secret"

	oss, err := NewOssSimple(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		t.Fatalf("NewOssSimple() failed: %v", err)
	}

	if oss.config.Endpoint != endpoint {
		t.Errorf("Endpoint = %v, want %v", oss.config.Endpoint, endpoint)
	}
	if oss.config.AccessKeyID != accessKeyID {
		t.Errorf("AccessKeyID = %v, want %v", oss.config.AccessKeyID, accessKeyID)
	}
	if oss.config.AccessKeySecret != accessKeySecret {
		t.Errorf("AccessKeySecret = %v, want %v", oss.config.AccessKeySecret, accessKeySecret)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *OssConfig
		wantErr bool
	}{
		{
			name: "有效配置",
			config: &OssConfig{
				Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
				AccessKeyID:     "test-key-id",
				AccessKeySecret: "test-key-secret",
			},
			wantErr: false,
		},
		{
			name: "空Endpoint",
			config: &OssConfig{
				Endpoint:        "",
				AccessKeyID:     "test-key-id",
				AccessKeySecret: "test-key-secret",
			},
			wantErr: true,
		},
		{
			name: "空AccessKeyID",
			config: &OssConfig{
				Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
				AccessKeyID:     "",
				AccessKeySecret: "test-key-secret",
			},
			wantErr: true,
		},
		{
			name: "空AccessKeySecret",
			config: &OssConfig{
				Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
				AccessKeyID:     "test-key-id",
				AccessKeySecret: "",
			},
			wantErr: true,
		},
		{
			name: "空Bucket",
			config: &OssConfig{
				Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
				AccessKeyID:     "test-key-id",
				AccessKeySecret: "test-key-secret",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultValues(t *testing.T) {
	config := &OssConfig{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyID:     "test-key-id",
		AccessKeySecret: "test-key-secret",
	}

	oss, err := NewOss(config)
	if err != nil {
		t.Fatalf("NewOss() failed: %v", err)
	}

	// 检查默认值
	if oss.config.MaxConnections != 100 {
		t.Errorf("MaxConnections = %v, want 100", oss.config.MaxConnections)
	}
	if oss.config.ConnectionTimeout != 30*time.Second {
		t.Errorf("ConnectionTimeout = %v, want 30s", oss.config.ConnectionTimeout)
	}
	if oss.config.RequestTimeout != 60*time.Second {
		t.Errorf("RequestTimeout = %v, want 60s", oss.config.RequestTimeout)
	}
	if oss.config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %v, want 3", oss.config.MaxRetries)
	}
	if oss.config.RetryDelay != 1*time.Second {
		t.Errorf("RetryDelay = %v, want 1s", oss.config.RetryDelay)
	}
}
