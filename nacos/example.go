package nacos

import (
	"fmt"
	"log"
	"time"
)

// ExampleNacosUsage Nacos 使用示例
func ExampleNacosUsage() {
	fmt.Println("---")
	// 1. 定义 Nacos 客户端配置 (通常来自配置文件)
	configs := map[string]*Config{
		"default": {
			Host:      "localhost",
			Port:      8848,
			Namespace: "public",
			Group:     "DEFAULT_GROUP",
			Username:  "",
			Password:  "",
			Timeout:   5000,
			LogLevel:  "info",
			LogDir:    "/tmp/nacos/default_log",
		},
		"dev": {
			Host:      "dev-nacos.example.com", // 假设这是一个实际的Nacos地址
			Port:      8848,
			Namespace: "dev",
			Group:     "DEFAULT_GROUP",
			Username:  "dev-user",
			Password:  "dev-pass",
			Timeout:   5000,
			LogLevel:  "warn",
			CacheDir:  "/tmp/nacos/dev_cache",
		},
	}

	// 2. 创建并初始化 Nacos 管理器
	// NewManager 会尝试连接所有配置的实例
	mgr, err := NewManager(configs)
	if err != nil {
		log.Printf("Failed to create Nacos Manager, some clients might not connect: %v", err)
	}
	// 确保在程序结束时关闭所有连接
	defer func() {
		if closeErr := mgr.Close(); closeErr != nil {
			log.Printf("Error closing Nacos Manager: %v", closeErr)
		}
		fmt.Println("Nacos Manager closed.")
	}()

	// 3. 获取客户端连接
	defaultClient, err := mgr.GetClient("default")
	if err != nil {
		log.Fatalf("Failed to get default Nacos client: %v", err)
	}

	devClient, err := mgr.GetClient("dev")
	if err != nil {
		log.Printf("Failed to get dev Nacos client: %v", err) // dev Nacos可能连接失败
	}

	// 4. 配置管理示例 (使用 defaultClient)
	// 发布配置
	configID := "app-config.json"
	configContent := `{"env": "development", "debug": true, "port": 8080}`
	err = defaultClient.PublishConfig(configID, configContent)
	if err != nil {
		log.Printf("Failed to publish config '%s': %v", configID, err)
	} else {
		fmt.Printf("Config '%s' published successfully\n", configID)
	}

	// 获取配置
	content, err := defaultClient.GetConfig(configID)
	if err != nil {
		log.Printf("Failed to get config '%s': %v", configID, err)
	} else {
		fmt.Printf("Config content '%s': %s\n", configID, content)
	}

	// 监听配置变化
	err = defaultClient.ListenConfig(configID, func(namespace, group, data string) {
		fmt.Printf("Config changed - Namespace: %s, Group: %s, Data: %s\n", namespace, group, data)
	})
	if err != nil {
		log.Printf("Failed to listen config '%s': %v", configID, err)
	}

	// 5. 服务注册与发现示例 (使用 defaultClient)
	serviceName := "example-service"
	serviceIP := "127.0.0.1"
	servicePort := uint64(8080)
	serviceMetadata := map[string]string{
		"env":      "development",
		"version":  "1.0.0",
		"instance": "dev-001",
	}

	// 注册服务
	err = defaultClient.RegisterService(serviceName, serviceIP, servicePort, serviceMetadata)
	if err != nil {
		log.Printf("Failed to register service '%s': %v", serviceName, err)
	} else {
		fmt.Printf("Service '%s' registered successfully\n", serviceName)
	}

	// 获取服务实例
	instances, err := defaultClient.GetService(serviceName)
	if err != nil {
		log.Printf("Failed to get service instances for '%s': %v", serviceName, err)
	} else {
		fmt.Printf("Found %d service instances for '%s'\n", len(instances), serviceName)
		for i, instance := range instances {
			fmt.Printf("  Instance %d: %s:%d (Healthy: %v)\n", i+1, instance.IP, instance.Port, instance.Healthy)
		}
	}

	// 订阅服务变化
	err = defaultClient.SubscribeService(serviceName, func(instances []ServiceInfo) {
		fmt.Printf("Service '%s' instances updated, count: %d\n", serviceName, len(instances))
	})
	if err != nil {
		log.Printf("Failed to subscribe service '%s': %v", serviceName, err)
	}

	// 模拟 devClient 发布配置
	if devClient != nil {
		devConfigID := "dev-app.json"
		devConfigContent := `{"component": "dev-frontend", "debug": true}`
		err = devClient.PublishConfig(devConfigID, devConfigContent)
		if err != nil {
			log.Printf("Failed to publish dev config '%s': %v", devConfigID, err)
		} else {
			fmt.Printf("Dev config '%s' published successfully\n", devConfigID)
		}
	}

	// 等待一段时间以观察配置和服务变化
	fmt.Println("Waiting for 5 seconds to observe Nacos changes...")
	time.Sleep(5 * time.Second)

	// 6. 清理资源 (使用 defaultClient)
	// 取消监听配置
	err = defaultClient.CancelListenConfig(configID)
	if err != nil {
		log.Printf("Failed to cancel listen config '%s': %v", configID, err)
	}

	// 注销服务
	err = defaultClient.DeregisterService(serviceName, serviceIP, servicePort)
	if err != nil {
		log.Printf("Failed to deregister service '%s': %v", serviceName, err)
	} else {
		fmt.Printf("Service '%s' deregistered successfully\n", serviceName)
	}

	// 删除配置
	err = defaultClient.DeleteConfig(configID)
	if err != nil {
		log.Printf("Failed to delete config '%s': %v", configID, err)
	} else {
		fmt.Printf("Config '%s' deleted successfully\n", configID)
	}

	if devClient != nil {
		// 删除 devClient 发布的配置
		devConfigID := "dev-app.json"
		err = devClient.DeleteConfig(devConfigID)
		if err != nil {
			log.Printf("Failed to delete dev config '%s': %v", devConfigID, err)
		} else {
			fmt.Printf("Dev config '%s' deleted successfully\n", devConfigID)
		}
	}

	fmt.Println("\n---")
	fmt.Println("--- Nacos 全局便利层使用示例 ---")
	// 1. 设置全局配置 (通常由 config 包自动加载)
	Cfgs = map[string]*Config{
		"global_default": {
			Host:      "localhost",
			Port:      8848,
			Namespace: "public",
			Group:     "DEFAULT_GROUP",
			Timeout:   5000,
			LogLevel:  "debug",
		},
	}

	// 2. 初始化全局客户端连接
	// 此处假设 config 包未自动调用 Init(), 故手动调用
	if err := Init(); err != nil {
		log.Printf("Failed to initialize global Nacos clients: %v", err)
	}
	// 确保在程序结束时关闭所有连接
	defer func() {
		if closeErr := Close(); closeErr != nil {
			log.Printf("Error closing global Nacos clients: %v", closeErr)
		}
		fmt.Println("Global Nacos clients closed.")
	}()

	// 3. 获取客户端 (注意使用 GetClient 而非 GetConn)
	globalClient, err := GetClient("global_default")
	if err != nil {
		log.Fatalf("Failed to get global_default Nacos client: %v", err)
	}

	// 4. 使用全局客户端
	globalConfigID := "global-app-config.json"
	globalConfigContent := `{"global_env": "production"}`
	err = globalClient.PublishConfig(globalConfigID, globalConfigContent)
	if err != nil {
		log.Printf("Failed to publish global config '%s': %v", globalConfigID, err)
	} else {
		fmt.Printf("Global config '%s' published successfully\n", globalConfigID)
	}

	globalContent, err := globalClient.GetConfig(globalConfigID)
	if err != nil {
		log.Printf("Failed to get global config '%s': %v", globalConfigID, err)
	} else {
		fmt.Printf("Global config content '%s': %s\n", globalConfigID, globalContent)
	}

	// 清理全局客户端发布的配置
	err = globalClient.DeleteConfig(globalConfigID)
	if err != nil {
		log.Printf("Failed to delete global config '%s': %v", globalConfigID, err)
	} else {
		fmt.Printf("Global config '%s' deleted successfully\n", globalConfigID)
	}

	fmt.Println("All Nacos examples completed.")
}

// ExampleMultiInstance 多实例使用示例 (已集成到 ExampleNacosUsage 中，此函数不再需要)
func ExampleMultiInstance() {
	fmt.Println("ExampleMultiInstance functionality is now demonstrated within ExampleNacosUsage.")
	fmt.Println("Please refer to ExampleNacosUsage for multi-instance demonstrations.")
}

// ExampleConfigBased 基于配置文件的使用示例 (已集成到 ExampleNacosUsage 中，此函数不再需要)
func ExampleConfigBased() {
	fmt.Println("ExampleConfigBased functionality is now demonstrated within ExampleNacosUsage.")
	fmt.Println("Configuration loading from file is typically handled by the 'config' package during application startup.")
}
