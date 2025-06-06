package common

import "github.com/jessewkun/gocommon/config"

const (
	ModeDebug   = "debug"
	ModeRelease = "release"
	ModeTest    = "test"
)

// isDebug 是否是 debug 模式
// 开发环境
func IsDebug() bool {
	return config.Cfg.Mode == ModeDebug
}

// IsRelease 是否是 release 模式
// 生产环境
func IsRelease() bool {
	return config.Cfg.Mode == ModeRelease
}

// IsTest 是否是 test 模式
// 测试环境
func IsTest() bool {
	return config.Cfg.Mode == ModeTest
}
