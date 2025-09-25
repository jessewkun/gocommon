# OSS 封装

这是一个对阿里云OSS SDK的封装，提供了更易用、更健壮的接口。

## 主要特性

- **连接池管理**: 使用单例模式管理OSS客户端连接，避免重复创建连接
- **重试机制**: 内置重试机制，提高操作成功率
- **参数验证**: 完善的输入参数验证
- **错误处理**: 详细的错误信息和错误包装
- **类型安全**: 使用Go的类型系统提供类型安全的接口
- **向后兼容**: 保持与原有API的兼容性

## 安装

```bash
go get github.com/aliyun/aliyun-oss-go-sdk/oss
```

## 快速开始

### 基本用法

```go
package main

import (
    "log"
    "github.com/your-project/oss"
)

func main() {
    // 创建OSS客户端
    ossClient, err := oss.NewOssSimple(
        "https://oss-cn-hangzhou.aliyuncs.com",
        "your-access-key-id",
        "your-access-key-secret",
        "your-bucket-name",
    )
    if err != nil {
        log.Fatal(err)
    }
    defer ossClient.Close()

    // 上传文件
    err = ossClient.PutObject("example.txt", "/path/to/local/file.txt")
    if err != nil {
        log.Printf("上传失败: %v", err)
        return
    }

    // 下载文件
    err = ossClient.GetObject("example.txt", "/path/to/download/file.txt")
    if err != nil {
        log.Printf("下载失败: %v", err)
        return
    }

    // 列出对象
    objects, err := ossClient.ListObjects()
    if err != nil {
        log.Printf("列出对象失败: %v", err)
        return
    }

    for _, obj := range objects {
        log.Printf("对象: %s", obj)
    }
}
```

### 高级配置

```go
config := &oss.OssConfig{
    Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
    AccessKeyID:     "your-access-key-id",
    AccessKeySecret: "your-access-key-secret",
    Bucket:          "your-bucket-name",
    MaxConnections:  200,                    // 最大连接数
    ConnectionTimeout: 30 * time.Second,    // 连接超时
    RequestTimeout:    60 * time.Second,    // 请求超时
    MaxRetries:       5,                    // 最大重试次数
    RetryDelay:       2 * time.Second,     // 重试延迟
}

ossClient, err := oss.NewOss(config)
if err != nil {
    log.Fatal(err)
}
```

## API 参考

### 存储桶操作

#### ListBuckets()
列出所有存储桶

```go
buckets, err := ossClient.ListBuckets()
if err != nil {
    log.Printf("列出存储桶失败: %v", err)
    return
}
```

#### CreateBucket(bucketName string) error
创建存储桶

```go
err := ossClient.CreateBucket("new-bucket-name")
if err != nil {
    log.Printf("创建存储桶失败: %v", err)
    return
}
```

#### DeleteBucket(bucketName string) error
删除存储桶

```go
err := ossClient.DeleteBucket("bucket-to-delete")
if err != nil {
    log.Printf("删除存储桶失败: %v", err)
    return
}
```

#### BucketExists(bucketName string) (bool, error)
检查存储桶是否存在

```go
exists, err := ossClient.BucketExists("bucket-name")
if err != nil {
    log.Printf("检查存储桶失败: %v", err)
    return
}
if exists {
    log.Println("存储桶存在")
} else {
    log.Println("存储桶不存在")
}
```

#### GetBucketInfo(bucketName string) (*oss.GetBucketInfoResult, error)
获取存储桶信息

```go
info, err := ossClient.GetBucketInfo("bucket-name")
if err != nil {
    log.Printf("获取存储桶信息失败: %v", err)
    return
}
log.Printf("存储桶名称: %s", info.BucketInfo.Name)
log.Printf("存储桶位置: %s", info.BucketInfo.Location)
```

#### SetBucketACL(bucketName string, acl oss.ACLType) error
设置存储桶访问权限

```go
err := ossClient.SetBucketACL("bucket-name", oss.ACLPublicRead)
if err != nil {
    log.Printf("设置存储桶ACL失败: %v", err)
    return
}
```

#### GetBucketACL(bucketName string) (string, error)
获取存储桶访问权限

```go
acl, err := ossClient.GetBucketACL("bucket-name")
if err != nil {
    log.Printf("获取存储桶ACL失败: %v", err)
    return
}
log.Printf("存储桶ACL: %s", acl)
```

### 对象操作

#### PutObject(objectKey, filePath string) error
上传文件到OSS

```go
err := ossClient.PutObject("remote/path/file.txt", "/local/path/file.txt")
if err != nil {
    log.Printf("上传文件失败: %v", err)
    return
}
```

#### PutObjectWithOptions(objectKey, filePath string, options *ObjectOptions) error
带选项上传文件

```go
options := &oss.ObjectOptions{
    ContentType: "text/plain",
    Metadata: map[string]string{
        "author": "张三",
        "version": "1.0",
    },
}

err := ossClient.PutObjectWithOptions("file.txt", "/local/file.txt", options)
if err != nil {
    log.Printf("上传文件失败: %v", err)
    return
}
```

#### PutObjectFromReader(objectKey string, reader io.Reader, options *ObjectOptions) error
从Reader上传对象

```go
content := strings.NewReader("Hello, OSS!")
options := &oss.ObjectOptions{
    ContentType: "text/plain",
}

err := ossClient.PutObjectFromReader("hello.txt", content, options)
if err != nil {
    log.Printf("上传对象失败: %v", err)
    return
}
```

#### GetObject(objectKey, filePath string) error
下载文件到本地

```go
err := ossClient.GetObject("remote/file.txt", "/local/download/file.txt")
if err != nil {
    log.Printf("下载文件失败: %v", err)
    return
}
```

#### GetObjectToReader(objectKey string) (io.ReadCloser, error)
获取对象为Reader

```go
reader, err := ossClient.GetObjectToReader("file.txt")
if err != nil {
    log.Printf("获取对象失败: %v", err)
    return
}
defer reader.Close()

// 读取内容
content, err := io.ReadAll(reader)
if err != nil {
    log.Printf("读取内容失败: %v", err)
    return
}
log.Printf("文件内容: %s", string(content))
```

#### ListObjects() ([]string, error)
列出所有对象

```go
objects, err := ossClient.ListObjects()
if err != nil {
    log.Printf("列出对象失败: %v", err)
    return
}

for _, obj := range objects {
    log.Printf("对象: %s", obj)
}
```

#### ListObjectsWithPrefix(prefix string) ([]string, error)
列出指定前缀的对象

```go
objects, err := ossClient.ListObjectsWithPrefix("images/")
if err != nil {
    log.Printf("列出对象失败: %v", err)
    return
}

for _, obj := range objects {
    log.Printf("图片对象: %s", obj)
}
```

#### ListObjectsWithDelimiter(prefix, delimiter string) ([]string, []string, error)
列出对象（带分隔符）

```go
objects, prefixes, err := ossClient.ListObjectsWithDelimiter("images/", "/")
if err != nil {
    log.Printf("列出对象失败: %v", err)
    return
}

log.Println("对象:")
for _, obj := range objects {
    log.Printf("  %s", obj)
}

log.Println("公共前缀:")
for _, prefix := range prefixes {
    log.Printf("  %s", prefix)
}
```

#### DeleteObject(objectKey string) error
删除对象

```go
err := ossClient.DeleteObject("file-to-delete.txt")
if err != nil {
    log.Printf("删除对象失败: %v", err)
    return
}
```

#### DeleteObjects(objectKeys []string) error
批量删除对象

```go
objectKeys := []string{"file1.txt", "file2.txt", "file3.txt"}
err := ossClient.DeleteObjects(objectKeys)
if err != nil {
    log.Printf("批量删除对象失败: %v", err)
    return
}
```

#### ObjectExists(objectKey string) (bool, error)
检查对象是否存在

```go
exists, err := ossClient.ObjectExists("file.txt")
if err != nil {
    log.Printf("检查对象失败: %v", err)
    return
}
if exists {
    log.Println("对象存在")
} else {
    log.Println("对象不存在")
}
```

#### GetObjectMeta(objectKey string) (http.Header, error)
获取对象元数据

```go
meta, err := ossClient.GetObjectMeta("file.txt")
if err != nil {
    log.Printf("获取对象元数据失败: %v", err)
    return
}

log.Printf("内容类型: %s", meta.Get("Content-Type"))
log.Printf("内容长度: %s", meta.Get("Content-Length"))
log.Printf("最后修改: %s", meta.Get("Last-Modified"))
```

#### CopyObject(srcObjectKey, destObjectKey string) error
复制对象

```go
err := ossClient.CopyObject("source/file.txt", "destination/file.txt")
if err != nil {
    log.Printf("复制对象失败: %v", err)
    return
}
```

#### GetObjectURL(objectKey string) (string, error)
获取对象访问URL

```go
url, err := ossClient.GetObjectURL("file.txt")
if err != nil {
    log.Printf("获取对象URL失败: %v", err)
    return
}
log.Printf("对象URL: %s", url)
```

#### GetObjectSignedURL(objectKey string, method oss.HTTPMethod, expiredInSec int64) (string, error)
获取对象签名URL

```go
url, err := ossClient.GetObjectSignedURL("file.txt", "GET", 3600)
if err != nil {
    log.Printf("获取签名URL失败: %v", err)
    return
}
log.Printf("签名URL: %s", url)
```

## 配置选项

### OssConfig 结构体

```go
type OssConfig struct {
    Endpoint        string        // OSS端点
    AccessKeyID     string        // 访问密钥ID
    AccessKeySecret string        // 访问密钥
    Bucket          string        // 默认存储桶
    MaxConnections  int           // 最大连接数
    ConnectionTimeout time.Duration // 连接超时
    RequestTimeout    time.Duration // 请求超时
    MaxRetries       int          // 最大重试次数
    RetryDelay       time.Duration // 重试延迟
}
```

### 默认值

- `MaxConnections`: 100
- `ConnectionTimeout`: 30秒
- `RequestTimeout`: 60秒
- `MaxRetries`: 3
- `RetryDelay`: 1秒

## 错误处理

所有方法都返回详细的错误信息，包括：

- 参数验证错误
- 网络连接错误
- OSS服务错误
- 重试失败错误

错误信息使用Go 1.13+的错误包装功能，提供完整的错误链。

## 重试机制

内置重试机制自动处理临时性错误：

- 网络超时
- 服务端错误
- 限流错误

重试次数和延迟可通过配置调整。

## 性能优化

- **连接复用**: 使用单例模式避免重复创建连接
- **内存预分配**: 切片操作使用预分配容量
- **并发安全**: 使用读写锁保护共享资源

## 测试

运行测试：

```bash
go test ./oss/...
```

测试覆盖了：

- 参数验证
- 错误处理
- 重试机制
- 边界情况

## 注意事项

1. **配置验证**: 创建客户端时会验证所有必需的配置参数
2. **资源管理**: 使用完毕后调用`Close()`方法清理资源
3. **错误处理**: 始终检查返回的错误
4. **权限控制**: 确保AccessKey有足够的权限执行操作
5. **网络环境**: 在生产环境中配置适当的超时和重试参数

## 示例

更多示例请参考 `example.go` 文件。

## 许可证

本项目采用MIT许可证。
