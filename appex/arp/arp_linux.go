//go:build linux

package arp

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// StartARPListener listens for ARP events and invokes the provided callback function.
func StartARPListener(callback func(event ARPEvent)) error {
	// 创建 Netlink 订阅
	ch := make(chan netlink.NeighUpdate)
	done := make(chan struct{})
	defer close(done)

	// 添加错误处理
	if err := FlushAllNeighbors(); err != nil {
		logrus.Errorf("failed to flush ARP cache: %v", err)
	}

	if err := netlink.NeighSubscribe(ch, done); err != nil {
		return fmt.Errorf("failed to subscribe to neigh updates: %v", err)
	}

	go func() {
		for update := range ch {
			event := ARPEvent{}

			if update.Neigh.IP != nil {
				event.SenderIP = update.Neigh.IP.String()
			}
			if update.Neigh.HardwareAddr != nil {
				event.SenderMAC = update.Neigh.HardwareAddr.String()
			}

			switch update.Type {
			case unix.RTM_NEWNEIGH:
				event.Type = NEW
				callback(event)
			case unix.RTM_DELNEIGH:
				event.Type = DELETE
				callback(event)
			}
		}
	}()

	return nil
}

// FlushAllNeighbors 清除所有网络接口的 ARP 缓存
func FlushAllNeighbors() error {
	// 获取所有网络接口
	links, err := netlink.LinkList()
	if err != nil {
		return fmt.Errorf("failed to get network interfaces: %v", err)
	}

	// 遍历所有接口并删除 ARP 条目
	for _, link := range links {
		// 获取接口的所有 ARP 条目
		neighs, err := netlink.NeighList(link.Attrs().Index, netlink.FAMILY_ALL)
		if err != nil {
			continue
		}

		// 删除每个 ARP 条目
		for _, neigh := range neighs {
			if err := netlink.NeighDel(&neigh); err != nil {
				continue
			}
		}
	}

	return nil
}
