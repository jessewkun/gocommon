package nacos

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// RegisterService 注册服务
func (c *Client) RegisterService(serviceName, ip string, port uint64, metadata map[string]string) error {
	success, err := c.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    metadata,
	})
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}
	if !success {
		return fmt.Errorf("failed to register service: success is false")
	}
	return nil
}

// DeregisterService 注销服务
func (c *Client) DeregisterService(serviceName, ip string, port uint64) error {
	success, err := c.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		Ephemeral:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}
	if !success {
		return fmt.Errorf("failed to deregister service: success is false")
	}
	return nil
}

// GetService 获取服务实例
func (c *Client) GetService(serviceName string) ([]ServiceInfo, error) {
	instances, err := c.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		HealthyOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	var result []ServiceInfo
	for _, ins := range instances {
		result = append(result, ServiceInfo{
			ServiceName: ins.ServiceName,
			IP:          ins.Ip,
			Port:        ins.Port,
			Weight:      ins.Weight,
			Enable:      ins.Enable,
			Healthy:     ins.Healthy,
			Metadata:    ins.Metadata,
		})
	}
	return result, nil
}

// GetServiceOne 获取一个服务实例
func (c *Client) GetServiceOne(serviceName string) (*ServiceInfo, error) {
	instance, err := c.namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get service one: %w", err)
	}
	return &ServiceInfo{
		ServiceName: instance.ServiceName,
		IP:          instance.Ip,
		Port:        instance.Port,
		Weight:      instance.Weight,
		Enable:      instance.Enable,
		Healthy:     instance.Healthy,
		Metadata:    instance.Metadata,
	}, nil
}

// SubscribeService 订阅服务变化
func (c *Client) SubscribeService(serviceName string, onUpdate func([]ServiceInfo)) error {
	return c.namingClient.Subscribe(&vo.SubscribeParam{
		ServiceName: serviceName,
		SubscribeCallback: func(services []model.Instance, err error) {
			if err != nil {
				return
			}
			var result []ServiceInfo
			for _, ins := range services {
				result = append(result, ServiceInfo{
					ServiceName: ins.ServiceName,
					IP:          ins.Ip,
					Port:        ins.Port,
					Weight:      ins.Weight,
					Enable:      ins.Enable,
					Healthy:     ins.Healthy,
					Metadata:    ins.Metadata,
				})
			}
			onUpdate(result)
		},
	})
}

// UnsubscribeService 取消订阅服务
func (c *Client) UnsubscribeService(serviceName string) error {
	return c.namingClient.Unsubscribe(&vo.SubscribeParam{
		ServiceName: serviceName,
	})
}

// GetServices 获取所有服务
func (c *Client) GetServices(pageNo, pageSize int) error {
	_, err := c.namingClient.GetAllServicesInfo(vo.GetAllServiceInfoParam{
		PageNo:   uint32(pageNo),
		PageSize: uint32(pageSize),
	})
	if err != nil {
		return fmt.Errorf("failed to get services: %w", err)
	}
	return nil
}

// Close 关闭客户端
func (c *Client) Close() error {
	// nacos客户端没有显式的关闭方法，这里可以做一些清理工作
	return nil
}
