package cfgagent

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/v2fly/v2ray-core/v5/app/proxyman/command"
)

var (
	ConfigFilePath = "/elink/elink.conf"
	// 固定的加密/解密口令
	FixedPassword = "your_fixed_password_here"
)

type tunnelState struct {
	item   model.ConfigItem
	inUsed bool
}

type userState struct {
	item *tunnelState

	ip  string
	mac string
}

type ConfigClient struct {
	// Nacos 用户名
	Username string
	// Nacos 密码
	Password string
	// Nacos 服务器地址
	ServerAddr string
	// Nacos 命名空间ID
	NamespaceID string
	// Nacos 配置分组ID
	GroupID string
	// 配置项数量
	Number int
	// 配置更新间隔时间（秒）
	Interval int

	//context
	context context.Context
	cancel  context.CancelFunc

	//nacos client
	client config_client.IConfigClient

	//v2ray client
	elink command.HandlerServiceClient

	//tunnels
	tunnels sync.Map

	//users
	users sync.Map

	//lock for new/del/mod user
	lock sync.Mutex
}

// SaveConfig 保存配置到文件
func (c *ConfigClient) SaveConfig() error {
	// 备份原配置文件
	if _, err := os.Stat(ConfigFilePath); err == nil {
		backupPath := ConfigFilePath + ".bak"
		if err := copyFile(ConfigFilePath, backupPath); err != nil {
			return fmt.Errorf("备份配置文件失败: %w", err)
		}
	}

	// 移除动态输入加密口令的代码
	// 生成加密密钥
	key := []byte(FixedPassword)
	if len(key) < aes.BlockSize {
		return fmt.Errorf("口令长度不能小于 %d 字节", aes.BlockSize)
	}
	block, err := aes.NewCipher(key[:aes.BlockSize])
	if err != nil {
		return err
	}

	// 序列化配置
	configData, err := json.Marshal(c)
	if err != nil {
		return err
	}

	// 加密配置数据
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	ciphertext := gcm.Seal(nonce, nonce, configData, nil)

	// 编码为 Base64
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	// 保存到文件
	file, err := os.Create(ConfigFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(encoded)
	if err != nil {
		return err
	}

	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

// LoadConfig 从文件加载配置
func (c *ConfigClient) LoadConfig() error {
	// 移除动态输入解密口令的代码
	// 生成解密密钥
	key := []byte(FixedPassword)
	if len(key) < aes.BlockSize {
		return fmt.Errorf("口令长度不能小于 %d 字节", aes.BlockSize)
	}
	block, err := aes.NewCipher(key[:aes.BlockSize])
	if err != nil {
		return err
	}

	// 读取文件内容
	file, err := os.Open(ConfigFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var encoded string
	for scanner.Scan() {
		encoded += scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// 解码 Base64
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}

	// 解密配置数据
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	// 反序列化配置
	err = json.Unmarshal(plaintext, c)
	if err != nil {
		return err
	}

	return nil
}
