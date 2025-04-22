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
		cfg.lock.Lock()
		cfg._delUser(event)
		cfg.lock.Unlock()
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

	cfg.lock.Lock()
	defer cfg.lock.Unlock()

	if oldUser, exists := cfg.users.LoadOrStore(event.SenderMAC, &user); exists {
		if oldUser.(*userState).ip == event.SenderIP {
			// 用户信息未变化，不执行任何操作
			return nil
		}
		if err := cfg._delUser(event); err != nil {
			return err
		}
	}

	allocOk := false

	cfg.tunnels.Range(func(key, value any) bool {
		item := key.(*tunnelState)

		if item.inUsed {
			return true
		}

		item.inUsed = true
		user.item = item

		cfg.users.Store(event.SenderMAC, user)

		allocOk = true

		return false
	})

	if allocOk {
		logrus.Info("New user:", event.SenderMAC)
	} else {
		logrus.Warn("New user, but no tunnel available:", event.SenderMAC)
	}

	return nil
}

func (cfg *ConfigClient) _delUser(event arp.ARPEvent) error {
	cfg.users.Delete(event.SenderMAC)
	logrus.Info("Del user:", event.SenderMAC)
	return nil
}

func (cfg *ConfigClient) modUser(nameSpace, dataId string) error {

	cfg.lock.Lock()
	defer cfg.lock.Unlock()

	cfg.users.Range(func(key, value any) bool {
		user := value.(*userState)
		if user.item.item.DataId == dataId {
			user.item.inUsed = false

			cfg._delUser(arp.ARPEvent{SenderMAC: user.mac, SenderIP: user.ip})

			logrus.Info("Mod user:", key)
			return false
		}
		return true
	})

	// 检查用户是否已存在
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
