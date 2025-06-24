package nacos

import (
	"fmt"
	"log"
	"time"
)

// ExampleNacosUsage Nacos 使用示例
func ExampleNacosUsage() {
	// 1. 初始化 Nacos 配置
	// 在真实应用中，这些配置通常来自配置文件，并通过config.Init()加载到Cfgs中
	// 这里为了演示，我们手动设置
	Cfgs = map[string]*Config{
		"default": {
			Host:      "localhost",
			Port:      8848,
			Namespace: "public",
			Group:     "DEFAULT_GROUP",
			Username:  "",
			Password:  "",
			Timeout:   5000,
		},
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

	// 2. 初始化 Nacos 连接
	if err := Init(); err != nil {
		log.Fatalf("Failed to initialize Nacos: %v", err)
	}
	// 在示例结束时，确保关闭连接
	defer Close()

	// 3. 获取默认客户端连接
	defaultClient, err := GetConn("default")
	if err != nil {
		log.Fatalf("Failed to get default Nacos connection: %v", err)
	}

	// 4. 配置管理示例
	// 发布配置
	err = defaultClient.PublishConfig("app-config.json", `{"env": "development", "debug": true, "port": 8080}`)
	if err != nil {
		log.Printf("Failed to publish config: %v", err)
	} else {
		fmt.Println("Config published successfully")
	}

	// 获取配置
	content, err := defaultClient.GetConfig("app-config.json")
	if err != nil {
		log.Printf("Failed to get config: %v", err)
	} else {
		fmt.Printf("Config content: %s\n", content)
	}

	// 监听配置变化
	err = defaultClient.ListenConfig("app-config.json", func(namespace, group, data string) {
		fmt.Printf("Config changed - Namespace: %s, Group: %s, Data: %s\n", namespace, group, data)
	})
	if err != nil {
		log.Printf("Failed to listen config: %v", err)
	}

	// 5. 服务注册与发现示例
	// 注册服务
	err = defaultClient.RegisterService("example-service", "127.0.0.1", 8080, map[string]string{
		"env":      "development",
		"version":  "1.0.0",
		"instance": "dev-001",
	})
	if err != nil {
		log.Printf("Failed to register service: %v", err)
	} else {
		fmt.Println("Service registered successfully")
	}

	// 获取服务实例
	instances, err := defaultClient.GetService("example-service")
	if err != nil {
		log.Printf("Failed to get service instances: %v", err)
	} else {
		fmt.Printf("Found %d service instances\n", len(instances))
		for i, instance := range instances {
			fmt.Printf("Instance %d: %s:%d (Healthy: %v, Weight: %.2f)\n",
				i+1, instance.IP, instance.Port, instance.Healthy, instance.Weight)
		}
	}

	// 获取一个健康实例
	instance, err := defaultClient.GetServiceOne("example-service")
	if err != nil {
		log.Printf("Failed to get one service instance: %v", err)
	} else {
		fmt.Printf("Selected instance: %s:%d\n", instance.IP, instance.Port)
	}

	// 订阅服务变化
	err = defaultClient.SubscribeService("example-service", func(instances []ServiceInfo) {
		fmt.Printf("Service instances updated, count: %d\n", len(instances))
		for i, instance := range instances {
			fmt.Printf("Updated instance %d: %s:%d\n", i+1, instance.IP, instance.Port)
		}
	})
	if err != nil {
		log.Printf("Failed to subscribe service: %v", err)
	}

	// 6. 多实例使用示例
	// 使用开发环境客户端
	devClient, err := GetConn("dev")
	if err != nil {
		log.Printf("Failed to get dev client: %v", err)
	} else {
		// 在开发环境发布配置
		err = devClient.PublishConfig("dev-config.json", `{"env": "development", "debug": true}`)
		if err != nil {
			log.Printf("Failed to publish dev config: %v", err)
		} else {
			fmt.Println("Dev config published successfully")
		}
	}

	// 使用生产环境客户端
	prodClient, err := GetConn("prod")
	if err != nil {
		log.Printf("Failed to get prod client: %v", err)
	} else {
		// 在生产环境发布配置
		err = prodClient.PublishConfig("prod-config.json", `{"env": "production", "debug": false}`)
		if err != nil {
			log.Printf("Failed to publish prod config: %v", err)
		} else {
			fmt.Println("Prod config published successfully")
		}
	}

	// 7. 清理资源
	// 注销服务
	err = defaultClient.DeregisterService("example-service", "127.0.0.1", 8080)
	if err != nil {
		log.Printf("Failed to deregister service: %v", err)
	} else {
		fmt.Println("Service deregistered successfully")
	}

	// 取消监听配置
	err = defaultClient.CancelListenConfig("app-config.json")
	if err != nil {
		log.Printf("Failed to cancel listen config: %v", err)
	}

	// 删除配置
	err = defaultClient.DeleteConfig("app-config.json")
	if err != nil {
		log.Printf("Failed to delete config: %v", err)
	} else {
		fmt.Println("Config deleted successfully")
	}

	// 等待一段时间以观察配置和服务变化
	time.Sleep(2 * time.Second)
}

// ExampleMultiInstance 多实例使用示例
func ExampleMultiInstance() {
	// 设置多个环境配置
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
		"test": {
			Host:      "test-nacos.example.com",
			Port:      8848,
			Namespace: "test",
			Group:     "DEFAULT_GROUP",
			Username:  "test-user",
			Password:  "test-pass",
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

	// 初始化所有连接
	if err := Init(); err != nil {
		log.Fatalf("Failed to initialize Nacos instances: %v", err)
	}
	defer Close()

	// 遍历所有环境进行操作
	environments := []string{"dev", "test", "prod"}
	for _, env := range environments {
		client, err := GetConn(env)
		if err != nil {
			log.Printf("Failed to get %s client: %v", env, err)
			continue
		}

		// 发布环境特定配置
		configContent := fmt.Sprintf(`{"environment": "%s", "timestamp": "%s"}`, env, time.Now().Format(time.RFC3339))
		err = client.PublishConfig(fmt.Sprintf("%s-config.json", env), configContent)
		if err != nil {
			log.Printf("Failed to publish %s config: %v", env, err)
		} else {
			fmt.Printf("%s config published successfully\n", env)
		}

		// 注册环境特定服务
		err = client.RegisterService(fmt.Sprintf("%s-service", env), "127.0.0.1", 8080, map[string]string{
			"environment": env,
			"version":     "1.0.0",
			"instance":    fmt.Sprintf("%s-001", env),
		})
		if err != nil {
			log.Printf("Failed to register %s service: %v", env, err)
		} else {
			fmt.Printf("%s service registered successfully\n", env)
		}
	}

	// 验证配置和服务
	for _, env := range environments {
		client, err := GetConn(env)
		if err != nil {
			continue
		}

		// 获取配置
		content, err := client.GetConfig(fmt.Sprintf("%s-config.json", env))
		if err != nil {
			log.Printf("Failed to get %s config: %v", env, err)
		} else {
			fmt.Printf("%s config: %s\n", env, content)
		}

		// 获取服务实例
		instances, err := client.GetService(fmt.Sprintf("%s-service", env))
		if err != nil {
			log.Printf("Failed to get %s service: %v", env, err)
		} else {
			fmt.Printf("%s service instances: %d\n", env, len(instances))
		}
	}
}

// ExampleConfigBased 基于配置文件的使用示例
func ExampleConfigBased() {
	// 这个示例展示了如何通过配置文件来管理多个 Nacos 实例
	// 配置文件示例 (config.toml):
	/*
		[nacos]
		[nacos.default]
		host = "localhost"
		port = 8848
		namespace = "public"
		group = "DEFAULT_GROUP"
		timeout = 5000

		[nacos.dev]
		host = "dev-nacos.example.com"
		port = 8848
		namespace = "dev"
		group = "DEFAULT_GROUP"
		username = "dev-user"
		password = "dev-pass"
		timeout = 5000

		[nacos.prod]
		host = "prod-nacos.example.com"
		port = 8848
		namespace = "prod"
		group = "DEFAULT_GROUP"
		username = "prod-user"
		password = "prod-pass"
		timeout = 5000
	*/

	// 在真实应用中，配置会通过 config.Init() 自动加载到 Cfgs 中
	// 然后调用 Init() 初始化所有连接

	fmt.Println("This example demonstrates how to use Nacos with configuration files")
	fmt.Println("The configuration will be automatically loaded and initialized")
}
