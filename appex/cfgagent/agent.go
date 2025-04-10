package cfgagent

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/sirupsen/logrus"
)

func (cfg *ConfigClient) newClient() error {
	// 创建客户端配置
	clientConfig := *constant.NewClientConfig(
		constant.WithNamespaceId(cfg.NamespaceID),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
		constant.WithUsername(cfg.Username),
		constant.WithPassword(cfg.Password),
	)

	// 创建服务端配置
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: cfg.ServerAddr,
			Port:   8848,
		},
	}

	// 创建配置客户端
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create config client: %w", err)
	}

	cfg.client = configClient

	return nil
}

func (cfg *ConfigClient) hotLoadConfig() error {
	// 列出指定命名空间下的所有配置
	listConfig, err := cfg.client.SearchConfig(vo.SearchConfigParam{
		Search:   "blur",
		PageNo:   1,
		PageSize: cfg.Number,
		DataId:   "",
		Group:    cfg.GroupID,
	})
	if err != nil {
		return fmt.Errorf("failed to search config: %w", err)
	}

	// 记录当前存在的配置的 DataId
	for _, item := range listConfig.PageItems {

		// 监听配置变化
		// 检查是否已经存在配置隧道
		if _, ok := cfg.Private.Load(item.DataId); ok {
			continue
		}
		if err = cfg.client.ListenConfig(vo.ConfigParam{
			DataId: item.DataId,
			Group:  item.Group,
			OnChange: func(namespace, group, dataId, data string) {

				if data == "" {
					// 删除配置隧道
					cfg.Private.Delete(dataId)
					// 释放资源
					cfg.client.CancelListenConfig(vo.ConfigParam{
						DataId: dataId,
						Group:  group,
					})
					logrus.Infof("Tunnel deleted:namespace %s group %s, zdataId: %s",
						namespace, group, dataId)
					return
				}

				logrus.Infof("Config changed: namespace %s group %s, zdataId: %s",
					namespace, group, dataId)
			},
		}); err != nil {
			return fmt.Errorf("failed to listen config: %w", err)
		}
		// 新增配置隧道
		cfg.Private.Store(item.DataId, &cfgState{
			item:   item,
			inUsed: false,
		})
		logrus.Infof("New tunnel: namespace %s group %s, dataId: %s",
			cfg.NamespaceID, item.Group, item.DataId)
	}

	return nil
}
