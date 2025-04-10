package cfgagent

import (
	"context"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func runRule(ctx context.Context) {
	// 执行 basicRuleFramwork 中的命令
	executeBasicRuleFramework()
	// 启动网段监听
	go listenLocalNetworkChanged(ctx)
	// 保持程序运行
}

// 存储上一次的网络接口网段信息
var lastNetworkSegments []string

func listenLocalNetworkChanged(ctx context.Context) {
	// 初始化 lastNetworkSegments
	lastNetworkSegments = getNetworkSegments()

	// 启动一个定时任务，每隔一段时间检查一次网络接口网段信息
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
		lastNetworkSegments = make([]string, 0)
	}()

	for {
		select {
		case <-ticker.C:
			// 获取当前的网络接口网段信息
			currentNetworkSegments := getNetworkSegments()

			// 检查网络接口网段信息是否有变化
			if hasNetworkSegmentChanged(currentNetworkSegments) {
				logrus.Info("Network segment has changed!")
				// 清空 V2RAY_LOCAL_NETWORK
				clearIPSet()
				// 把最新的网段加进去
				addNewSegmentsToIPSet(currentNetworkSegments)
			}

			// 更新 lastNetworkSegments
			lastNetworkSegments = currentNetworkSegments
		case <-ctx.Done():
			// 当上下文被取消时，退出循环
			logrus.Info("Stopping network segment monitoring due to context cancellation.")
			return
		}
	}
}

// 执行 basicRuleFramwork 中的命令
func executeBasicRuleFramework() {
	// 拆分命令
	commands := strings.Split(basicRuleFramwork, "\n")
	for _, cmdStr := range commands {
		// 去除前后空格
		cmdStr = strings.TrimSpace(cmdStr)
		// 跳过注释和空行
		if strings.HasPrefix(cmdStr, "#") || cmdStr == "" {
			continue
		}
		// 替换环境变量
		cmdStr = strings.ReplaceAll(cmdStr, "${MARK}", "0Xff")
		// 拆分命令和参数
		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			continue
		}
		cmdName := parts[0]
		cmdArgs := parts[1:]

		// 创建命令对象
		cmd := exec.Command(cmdName, cmdArgs...)
		// 执行命令
		output, err := cmd.CombinedOutput()
		if err != nil {
			logrus.Errorf("Failed to execute command %s: %v, output: %s", cmdStr, err, string(output))
		} else {
			logrus.Infof("Successfully executed command: %s", cmdStr)
		}
	}
}

// 清空 V2RAY_LOCAL_NETWORK
func clearIPSet() {
	logrus.Info("Clearing V2RAY_LOCAL_NETWORK ipset...")
	// 执行清空 ipset 的命令
	cmd := exec.Command("ipset", "flush", "V2RAY_LOCAL_NETWORK")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("Failed to clear V2RAY_LOCAL_NETWORK ipset: %v, output: %s", err, string(output))
	} else {
		logrus.Info("Successfully cleared V2RAY_LOCAL_NETWORK ipset")
	}
}

// 把最新的网段加进去
func addNewSegmentsToIPSet(segments []string) {
	logrus.Info("Adding new network segments to V2RAY_LOCAL_NETWORK ipset...")
	for _, segment := range segments {
		logrus.Infof("Adding %s to V2RAY_LOCAL_NETWORK ipset...", segment)
		// 执行添加网段到 ipset 的命令
		cmd := exec.Command("ipset", "add", "V2RAY_LOCAL_NETWORK", segment)
		output, err := cmd.CombinedOutput()
		if err != nil {
			logrus.Errorf("Failed to add %s to V2RAY_LOCAL_NETWORK ipset: %v, output: %s", segment, err, string(output))
		} else {
			logrus.Infof("Successfully added %s to V2RAY_LOCAL_NETWORK ipset", segment)
		}
	}
}

// 获取当前网络接口的网段信息
func getNetworkSegments() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		logrus.Errorf("Failed to get network interfaces: %v", err)
		return nil
	}

	var segments []string
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			logrus.Errorf("Failed to get addresses for interface %s: %v", iface.Name, err)
			continue
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				segments = append(segments, ipNet.String())
			}
		}
	}

	return segments
}

// 检查网络接口网段信息是否有变化
func hasNetworkSegmentChanged(currentSegments []string) bool {
	if len(currentSegments) != len(lastNetworkSegments) {
		return true
	}

	// 简单比较网段信息
	for _, currentSeg := range currentSegments {
		found := false
		for _, lastSeg := range lastNetworkSegments {
			if currentSeg == lastSeg {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	return false
}

var (
	basicRuleFramwork = `

MARK=0Xff
ipset", "flush", "V2RAY_LOCAL_NETWORK
ipset create V2RAY_LOCAL_NETWORK hash:net

# policy route
ip rule add fwmark 1 table 100 
ip route add local 0.0.0.0/0 dev lo table 100

# Framwork
## basic
iptables -t mangle -N V2RAY
iptables -t mangle -F V2RAY

iptables -t mangle -A V2RAY -d 127.0.0.1/32 -j RETURN
iptables -t mangle -A V2RAY -d 224.0.0.0/4 -j RETURN 
iptables -t mangle -A V2RAY -d 255.255.255.255/32 -j RETURN
iptables -t mangle -A V2RAY -j RETURN -m mark --mark ${MARK}  

iptables -t mangle -A V2RAY -m set --match-set V2RAY_LOCAL_NETWORK dst -p tcp -j RETURN 
iptables -t mangle -A V2RAY -m set --match-set V2RAY_LOCAL_NETWORK dst -p udp ! --dport 53 -j RETURN 

# apply
iptables -t mangle -A PREROUTING -j V2RAY 
    `
)

// 添加 redirect 规则
func addRedirectRule(redirectPort string) {
	rule := `iptables -t mangle -A V2RAY -p udp -j TPROXY --on-ip 127.0.0.1 --on-port ` + redirectPort + ` --tproxy-mark 1`
	ruleTCP := `iptables -t mangle -A V2RAY -p tcp -j TPROXY --on-ip 127.0.0.1 --on-port ` + redirectPort + ` --tproxy-mark 1`

	// 执行 UDP 规则
	cmd := exec.Command("bash", "-c", rule)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("Failed to add UDP redirect rule: %v, output: %s", err, string(output))
	} else {
		logrus.Infof("Successfully added UDP redirect rule with port %s", redirectPort)
	}

	// 执行 TCP 规则
	cmd = exec.Command("bash", "-c", ruleTCP)
	output, err = cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("Failed to add TCP redirect rule: %v, output: %s", err, string(output))
	} else {
		logrus.Infof("Successfully added TCP redirect rule with port %s", redirectPort)
	}
}

// 删除 redirect 规则
func deleteRedirectRule(redirectPort string) {
	rule := `iptables -t mangle -D V2RAY -p udp -j TPROXY --on-ip 127.0.0.1 --on-port ` + redirectPort + ` --tproxy-mark 1`
	ruleTCP := `iptables -t mangle -D V2RAY -p tcp -j TPROXY --on-ip 127.0.0.1 --on-port ` + redirectPort + ` --tproxy-mark 1`

	// 执行 UDP 规则删除
	cmd := exec.Command("bash", "-c", rule)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("Failed to delete UDP redirect rule: %v, output: %s", err, string(output))
	} else {
		logrus.Infof("Successfully deleted UDP redirect rule with port %s", redirectPort)
	}

	// 执行 TCP 规则删除
	cmd = exec.Command("bash", "-c", ruleTCP)
	output, err = cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("Failed to delete TCP redirect rule: %v, output: %s", err, string(output))
	} else {
		logrus.Infof("Successfully deleted TCP redirect rule with port %s", redirectPort)
	}
}
