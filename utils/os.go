package utils

import "os"

// EnsureDir 确保目录存在，如果不存在则创建
func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}
