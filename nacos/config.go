package nacos

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// GetConfig 获取配置
func (c *Client) GetConfig(dataId string) (string, error) {
	content, err := c.configClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  c.config.Group,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}
	return content, nil
}

// PublishConfig 发布配置
func (c *Client) PublishConfig(dataId, content string) error {
	success, err := c.configClient.PublishConfig(vo.ConfigParam{
		DataId:  dataId,
		Group:   c.config.Group,
		Content: content,
	})
	if err != nil {
		return fmt.Errorf("failed to publish config: %w", err)
	}
	if !success {
		return fmt.Errorf("failed to publish config: success is false")
	}
	return nil
}

// DeleteConfig 删除配置
func (c *Client) DeleteConfig(dataId string) error {
	success, err := c.configClient.DeleteConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  c.config.Group,
	})
	if err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}
	if !success {
		return fmt.Errorf("failed to delete config: success is false")
	}
	return nil
}

// ListenConfig 监听配置变化
func (c *Client) ListenConfig(dataId string, onChange func(string, string, string)) error {
	return c.configClient.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  c.config.Group,
		OnChange: func(namespace, group, dataId, data string) {
			onChange(namespace, group, data)
		},
	})
}

// CancelListenConfig 取消监听配置
func (c *Client) CancelListenConfig(dataId string) error {
	return c.configClient.CancelListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  c.config.Group,
	})
}
