package command

import (
	"os"
	"path/filepath"
	"strings"
)

// altSignalFilePath 返回 IDEA 调试插件（ms-localdebug）约定的停止信号备用路径：
// ~/.localdebug/signal/<mainSignalFile 的文件名>。
// 插件把停止信号写到该目录，ktctl 同时监听主路径（系统临时目录）与该备用路径，两种停止方式都生效。
func altSignalFilePath(mainSignalFile string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".localdebug", "signal", filepath.Base(mainSignalFile))
}

// isStopSignal 判断指定路径的信号文件内容是否为 "stop"（空路径视为未触发）
func isStopSignal(path string) bool {
	if path == "" {
		return false
	}
	content, err := os.ReadFile(path)
	return err == nil && strings.TrimSpace(string(content)) == "stop"
}
