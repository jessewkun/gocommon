# gocommon

## 简介

`prometheus` 提供 Prometheus 格式的监控指标采集，配合中间件自动收集 HTTP 和系统指标。

## 监控指标（prometheus/）

**HTTP 业务指标**：
- **请求统计**：记录请求总数、响应状态码等（`http_requests_total`）
- **性能监控**：记录请求处理时长分布，支持 P50/P95/P99 分位数统计（`http_request_duration_seconds`）
- **路由识别**：智能识别注册的路由，避免动态路径产生过多标签

**Go 运行时指标**（自动包含）：
- **Goroutine 监控**：`go_goroutines`（数量）、`go_threads`（线程数）
- **内存监控**：`go_memstats_heap_alloc_bytes`（堆内存）、`go_memstats_sys_bytes`（系统内存）等
- **GC 监控**：`go_gc_duration_seconds`（GC 耗时）、`go_memstats_gc_cpu_fraction`（GC CPU 占用）等

**进程指标**（自动包含）：
- **CPU 使用**：`process_cpu_seconds_total`
- **内存占用**：`process_resident_memory_bytes`（RSS）
- **文件描述符**：`process_open_fds`、`process_max_fds`
- **启动时间**：`process_start_time_seconds`

> 💡 **说明**：从 prometheus/client_golang v1.12 开始，DefaultRegistry 在初始化时自动注册了 Go runtime 和进程相关的 collectors，无需手动注册
