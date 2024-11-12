package ips

import (
	"fmt"
	"net"
)

// GetIP 获取当前机器的非回环IP地址。
// 该函数尝试枚举所有网络接口并筛选出一个有效的非回环IP地址。
// 返回值:
//   addr string - 有效的非回环IP地址字符串，如果无法获取则为空字符串。
func GetIP() (addr string) {
	// 获取所有网络接口信息
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("获取网卡信息出错", err)
		return
	}
	// 遍历所有网络接口
	for _, iface := range interfaces {
		// 获取接口上的所有地址
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println("获取IP地址出错", err)
			continue
		}
		// 遍历所有地址
		for _, addr := range addrs {
			// 尝试将地址转换为IP网段类型
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				// 检查是否为IPv4地址
				if ipnet.IP.To4() != nil {
					// 返回第一个有效的非回环IPv4地址
					return ipnet.IP.String()
				}
			}
		}
	}
	// 如果没有找到有效的IP地址，则返回空字符串
	return
}
