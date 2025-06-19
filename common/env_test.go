package common

import (
	"testing"

	"github.com/jessewkun/gocommon/config"
)

func TestIsDebug(t *testing.T) {
	oldMode := config.Cfg.Mode
	defer func() { config.Cfg.Mode = oldMode }()
	config.Cfg.Mode = ModeDebug
	if !IsDebug() {
		t.Error("IsDebug 应该返回 true")
	}
	if IsRelease() {
		t.Error("IsRelease 应该返回 false")
	}
	if IsTest() {
		t.Error("IsTest 应该返回 false")
	}
}

func TestIsRelease(t *testing.T) {
	oldMode := config.Cfg.Mode
	defer func() { config.Cfg.Mode = oldMode }()
	config.Cfg.Mode = ModeRelease
	if !IsRelease() {
		t.Error("IsRelease 应该返回 true")
	}
	if IsDebug() {
		t.Error("IsDebug 应该返回 false")
	}
	if IsTest() {
		t.Error("IsTest 应该返回 false")
	}
}

func TestIsTest(t *testing.T) {
	oldMode := config.Cfg.Mode
	defer func() { config.Cfg.Mode = oldMode }()
	config.Cfg.Mode = ModeTest
	if !IsTest() {
		t.Error("IsTest 应该返回 true")
	}
	if IsDebug() {
		t.Error("IsDebug 应该返回 false")
	}
	if IsRelease() {
		t.Error("IsRelease 应该返回 false")
	}
}
