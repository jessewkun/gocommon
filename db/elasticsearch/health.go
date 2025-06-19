package elasticsearch

import (
	"context"
	"time"
)

// HealthCheck ES健康检查
// key 可用 "elasticsearch" 或集群名
func (c *Client) HealthCheck() map[string]*HealthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status := &HealthStatus{
		Timestamp: time.Now().UnixMilli(),
	}

	start := time.Now()
	res, err := c.ES.Cluster.Health(
		c.ES.Cluster.Health.WithContext(ctx),
	)
	status.Latency = time.Since(start).Milliseconds()

	if err != nil {
		status.Status = "error"
		status.Error = err.Error()
		return map[string]*HealthStatus{"elasticsearch": status}
	}
	defer res.Body.Close()
	if res.IsError() {
		status.Status = "error"
		status.Error = res.String()
	} else {
		status.Status = "success"
	}
	return map[string]*HealthStatus{"elasticsearch": status}
}
