package nacos

import (
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	// 清空配置
	Cfgs = make(map[string]*Config)

	// 测试空配置
	err := Init()
	if err != nil {
		t.Errorf("Init with empty config should not return error, got: %v", err)
	}

	// 测试有效配置
	Cfgs = map[string]*Config{
		"test": {
			Host:      "localhost",
			Port:      8848,
			Namespace: "public",
			Group:     "DEFAULT_GROUP",
			Timeout:   5000,
		},
	}

	err = Init()
	if err != nil {
		t.Errorf("Init with valid config failed: %v", err)
	}

	// 验证连接是否创建
	client, err := GetConn("test")
	if err != nil {
		t.Errorf("Failed to get test client: %v", err)
	}
	if client == nil {
		t.Error("Client should not be nil")
	}

	// 清理
	Close()
}

func TestSetDefaultConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected *Config
	}{
		{
			name:   "empty config",
			config: &Config{},
			expected: &Config{
				Host:      "localhost",
				Port:      8848,
				Namespace: "public",
				Group:     "DEFAULT_GROUP",
				Timeout:   5000,
			},
		},
		{
			name: "partial config",
			config: &Config{
				Host: "test-host",
				Port: 9999,
			},
			expected: &Config{
				Host:      "test-host",
				Port:      9999,
				Namespace: "public",
				Group:     "DEFAULT_GROUP",
				Timeout:   5000,
			},
		},
		{
			name: "full config",
			config: &Config{
				Host:      "full-host",
				Port:      7777,
				Namespace: "test-namespace",
				Group:     "TEST_GROUP",
				Username:  "test-user",
				Password:  "test-pass",
				Timeout:   3000,
			},
			expected: &Config{
				Host:      "full-host",
				Port:      7777,
				Namespace: "test-namespace",
				Group:     "TEST_GROUP",
				Username:  "test-user",
				Password:  "test-pass",
				Timeout:   3000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setDefaultConfig(tt.config)
			if err != nil {
				t.Errorf("setDefaultConfig failed: %v", err)
			}

			if tt.config.Host != tt.expected.Host {
				t.Errorf("Host mismatch: got %s, want %s", tt.config.Host, tt.expected.Host)
			}
			if tt.config.Port != tt.expected.Port {
				t.Errorf("Port mismatch: got %d, want %d", tt.config.Port, tt.expected.Port)
			}
			if tt.config.Namespace != tt.expected.Namespace {
				t.Errorf("Namespace mismatch: got %s, want %s", tt.config.Namespace, tt.expected.Namespace)
			}
			if tt.config.Group != tt.expected.Group {
				t.Errorf("Group mismatch: got %s, want %s", tt.config.Group, tt.expected.Group)
			}
			if tt.config.Timeout != tt.expected.Timeout {
				t.Errorf("Timeout mismatch: got %d, want %d", tt.config.Timeout, tt.expected.Timeout)
			}
		})
	}
}

func TestGetConn(t *testing.T) {
	// 清空配置和连接
	Cfgs = make(map[string]*Config)
	connList.clients = make(map[string]*Client)

	// 测试获取不存在的连接
	_, err := GetConn("nonexistent")
	if err == nil {
		t.Error("GetConn should return error for nonexistent client")
	}

	// 创建测试连接
	Cfgs["test"] = &Config{
		Host:      "localhost",
		Port:      8848,
		Namespace: "public",
		Group:     "DEFAULT_GROUP",
		Timeout:   5000,
	}

	err = Init()
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 测试获取存在的连接
	client, err := GetConn("test")
	if err != nil {
		t.Errorf("GetConn failed: %v", err)
	}
	if client == nil {
		t.Error("Client should not be nil")
	}

	// 清理
	Close()
}

func TestClose(t *testing.T) {
	// 创建测试连接
	Cfgs = map[string]*Config{
		"test1": {
			Host:      "localhost",
			Port:      8848,
			Namespace: "public",
			Group:     "DEFAULT_GROUP",
			Timeout:   5000,
		},
		"test2": {
			Host:      "localhost",
			Port:      8848,
			Namespace: "public",
			Group:     "DEFAULT_GROUP",
			Timeout:   5000,
		},
	}

	err := Init()
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 验证连接已创建
	if len(connList.clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(connList.clients))
	}

	// 测试关闭
	err = Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// 验证连接已清空
	if len(connList.clients) != 0 {
		t.Errorf("Expected 0 clients after close, got %d", len(connList.clients))
	}
}

func TestMultiInstance(t *testing.T) {
	// 设置多个实例配置
	Cfgs = map[string]*Config{
		"dev": {
			Host:      "dev-nacos.example.com",
			Port:      8848,
			Namespace: "dev",
			Group:     "DEFAULT_GROUP",
			Username:  "dev-user",
			Password:  "dev-pass",
			Timeout:   5000,
		},
		"prod": {
			Host:      "prod-nacos.example.com",
			Port:      8848,
			Namespace: "prod",
			Group:     "DEFAULT_GROUP",
			Username:  "prod-user",
			Password:  "prod-pass",
			Timeout:   5000,
		},
	}

	// 初始化
	err := Init()
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	// 验证所有实例都已创建
	expectedClients := []string{"dev", "prod"}
	for _, name := range expectedClients {
		client, err := GetConn(name)
		if err != nil {
			t.Errorf("Failed to get %s client: %v", name, err)
		}
		if client == nil {
			t.Errorf("Client %s should not be nil", name)
		}
	}

	// 验证连接数量
	if len(connList.clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(connList.clients))
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	expected := &Config{
		Host:      "localhost",
		Port:      8848,
		Namespace: "public",
		Group:     "DEFAULT_GROUP",
		Username:  "",
		Password:  "",
		Timeout:   5000,
	}

	if config.Host != expected.Host {
		t.Errorf("Host mismatch: got %s, want %s", config.Host, expected.Host)
	}
	if config.Port != expected.Port {
		t.Errorf("Port mismatch: got %d, want %d", config.Port, expected.Port)
	}
	if config.Namespace != expected.Namespace {
		t.Errorf("Namespace mismatch: got %s, want %s", config.Namespace, expected.Namespace)
	}
	if config.Group != expected.Group {
		t.Errorf("Group mismatch: got %s, want %s", config.Group, expected.Group)
	}
	if config.Timeout != expected.Timeout {
		t.Errorf("Timeout mismatch: got %d, want %d", config.Timeout, expected.Timeout)
	}
}

func TestConnectionReuse(t *testing.T) {
	// 清空配置
	Cfgs = make(map[string]*Config)

	// 创建配置
	Cfgs["test"] = &Config{
		Host:      "localhost",
		Port:      8848,
		Namespace: "public",
		Group:     "DEFAULT_GROUP",
		Timeout:   5000,
	}

	// 第一次初始化
	err := Init()
	if err != nil {
		t.Fatalf("First Init failed: %v", err)
	}

	// 获取第一个客户端
	client1, err := GetConn("test")
	if err != nil {
		t.Fatalf("Failed to get first client: %v", err)
	}

	// 第二次初始化（应该重用现有连接）
	err = Init()
	if err != nil {
		t.Fatalf("Second Init failed: %v", err)
	}

	// 获取第二个客户端
	client2, err := GetConn("test")
	if err != nil {
		t.Fatalf("Failed to get second client: %v", err)
	}

	// 验证是同一个客户端实例
	if client1 != client2 {
		t.Error("Clients should be the same instance")
	}

	// 清理
	Close()
}

func TestConcurrentAccess(t *testing.T) {
	// 清空配置
	Cfgs = make(map[string]*Config)

	// 创建配置
	Cfgs["test"] = &Config{
		Host:      "localhost",
		Port:      8848,
		Namespace: "public",
		Group:     "DEFAULT_GROUP",
		Timeout:   5000,
	}

	// 初始化
	err := Init()
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	// 并发访问测试
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			client, err := GetConn("test")
			if err != nil {
				t.Errorf("Goroutine %d: Failed to get client: %v", id, err)
				return
			}
			if client == nil {
				t.Errorf("Goroutine %d: Client is nil", id)
				return
			}

			// 模拟一些操作
			time.Sleep(10 * time.Millisecond)
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

func BenchmarkGetConn(b *testing.B) {
	// 设置测试配置
	Cfgs = map[string]*Config{
		"bench": {
			Host:      "localhost",
			Port:      8848,
			Namespace: "public",
			Group:     "DEFAULT_GROUP",
			Timeout:   5000,
		},
	}

	// 初始化
	err := Init()
	if err != nil {
		b.Fatalf("Init failed: %v", err)
	}
	defer Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetConn("bench")
		if err != nil {
			b.Fatalf("GetConn failed: %v", err)
		}
	}
}
