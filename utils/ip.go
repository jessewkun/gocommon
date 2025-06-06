package utils

import (
	"bytes"
	"fmt"
	"net"
	"sort"
)

// IsBan 检查客户端IP是否被禁止访问
// clientIP: 客户端IP地址
// whiteList: IP白名单，支持具体IP和CIDR格式
// 返回值: true表示被禁止，false表示允许访问
func IsBan(clientIP string, whiteList []string) bool {
	// 参数验证
	if clientIP == "" {
		return true
	}
	if len(whiteList) == 0 {
		return false
	}

	// 解析客户端IP
	parsedIP := net.ParseIP(clientIP)
	if parsedIP == nil {
		return true // 无效的IP地址格式
	}

	// 检查白名单
	for _, rule := range whiteList {
		// 通配符规则
		if rule == "*" {
			return false
		}

		// CIDR格式检查
		if _, ipNet, err := net.ParseCIDR(rule); err == nil {
			if ipNet.Contains(parsedIP) {
				return false
			}
			continue
		}

		// 具体IP匹配
		if ruleIP := net.ParseIP(rule); ruleIP != nil {
			if ruleIP.Equal(parsedIP) {
				return false
			}
			continue
		}
	}

	return true
}

// GetLocalIP 获取本地优先级最高的非回环IPv4地址
// 如果需要IPv6地址，可以通过 preferIPv6 参数控制
func GetLocalIP(preferIPv6 ...bool) (string, error) {
	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var candidateIPs []net.IP
	wantIPv6 := len(preferIPv6) > 0 && preferIPv6[0]

	// 遍历所有网络接口
	for _, iface := range interfaces {
		// 忽略关闭和未启用的接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// 获取接口的地址
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}

			// 根据需求选择 IPv4 或 IPv6
			if wantIPv6 {
				if ipNet.IP.To4() == nil && ipNet.IP.To16() != nil {
					candidateIPs = append(candidateIPs, ipNet.IP)
				}
			} else {
				if ip := ipNet.IP.To4(); ip != nil {
					candidateIPs = append(candidateIPs, ip)
				}
			}
		}
	}

	if len(candidateIPs) == 0 {
		return "", fmt.Errorf("no available %s address found",
			map[bool]string{true: "IPv6", false: "IPv4"}[wantIPv6])
	}

	// 按优先级排序 IP 地址
	sort.Slice(candidateIPs, func(i, j int) bool {
		return isPreferredIP(candidateIPs[i], candidateIPs[j])
	})

	return candidateIPs[0].String(), nil
}

// isPreferredIP 判断 IP 地址优先级
// 优先级规则：
// 1. 公网地址优先于私有地址
// 2. 较小的私有网段优先（如 10.0.0.0/8 优先于 172.16.0.0/12）
func isPreferredIP(ip1, ip2 net.IP) bool {
	// 转换为 IPv4 格式
	ip1v4 := ip1.To4()
	ip2v4 := ip2.To4()
	if ip1v4 == nil || ip2v4 == nil {
		return false
	}

	// 检查是否为私有地址
	ip1Private := IsPrivateIP(ip1v4)
	ip2Private := IsPrivateIP(ip2v4)

	// 公网地址优先
	if !ip1Private && ip2Private {
		return true
	}
	if ip1Private && !ip2Private {
		return false
	}

	// 都是私有地址时，比较网段
	if ip1Private && ip2Private {
		return ip1v4[0] < ip2v4[0]
	}

	// 都是公网地址时，保持原顺序
	return true
}

// IsPrivateIP 检查是否为私有 IP 地址
func IsPrivateIP(ip net.IP) bool {
	privateNetworks := []struct {
		start net.IP
		end   net.IP
	}{
		{net.ParseIP("10.0.0.0"), net.ParseIP("10.255.255.255")},
		{net.ParseIP("172.16.0.0"), net.ParseIP("172.31.255.255")},
		{net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255")},
	}

	for _, network := range privateNetworks {
		if bytes.Compare(ip, network.start) >= 0 && bytes.Compare(ip, network.end) <= 0 {
			return true
		}
	}
	return false
}
