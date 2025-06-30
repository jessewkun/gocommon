package nacos

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// GetConfig 获取配置
func (c *Client) GetConfig(dataID string) (string, error) {
	content, err := c.configClient.GetConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  c.config.Group,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}
	return content, nil
}

// PublishConfig 发布配置
func (c *Client) PublishConfig(dataID, content string) error {
	success, err := c.configClient.PublishConfig(vo.ConfigParam{
		DataId:  dataID,
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
func (c *Client) DeleteConfig(dataID string) error {
	success, err := c.configClient.DeleteConfig(vo.ConfigParam{
		DataId: dataID,
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
func (c *Client) ListenConfig(dataID string, onChange func(string, string, string)) error {
	return c.configClient.ListenConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  c.config.Group,
		OnChange: func(namespace, group, dataID, data string) {
			onChange(namespace, group, data)
		},
	})
}

// CancelListenConfig 取消监听配置
func (c *Client) CancelListenConfig(dataID string) error {
	return c.configClient.CancelListenConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  c.config.Group,
	})
}
