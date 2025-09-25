package oss

import (
	"log"
	"time"
)

// ExampleBasicUsage 基本用法示例
func ExampleBasicUsage() {
	// 创建OSS客户端
	ossClient, err := NewOssSimple(
		"https://oss-cn-hangzhou.aliyuncs.com",
		"your-access-key-id",
		"your-access-key-secret",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer ossClient.Close()

	log.Println("OSS客户端创建成功")
}

// ExampleAdvancedConfig 高级配置示例
func ExampleAdvancedConfig() {
	config := &OssConfig{
		Endpoint:          "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyID:       "your-access-key-id",
		AccessKeySecret:   "your-access-key-secret",
		MaxConnections:    200,              // 最大连接数
		ConnectionTimeout: 30 * time.Second, // 连接超时
		RequestTimeout:    60 * time.Second, // 请求超时
		MaxRetries:        5,                // 最大重试次数
		RetryDelay:        2 * time.Second,  // 重试延迟
	}

	ossClient, err := NewOss(config)
	if err != nil {
		log.Fatal(err)
	}
	defer ossClient.Close()

	log.Println("高级配置OSS客户端已创建")
}

// ExampleConfigurationValidation 配置验证示例
func ExampleConfigurationValidation() {
	// 测试无效配置
	invalidConfigs := []*OssConfig{
		nil, // 空配置
		{
			// 缺少Endpoint
			AccessKeyID:     "test",
			AccessKeySecret: "test",
		},
		{
			Endpoint: "https://oss-cn-hangzhou.aliyuncs.com",
			// 缺少AccessKeyID
			AccessKeySecret: "test",
		},
		{
			Endpoint:    "https://oss-cn-hangzhou.aliyuncs.com",
			AccessKeyID: "test",
			// 缺少AccessKeySecret
		},
	}

	for i, config := range invalidConfigs {
		_, err := NewOss(config)
		if err != nil {
			log.Printf("配置 %d 验证失败 (预期): %v", i+1, err)
		} else {
			log.Printf("配置 %d 验证通过 (意外)", i+1)
		}
	}

	// 测试有效配置
	validConfig := &OssConfig{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyID:     "test",
		AccessKeySecret: "test",
	}

	_, err := NewOss(validConfig)
	if err != nil {
		log.Printf("有效配置验证失败: %v", err)
	} else {
		log.Println("有效配置验证通过")
	}
}

// ExampleRetryMechanism 重试机制示例
func ExampleRetryMechanism() {
	// 配置不同的重试策略
	retryConfigs := []*OssConfig{
		{
			Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
			AccessKeyID:     "test",
			AccessKeySecret: "test",
			MaxRetries:      1, // 最少重试
			RetryDelay:      100 * time.Millisecond,
		},
		{
			Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
			AccessKeyID:     "test",
			AccessKeySecret: "test",
			MaxRetries:      5, // 中等重试
			RetryDelay:      1 * time.Second,
		},
		{
			Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
			AccessKeyID:     "test",
			AccessKeySecret: "test",
			MaxRetries:      10, // 最多重试
			RetryDelay:      5 * time.Second,
		},
	}

	for i, config := range retryConfigs {
		log.Printf("重试配置 %d: 最大重试次数=%d, 重试延迟=%v",
			i+1, config.MaxRetries, config.RetryDelay)

		ossClient, err := NewOss(config)
		if err != nil {
			log.Printf("创建客户端失败: %v", err)
			continue
		}

		log.Printf("重试配置 %d 应用成功", i+1)
		ossClient.Close()
	}
}

// ExamplePerformanceOptimization 性能优化示例
func ExamplePerformanceOptimization() {
	// 配置高性能参数
	config := &OssConfig{
		Endpoint:          "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyID:       "your-access-key-id",
		AccessKeySecret:   "your-access-key-secret",
		MaxConnections:    500,                    // 增加连接数
		ConnectionTimeout: 10 * time.Second,       // 减少连接超时
		RequestTimeout:    30 * time.Second,       // 减少请求超时
		MaxRetries:        2,                      // 减少重试次数
		RetryDelay:        500 * time.Millisecond, // 减少重试延迟
	}

	ossClient, err := NewOss(config)
	if err != nil {
		log.Fatal(err)
	}
	defer ossClient.Close()

	// 使用连接池进行并发操作
	// 这里可以启动多个goroutine进行并发操作
	// OSS客户端会自动管理连接池
	log.Println("高性能配置已应用")
}

// ExampleCleanup 资源清理示例
func ExampleCleanup() {
	ossClient, err := NewOssSimple(
		"https://oss-cn-hangzhou.aliyuncs.com",
		"your-access-key-id",
		"your-access-key-secret",
	)
	if err != nil {
		log.Fatal(err)
	}

	// 确保在函数结束时清理资源
	defer func() {
		if err := ossClient.Close(); err != nil {
			log.Printf("关闭OSS客户端失败: %v", err)
		} else {
			log.Println("OSS客户端已关闭")
		}
	}()

	// 执行OSS操作
	// ...

	log.Println("OSS操作完成，资源已清理")
}

// ExampleMain 主函数示例
func ExampleMain() {
	// 这是完整的使用示例
	log.Println("OSS封装使用示例")

	// 1. 基本用法
	ExampleBasicUsage()

	// 2. 高级配置
	ExampleAdvancedConfig()

	// 3. 配置验证
	ExampleConfigurationValidation()

	// 4. 重试机制
	ExampleRetryMechanism()

	// 5. 性能优化
	ExamplePerformanceOptimization()

	// 6. 资源清理
	ExampleCleanup()

	log.Println("所有示例执行完成")
}
