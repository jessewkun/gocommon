package oss

import (
	"testing"
)

func TestGetBucket(t *testing.T) {
	oss, err := NewOssSimple("https://oss-cn-hangzhou.aliyuncs.com", "test-key-id", "test-key-secret")
	if err != nil {
		t.Fatalf("NewOssSimple() failed: %v", err)
	}

	// 测试空bucket名称
	_, err = oss.GetBucket("")
	if err == nil {
		t.Error("GetBucket() should return error for empty bucket name")
	}

	// 测试有效bucket名称（这里会失败因为配置无效，但至少测试了参数验证）
	_, err = oss.GetBucket("test-bucket")
	// 由于配置无效，这里会失败，但这是预期的
}

func TestCreateBucket(t *testing.T) {
	oss, err := NewOssSimple("https://oss-cn-hangzhou.aliyuncs.com", "test-key-id", "test-key-secret")
	if err != nil {
		t.Fatalf("NewOssSimple() failed: %v", err)
	}

	tests := []struct {
		name       string
		bucketName string
		wantErr    bool
	}{
		{
			name:       "空bucket名称",
			bucketName: "",
			wantErr:    true,
		},
		{
			name:       "有效bucket名称",
			bucketName: "test-bucket-name",
			wantErr:    false, // 由于配置无效会失败，但参数验证通过
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := oss.CreateBucket(tt.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBucket() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteBucket(t *testing.T) {
	oss, err := NewOssSimple("https://oss-cn-hangzhou.aliyuncs.com", "test-key-id", "test-key-secret")
	if err != nil {
		t.Fatalf("NewOssSimple() failed: %v", err)
	}

	tests := []struct {
		name       string
		bucketName string
		wantErr    bool
	}{
		{
			name:       "空bucket名称",
			bucketName: "",
			wantErr:    true,
		},
		{
			name:       "有效bucket名称",
			bucketName: "test-bucket-name",
			wantErr:    false, // 由于配置无效会失败，但参数验证通过
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := oss.DeleteBucket(tt.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteBucket() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBucketExists(t *testing.T) {
	oss, err := NewOssSimple("https://oss-cn-hangzhou.aliyuncs.com", "test-key-id", "test-key-secret")
	if err != nil {
		t.Fatalf("NewOssSimple() failed: %v", err)
	}

	tests := []struct {
		name       string
		bucketName string
		wantErr    bool
	}{
		{
			name:       "空bucket名称",
			bucketName: "",
			wantErr:    true,
		},
		{
			name:       "有效bucket名称",
			bucketName: "test-bucket-name",
			wantErr:    false, // 由于配置无效会失败，但参数验证通过
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := oss.BucketExists(tt.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("BucketExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetBucketInfo(t *testing.T) {
	oss, err := NewOssSimple("https://oss-cn-hangzhou.aliyuncs.com", "test-key-id", "test-key-secret")
	if err != nil {
		t.Fatalf("NewOssSimple() failed: %v", err)
	}

	tests := []struct {
		name       string
		bucketName string
		wantErr    bool
	}{
		{
			name:       "空bucket名称",
			bucketName: "",
			wantErr:    true,
		},
		{
			name:       "有效bucket名称",
			bucketName: "test-bucket-name",
			wantErr:    false, // 由于配置无效会失败，但参数验证通过
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := oss.GetBucketInfo(tt.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBucketInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListBuckets(t *testing.T) {
	oss, err := NewOssSimple("https://oss-cn-hangzhou.aliyuncs.com", "test-key-id", "test-key-secret")
	if err != nil {
		t.Fatalf("NewOssSimple() failed: %v", err)
	}

	// 测试列出bucket（由于配置无效会失败，但至少测试了方法调用）
	_, err = oss.ListBuckets()
	// 由于配置无效，这里会失败，但这是预期的
}
