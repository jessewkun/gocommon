# gocommon

## 简介

`gocommon` 是一套面向 Go 后端项目的通用基础库。涵盖数据库、缓存、日志、配置、API 响应、工具函数、中间件、告警、调试、服务发现等常用能力，助力快速构建高质量、可扩展的微服务应用。

## 目录结构

```
├── alarm/              # 告警模块（如 Bark 推送）
├── common/             # 通用基础类型与错误处理
├── config/             # 配置加载与管理
├── constant/           # 常量定义
├── cron/               # 定时任务管理
├── db/                 # 数据库与存储
│   ├── mysql/          # mysql模块
│   ├── mongodb/        # mongodb模块
│   ├── redis/          # redis模块
│   ├── elasticsearch/  # elasticsearch模块
│   └── localcache/     # 本地缓存模块
├── debug/              # 调试与动态开关
├── http/               # HTTP 客户端封装
├── logger/             # 日志组件
├── middleware/         # Gin/HTTP 通用中间件
├── nacos/              # Nacos 配置管理与服务发现
├── oss/                # 阿里云 oss 相关实现
├── prometheus/         # prometheus 监控
├── response/           # API 响应结构与绑定
├── router/             # 系统路由定义
├── safego/             # goroutine 安全工具
├── utils/              # 常用工具函数（加解密、IP、时间、随机数等）
└── .vscode/            # VS Code 开发配置
```

## 主要功能模块

### 告警（alarm/）

-   [多渠道告警](./alarm/README.md)

### 配置（common/）

-   [配置管理和热重载](./common/README.md)

### 配置（config/）

-   [配置管理和热重载](./config/README.md)

### 常量（constant/）

- 项目通用常量定义，包含上下文键等

### 定时任务（cron/）

-   [定时任务管理](./cron/README.md)

### 数据库与存储（db/）

-   **MySQL**：[连接池、健康检查、事务、模型等](./db/mysql/README.md)
-   **MongoDB**：[连接池、健康检查、事务等](./db/mongodb/README.md)
-   **Redis**：[连接池、健康检查、Hook 等](./db/redis/README.md)
-   **Elasticsearch**：[索引/文档管理、健康检查等](./db/elasticsearch/README.md)
-   **localcache**：[高性能本地缓存，基于 BigCache](./db/localcache/README.md)

### 调试（debug/）

-   [动态调试](./debug/README.md)

### HTTP 客户端（http/）

-   [高性能 http](./http/README.md)

### 日志（logger/）

-   [多级日志](./logger/README.md)

### 中间件（middleware/）

-   [trace， iolog，cors 等](./middleware/README.md)

### 服务发现与配置管理（nacos/）

-   [配置管理和服务发现功能](./nacos/README.md)

### 对象存储（oss/）

-   [阿里云对象存储](./oss/README.md)

### 性能监控（prometheus/）

-   [阿里云对象存储](./prometheus/README.md)

### API 响应（response/）

-   [标准化 API](./response/README.md)

### 系统路由（router/）

-   [系统路由](./router/README.md)

### goroutine 安全（safego/）

-   [goroutine 安全](./safego/README.md)

### 工具函数（utils/）

-   [工具函数](./utils/README.md)
