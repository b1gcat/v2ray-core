package cfgagent

import (
	"fmt"
	"testing"
)

func Test_ListAndListenConfigsInNamespace(t *testing.T) {
	ConfigFilePath = "/tmp/elink.conf"

	cfg := &ConfigClient{
		Username:    "nacos",
		Password:    "fucku@2025",
		ServerAddr:  "106.75.239.178",
		NamespaceID: "vps-pool-1",
		GroupID:     "",
		Number:      256,
	}
	cfg.SaveConfig()

	Run()

	fmt.Println("listening...")
	select {}
}
