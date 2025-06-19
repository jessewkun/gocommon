# gocommon

## 简介

`gocommon` 是一套面向 Go 后端项目的通用基础库，涵盖数据库、缓存、日志、配置、API 响应、工具函数、中间件、告警、调试等常用能力，助力快速构建高质量服务。

## 目录结构

```
├── alarm/           # 告警模块（如 Bark 推送）
├── common/          # 通用基础类型与错误处理
├── config/          # 配置加载与管理
├── constant/        # 常量定义
├── db/              # 数据库与存储
│   ├── mysql/
│   ├── mongodb/
│   ├── redis/
│   └── elasticsearch/
├── debug/           # 调试与动态开关
├── http/            # HTTP 客户端封装
├── logger/          # 日志组件
├── middleware/      # Gin/HTTP 通用中间件
├── response/        # API 响应结构与绑定
├── safego/          # goroutine 安全工具
├── utils/           # 常用工具函数（加解密、IP、时间、随机数等）
```

## 主要功能模块

### 数据库与存储（db/）

-   **MySQL**：[连接池、健康检查、事务、模型等](./db/mysql/README.md)
-   **MongoDB**：[连接池、健康检查、事务等](./db/mongodb/README.md)
-   **Redis**：[连接池、健康检查、Hook 等](./db/redis/README.md)
-   **Elasticsearch**：[索引/文档管理、健康检查等](./db/elasticsearch/README.md)

### 日志（logger/）

-   基于 zap 的高性能日志，支持多级别、上下文、trace、报警等

### 配置（config/）

-   支持 toml/yaml/json，自动映射结构体

### API 响应（response/）

-   标准 API 返回结构、错误处理、参数绑定与校验

### 工具函数（utils/）

-   加解密、IP、时间、随机数、脱敏、类型判断等

### 中间件（middleware/）

-   JWT、CORS、限流、异常恢复、链路追踪、IO 日志等

### 告警（alarm/）

-   支持 Bark 推送等

### goroutine 安全（safego/）

-   panic recover、WaitGroup 封装

### 调试（debug/）

-   动态模块开关、调试日志输出

### 常量（constant/）

-   项目通用常量

### 其他

-   HTTP 客户端（http/）
-   通用类型与错误（common/）

## 快速开始

1. 安装依赖
    ```sh
    go mod tidy
    ```
2. 参考各模块目录下的 example.go 或测试用例，快速集成到你的项目。

## 测试

建议在本地或 CI 环境下运行所有测试：

```sh
go test ./...
```

## 贡献

欢迎 issue、PR 及建议！

## License

MIT
