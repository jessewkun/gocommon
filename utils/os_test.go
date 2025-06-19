package utils

import (
	"os"
	"testing"
)

func TestEnsureDir(t *testing.T) {
	dir := "./test_tmp_dir"
	defer os.RemoveAll(dir)
	// 目录不存在时
	err := EnsureDir(dir)
	if err != nil {
		t.Errorf("EnsureDir 创建目录失败: %v", err)
	}
	// 目录已存在时
	err = EnsureDir(dir)
	if err != nil {
		t.Errorf("EnsureDir 已存在目录时失败: %v", err)
	}
}
