package cfgagent

import (
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/v2fly/v2ray-core/v5/appex/arp"
)

func (cfg *ConfigClient) handleARP(event arp.ARPEvent) {
	logrus.Info("Handling ARP event:", event)

	switch event.Type {
	case arp.NEW:
		if err := cfg.newUser(event); err != nil {
			logrus.Error("Failed to handle new user: ", err)
		}
	case arp.DELETE:
		cfg.locker.Lock()
		cfg._delUser(event)
		cfg.locker.Unlock()
	default:
		logrus.Warn("Unknown ARP event type:", event.Type)
	}
}

func (cfg *ConfigClient) newUser(event arp.ARPEvent) error {
	// 检查用户是否已存在
	user := &userState{
		ip:  event.SenderIP,
		mac: event.SenderMAC,
	}
	cfg.locker.Lock()
	defer cfg.locker.Unlock()

	if oldUser, exists := cfg.users.LoadOrStore(event.SenderMAC, &user); exists {
		if oldUser.(*userState).ip == event.SenderIP {
			// 用户信息未变化，不执行任何操作
			return nil
		}
		if err := cfg._delUser(event); err != nil {
			return err
		}
	}

	cfg.users.Store(event.SenderMAC, user)
	logrus.Info("New user:", event.SenderMAC)
	return nil
}

func (cfg *ConfigClient) _delUser(event arp.ARPEvent) error {
	cfg.users.Delete(event.SenderMAC)
	logrus.Info("Del user:", event.SenderMAC)
	return nil
}

func ip2Port(ip string) int {
	// 获取IP地址的最后一部分
	lastDotIndex := strings.LastIndex(ip, ".")
	if lastDotIndex == -1 || lastDotIndex >= len(ip)-1 {
		return 1000 // 如果IP格式不正确，返回默认端口
	}
	lastPart := ip[lastDotIndex+1:]
	port, err := strconv.Atoi(lastPart)
	if err != nil {
		return 1000 // 如果转换失败，返回默认端口
	}
	return 1000 + port
}
