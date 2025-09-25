package oss

import (
	"errors"
	"fmt"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// GetBucket 获取指定的Bucket
func (o *Oss) GetBucket(bucketName string) (*oss.Bucket, error) {
	if bucketName == "" {
		return nil, errors.New("Bucket名称不能为空")
	}

	err := o.newClient()
	if err != nil {
		return nil, err
	}

	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return nil, fmt.Errorf("获取Bucket失败: %w", err)
	}
	return bucket, nil
}

// ListBuckets 列出所有bucket
func (o *Oss) ListBuckets() ([]string, error) {
	err := o.newClient()
	if err != nil {
		return nil, err
	}

	lsRes, err := o.client.ListBuckets()
	if err != nil {
		return nil, fmt.Errorf("列出Bucket失败: %w", err)
	}

	buckets := make([]string, 0, len(lsRes.Buckets))
	for _, bucket := range lsRes.Buckets {
		buckets = append(buckets, bucket.Name)
	}

	return buckets, nil
}

// CreateBucket 创建bucket
func (o *Oss) CreateBucket(bucketName string) error {
	if bucketName == "" {
		return errors.New("Bucket名称不能为空")
	}

	err := o.newClient()
	if err != nil {
		return err
	}

	// 使用重试机制
	var lastErr error
	for i := 0; i <= o.config.MaxRetries; i++ {
		err = o.client.CreateBucket(bucketName)
		if err == nil {
			return nil
		}
		lastErr = err

		if i < o.config.MaxRetries {
			time.Sleep(o.config.RetryDelay)
		}
	}

	return fmt.Errorf("创建Bucket失败，重试%d次后仍然失败: %w", o.config.MaxRetries, lastErr)
}

// DeleteBucket 删除bucket
func (o *Oss) DeleteBucket(bucketName string) error {
	if bucketName == "" {
		return errors.New("Bucket名称不能为空")
	}

	err := o.newClient()
	if err != nil {
		return err
	}

	// 使用重试机制
	var lastErr error
	for i := 0; i <= o.config.MaxRetries; i++ {
		err = o.client.DeleteBucket(bucketName)
		if err == nil {
			return nil
		}
		lastErr = err

		if i < o.config.MaxRetries {
			time.Sleep(o.config.RetryDelay)
		}
	}

	return fmt.Errorf("删除Bucket失败，重试%d次后仍然失败: %w", o.config.MaxRetries, lastErr)
}

// BucketExists 检查bucket是否存在
func (o *Oss) BucketExists(bucketName string) (bool, error) {
	if bucketName == "" {
		return false, errors.New("Bucket名称不能为空")
	}

	err := o.newClient()
	if err != nil {
		return false, err
	}

	exists, err := o.client.IsBucketExist(bucketName)
	if err != nil {
		return false, fmt.Errorf("检查Bucket是否存在失败: %w", err)
	}

	return exists, nil
}

// GetBucketInfo 获取bucket信息
func (o *Oss) GetBucketInfo(bucketName string) (*oss.GetBucketInfoResult, error) {
	if bucketName == "" {
		return nil, errors.New("Bucket名称不能为空")
	}

	err := o.newClient()
	if err != nil {
		return nil, err
	}

	info, err := o.client.GetBucketInfo(bucketName)
	if err != nil {
		return nil, fmt.Errorf("获取Bucket信息失败: %w", err)
	}

	return &info, nil
}
