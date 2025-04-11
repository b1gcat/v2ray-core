package cfgagent

import (
	"context"
	"fmt"
	"time"

	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/main/commands/all/api"
	"github.com/v2fly/v2ray-core/v5/main/commands/helpers"
	"google.golang.org/grpc"
)

// AddInbound 添加inbound配置
func (cfg *ConfigClient) addInbound(config *command.AddInboundRequest) error {
	_, err := cfg.elink.AddInbound(context.Background(), config)
	return err
}

// AddOutbound 添加outbound配置

func (cfg *ConfigClient) addOutbound(client command.HandlerServiceClient, config *command.AddOutboundRequest) error {
	_, err := cfg.elink.AddOutbound(context.Background(), config)
	return err
}

// RemoveInbound 删除inbound配置
func (cfg *ConfigClient) removeInbound(tag string) error {
	_, err := cfg.elink.RemoveInbound(context.Background(), &command.RemoveInboundRequest{
		Tag: tag,
	})
	return err
}

// RemoveOutbound 删除outbound配置
func (cfg *ConfigClient) removeOutbound(tag string) error {
	_, err := cfg.elink.RemoveOutbound(context.Background(), &command.RemoveOutboundRequest{
		Tag: tag,
	})
	return err
}

// CreateSocksInboundConfig 创建Socks5 inbound配置
func (cfg *ConfigClient) createSocksInboundConfig(ctx context.Context, in string) error {
	template := []string{in}
	c, err := helpers.LoadConfig(template, api.APIConfigFormat, false)
	if err != nil {
		return err
	}

	for _, out := range c.InboundConfigs {
		o, err := out.Build()
		if err != nil {
			return fmt.Errorf("failed to build conf: %s", err)
		}
		r := &command.AddInboundRequest{
			Inbound: o,
		}
		_, err = cfg.elink.AddInbound(ctx, r)
		if err != nil {
			return fmt.Errorf("failed to build conf: %s", err)
		}
	}
	return nil
}

// CreateSocksOutboundConfig 创建Socks5 outbound配置
func (cfg *ConfigClient) createSocksOutboundConfig(ctx context.Context, out string) error {
	template := []string{out}
	c, err := helpers.LoadConfig(template, api.APIConfigFormat, false)
	if err != nil {
		return err
	}

	for _, out := range c.OutboundConfigs {
		o, err := out.Build()
		if err != nil {
			return fmt.Errorf("failed to build conf: %s", err)
		}
		r := &command.AddOutboundRequest{
			Outbound: o,
		}
		_, err = cfg.elink.AddOutbound(ctx, r)
		if err != nil {
			return fmt.Errorf("failed to build conf: %s", err)
		}
	}
	return nil
}

// initElinikClient 初始化HandlerServiceClient
func initElinikClient() command.HandlerServiceClient {
	// 杀掉当前的elink进程
	killCmd := exec.Command("pkill", "elink")
	err := killCmd.Run()
	if err != nil {
		logrus.Warnf("杀掉elink进程失败: %v", err)
	}

	// 等待一段时间以确保进程完全关闭
	time.Sleep(2 * time.Second)

	mkElinkCn2Cn()

	// 启动elink进程
	startCmd := exec.Command("elink", "run", "-d", "/tmp/elink/cn2cn")
	err = startCmd.Start()
	if err != nil {
		logrus.Fatal("启动elink进程失败: %w", err)
	}

	// 重新建立gRPC连接
	var conn *grpc.ClientConn
	for {
		conn, err = grpc.Dial("localhost:10086", grpc.WithInsecure())
		if err == nil {
			break
		}
		logrus.Warnf("连接失败，正在重试: %v", err)
		time.Sleep(5 * time.Second) // 等待5秒后重试
	}

	logrus.Info("成功连接到gRPC服务")
	return command.NewHandlerServiceClient(conn)
}
