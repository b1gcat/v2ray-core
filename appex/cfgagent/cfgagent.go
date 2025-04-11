package cfgagent

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/v2fly/v2ray-core/v5/appex/arp"
)

var (
	cfgClient = &ConfigClient{}
)

func Run() {
	cfgClient.context, cfgClient.cancel =
		context.WithCancel(context.Background())

	restart()
}

func restart() {
	logrus.Info("Starting...")

	cfgClient.LoadConfig()

	cfgClient.context, cfgClient.cancel =
		context.WithCancel(context.Background())

	runRule(cfgClient.context)

	cfgClient.elink = initElinikClient()

	go arp.StartARPListener(func(event arp.ARPEvent) {
		cfgClient.handleARP(event)
	})

	if err := cfgClient.newClient(); err != nil {
		logrus.Error("Failed to create new client: ", err)
		return
	}

	// 启动定期执行任务
	ticker := time.NewTicker(time.Duration(cfgClient.Interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := cfgClient.hotLoadConfig(); err != nil {
				logrus.Error("Failed to hot load config: ", err)
			}
		case <-cfgClient.context.Done():
			logrus.Info("Periodic hot load config stopped.")
			return
		}
	}
}
