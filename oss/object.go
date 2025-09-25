package oss

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// PutObject 上传对象
func (o *Oss) PutObject(bucket string, objectKey string, filePath string) error {
	if objectKey == "" {
		return errors.New("对象键不能为空")
	}
	if filePath == "" {
		return errors.New("文件路径不能为空")
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}

	err := o.newClient()
	if err != nil {
		return err
	}

	bucketObj, err := o.client.Bucket(bucket)
	if err != nil {
		return err
	}

	// 使用重试机制
	var lastErr error
	for i := 0; i <= o.config.MaxRetries; i++ {
		err = bucketObj.PutObjectFromFile(objectKey, filePath)
		if err == nil {
			return nil
		}
		lastErr = err

		if i < o.config.MaxRetries {
			time.Sleep(o.config.RetryDelay)
		}
	}

	return fmt.Errorf("上传对象失败，重试%d次后仍然失败: %w", o.config.MaxRetries, lastErr)
}

// PutObjectFromReader 从Reader上传对象
func (o *Oss) PutObjectFromReader(bucket string, objectKey string, reader io.Reader) error {
	if objectKey == "" {
		return errors.New("对象键不能为空")
	}
	if reader == nil {
		return errors.New("Reader不能为空")
	}

	err := o.newClient()
	if err != nil {
		return err
	}

	bucketObj, err := o.client.Bucket(bucket)
	if err != nil {
		return err
	}

	// 使用重试机制
	var lastErr error
	for i := 0; i <= o.config.MaxRetries; i++ {
		err = bucketObj.PutObject(objectKey, reader)
		if err == nil {
			return nil
		}
		lastErr = err

		if i < o.config.MaxRetries {
			time.Sleep(o.config.RetryDelay)
		}
	}

	return fmt.Errorf("上传对象失败，重试%d次后仍然失败: %w", o.config.MaxRetries, lastErr)
}

// GetObjectToFile 下载对象
func (o *Oss) GetObjectToFile(bucket string, objectKey string, filePath string) error {
	if objectKey == "" {
		return errors.New("对象键不能为空")
	}
	if filePath == "" {
		return errors.New("文件路径不能为空")
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	err := o.newClient()
	if err != nil {
		return err
	}

	bucketObj, err := o.client.Bucket(bucket)
	if err != nil {
		return err
	}

	// 使用重试机制
	var lastErr error
	for i := 0; i <= o.config.MaxRetries; i++ {
		err = bucketObj.GetObjectToFile(objectKey, filePath)
		if err == nil {
			return nil
		}
		lastErr = err

		if i < o.config.MaxRetries {
			time.Sleep(o.config.RetryDelay)
		}
	}

	return fmt.Errorf("下载对象失败，重试%d次后仍然失败: %w", o.config.MaxRetries, lastErr)
}

// GetObjectToReader 下载对象到Reader
func (o *Oss) GetObjectToReader(bucket string, objectKey string) (io.ReadCloser, error) {
	if objectKey == "" {
		return nil, errors.New("对象键不能为空")
	}

	err := o.newClient()
	if err != nil {
		return nil, err
	}

	bucketObj, err := o.client.Bucket(bucket)
	if err != nil {
		return nil, err
	}

	object, err := bucketObj.GetObject(objectKey)
	if err != nil {
		return nil, fmt.Errorf("获取对象失败: %w", err)
	}

	return object, nil
}

// DeleteObject 删除对象
func (o *Oss) DeleteObject(bucket string, objectKey string) error {
	if objectKey == "" {
		return errors.New("对象键不能为空")
	}

	err := o.newClient()
	if err != nil {
		return err
	}

	bucketObj, err := o.client.Bucket(bucket)
	if err != nil {
		return err
	}

	// 使用重试机制
	var lastErr error
	for i := 0; i <= o.config.MaxRetries; i++ {
		err = bucketObj.DeleteObject(objectKey)
		if err == nil {
			return nil
		}
		lastErr = err

		if i < o.config.MaxRetries {
			time.Sleep(o.config.RetryDelay)
		}
	}

	return fmt.Errorf("删除对象失败，重试%d次后仍然失败: %w", o.config.MaxRetries, lastErr)
}

// ObjectExists 检查对象是否存在
func (o *Oss) ObjectExists(bucket string, objectKey string) (bool, error) {
	if objectKey == "" {
		return false, errors.New("对象键不能为空")
	}

	err := o.newClient()
	if err != nil {
		return false, err
	}

	bucketObj, err := o.client.Bucket(bucket)
	if err != nil {
		return false, err
	}

	exists, err := bucketObj.IsObjectExist(objectKey)
	if err != nil {
		return false, fmt.Errorf("检查对象是否存在失败: %w", err)
	}

	return exists, nil
}

// GetObjectURL 获取对象访问URL
func (o *Oss) GetObjectURL(bucket string, objectKey string, expiredInSec int64) (string, error) {
	if objectKey == "" {
		return "", errors.New("对象键不能为空")
	}
	if expiredInSec <= 0 {
		expiredInSec = 3600 // 默认1小时
	}

	err := o.newClient()
	if err != nil {
		return "", err
	}

	bucketObj, err := o.client.Bucket(bucket)
	if err != nil {
		return "", err
	}

	url, err := bucketObj.SignURL(objectKey, oss.HTTPGet, expiredInSec)
	if err != nil {
		return "", fmt.Errorf("生成对象URL失败: %w", err)
	}
	return url, nil
}
