package oss

import (
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// ACLType 访问控制类型
type ACLType string

const (
	ACLPrivate                 ACLType = "private"
	ACLPublicRead              ACLType = "public-read"
	ACLPublicReadWrite         ACLType = "public-read-write"
	ACLDefault                 ACLType = "default"
	ACLRead                    ACLType = "read"
	ACLWrite                   ACLType = "write"
	ACLFullControl             ACLType = "full-control"
	ACLReadAcp                 ACLType = "read-acp"
	ACLWriteAcp                ACLType = "write-acp"
	ACLReadWriteAcp            ACLType = "read-write-acp"
	ACLReadWriteAcpFullControl ACLType = "read-write-acp-full-control"
)

// StorageClass 存储类型
type StorageClass string

const (
	StorageClassStandard        StorageClass = "Standard"
	StorageClassIA              StorageClass = "IA"
	StorageClassArchive         StorageClass = "Archive"
	StorageClassColdArchive     StorageClass = "ColdArchive"
	StorageClassDeepColdArchive StorageClass = "DeepColdArchive"
)

// HTTPMethod HTTP方法
type HTTPMethod oss.HTTPMethod

const (
	HTTPGet    HTTPMethod = "GET"
	HTTPPut    HTTPMethod = "PUT"
	HTTPPost   HTTPMethod = "POST"
	HTTPDelete HTTPMethod = "DELETE"
	HTTPHead   HTTPMethod = "HEAD"
)

// ObjectInfo 对象信息
type ObjectInfo struct {
	Key          string
	LastModified time.Time
	ETag         string
	Type         string
	Size         int64
	StorageClass StorageClass
	Owner        *Owner
}

// Owner 对象所有者信息
type Owner struct {
	ID          string
	DisplayName string
}

// BucketInfo 存储桶信息
type BucketInfo struct {
	Name         string
	Location     string
	CreationDate time.Time
	StorageClass StorageClass
	Owner        *Owner
}

// ListObjectsResult 列出对象结果
type ListObjectsResult struct {
	Objects        []ObjectInfo
	CommonPrefixes []string
	IsTruncated    bool
	NextMarker     string
	MaxKeys        int
	Delimiter      string
	Prefix         string
}

// UploadPartInfo 分片上传信息
type UploadPartInfo struct {
	PartNumber int
	ETag       string
	Size       int64
}

// MultipartUploadInfo 分片上传信息
type MultipartUploadInfo struct {
	UploadID     string
	Key          string
	Initiated    time.Time
	StorageClass StorageClass
	Owner        *Owner
}

// CopyObjectResult 复制对象结果
type CopyObjectResult struct {
	ETag         string
	LastModified time.Time
}

// DeleteObjectsResult 删除对象结果
type DeleteObjectsResult struct {
	DeletedObjects []string
	Errors         []DeleteError
}

// DeleteError 删除错误信息
type DeleteError struct {
	Key     string
	Code    string
	Message string
}

// BaseObjectOptions 基础对象操作选项
type BaseObjectOptions struct {
	ContentType     string
	ContentEncoding string
	CacheControl    string
	Expires         time.Time
	Metadata        map[string]string
	StorageClass    StorageClass
	ACL             ACLType
}

// ListObjectsOptions 列出对象选项
type ListObjectsOptions struct {
	Prefix    string
	Delimiter string
	Marker    string
	MaxKeys   int
}

// PutObjectOptions 上传对象选项
type PutObjectOptions struct {
	BaseObjectOptions
	ServerSideEncryption string
	SSECustomerAlgorithm string
	SSECustomerKey       string
}

// GetObjectOptions 下载对象选项
type GetObjectOptions struct {
	Range                string
	IfMatch              string
	IfNoneMatch          string
	IfModifiedSince      time.Time
	IfUnmodifiedSince    time.Time
	SSECustomerAlgorithm string
	SSECustomerKey       string
}

// CopyObjectOptions 复制对象选项
type CopyObjectOptions struct {
	BaseObjectOptions
	CopySourceIfMatch           string
	CopySourceIfNoneMatch       string
	CopySourceIfModifiedSince   time.Time
	CopySourceIfUnmodifiedSince time.Time
	MetadataDirective           string
	ServerSideEncryption        string
	SSECustomerAlgorithm        string
	SSECustomerKey              string
}
