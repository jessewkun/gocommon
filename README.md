# gocommon

## 简介

`gocommon` 是一套面向 Go 后端项目的通用基础库，涵盖数据库、缓存、日志、配置、API 响应、工具函数、中间件、告警、调试、服务发现等常用能力，助力快速构建高质量服务。

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
│   ├── elasticsearch/
│   └── localcache/  # 本地缓存模块
├── debug/           # 调试与动态开关
├── http/            # HTTP 客户端封装
├── logger/          # 日志组件
├── middleware/      # Gin/HTTP 通用中间件
├── nacos/           # Nacos 配置管理与服务发现
├── response/        # API 响应结构与绑定
├── safego/          # goroutine 安全工具
├── utils/           # 常用工具函数（加解密、IP、时间、随机数等）
└── .vscode/         # VS Code 开发配置
```

## 主要功能模块

### 数据库与存储（db/）

-   **MySQL**：[连接池、健康检查、事务、模型等](./db/mysql/README.md)
    -   支持 BaseModel 基础模型，包含 ID、CreatedAt、ModifiedAt 字段
    -   自定义 DateTime 类型，支持 JSON 序列化
-   **MongoDB**：[连接池、健康检查、事务等](./db/mongodb/README.md)
-   **Redis**：[连接池、健康检查、Hook 等](./db/redis/README.md)
-   **Elasticsearch**：[索引/文档管理、健康检查等](./db/elasticsearch/README.md)
-   **localcache**：[高性能本地缓存，基于 BigCache](./db/localcache/README.md)
    -   零 GC 压力，适合高并发、大容量缓存场景
    -   支持 TTL、类型安全缓存、缓存管理等功能
    -   提供 Cache、TypedCache、Manager 三种接口
    -   包含完整的技术选型对比文档

### 服务发现与配置管理（nacos/）

-   **Nacos 模块**：[配置管理和服务发现功能](./nacos/README.md)
-   支持多实例配置管理
-   配置的发布、获取、删除及监听
-   服务注册与发现
-   自动配置加载和初始化

### 日志（logger/）

-   基于 zap 的高性能日志，支持多级别、上下文、trace、报警等

### 配置（config/）

-   支持 TOML/JSON/YAML 格式配置文件，自动映射结构体
-   支持热更新配置
    -   http/
    -   alarm/
    -   debug/
-   提供完整的配置文件示例：`config.toml.example` 和 `config.json.example`

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
