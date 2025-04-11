package cfgagent

import (
	"embed"
	"os"
	"path/filepath"
)

//go:embed cn2cn/*.json
var cn2cnFiles embed.FS

func mkElinkCn2Cn() {
	// 创建目标目录
	targetDir := "/tmp/elink/cn2cn"
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		panic(err)
	}

	// 遍历嵌入的文件
	entries, err := cn2cnFiles.ReadDir("cn2cn")
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			// 读取嵌入的文件内容
			fileContent, err := cn2cnFiles.ReadFile("cn2cn/" + entry.Name())
			if err != nil {
				panic(err)
			}

			// 写入目标文件
			targetFilePath := filepath.Join(targetDir, entry.Name())
			targetFile, err := os.Create(targetFilePath)
			if err != nil {
				panic(err)
			}
			defer targetFile.Close()

			// 修复部分：直接使用 io.Writer.Write 方法写入 []byte 数据
			_, err = targetFile.Write(fileContent)
			if err != nil {
				panic(err)
			}
		}
	}
}
