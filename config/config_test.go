package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoadTOML(t *testing.T) {
	// 创建临时TOML配置文件
	tomlContent := `
mode = "test"
port = ":9000"
domain = "http://localhost:9000"

[log]
path = "./test.log"
closed = false
max_size = 50
max_age = 7
max_backup = 5
transparent_parameter = ["test_id"]
alarm_level = "error"

[alarm]
bark_ids = ["test-device"]
timeout = 3

[http]
transparent_parameter = ["test_id"]
is_log = false

[debug]
module = ["test"]
mode = "log"
`

	tmpFile, err := os.CreateTemp("", "test_config_*.toml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(tomlContent)
	assert.NoError(t, err)
	tmpFile.Close()

	// 测试加载TOML配置
	cfg, err := Init(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// 验证基础配置
	assert.Equal(t, "test", cfg.Mode)
	assert.Equal(t, ":9000", cfg.Port)
	assert.Equal(t, "http://localhost:9000", cfg.Domain)
}

func TestConfigLoadJSON(t *testing.T) {
	// 创建临时JSON配置文件
	jsonContent := `{
		"mode": "test",
		"port": ":9000",
		"domain": "http://localhost:9000",
		"log": {
			"path": "./test.log",
			"closed": false,
			"max_size": 50,
			"max_age": 7,
			"max_backup": 5,
			"transparent_parameter": ["test_id"],
			"alarm_level": "error"
		},
		"alarm": {
			"bark_ids": ["test-device"],
			"timeout": 3
		},
		"http": {
			"transparent_parameter": ["test_id"],
			"is_log": false
		},
		"debug": {
			"module": ["test"],
			"mode": "log"
		}
	}`

	tmpFile, err := os.CreateTemp("", "test_config_*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(jsonContent)
	assert.NoError(t, err)
	tmpFile.Close()

	// 测试加载JSON配置
	cfg, err := Init(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// 验证基础配置
	assert.Equal(t, "test", cfg.Mode)
	assert.Equal(t, ":9000", cfg.Port)
	assert.Equal(t, "http://localhost:9000", cfg.Domain)
}

func TestConfigLoadYAML(t *testing.T) {
	// 创建临时YAML配置文件
	yamlContent := `
mode: test
port: ":9000"
domain: "http://localhost:9000"

log:
  path: "./test.log"
  closed: false
  max_size: 50
  max_age: 7
  max_backup: 5
  transparent_parameter: ["test_id"]
  alarm_level: error

alarm:
  bark_ids: ["test-device"]
  timeout: 3

http:
  transparent_parameter: ["test_id"]
  is_log: false

debug:
  module: ["test"]
  mode: log
`

	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	assert.NoError(t, err)
	tmpFile.Close()

	// 测试加载YAML配置
	cfg, err := Init(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// 验证基础配置
	assert.Equal(t, "test", cfg.Mode)
	assert.Equal(t, ":9000", cfg.Port)
	assert.Equal(t, "http://localhost:9000", cfg.Domain)
}
