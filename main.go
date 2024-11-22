// main.go
package main

import (
	"LOD2_BE/execute"
	"LOD2_BE/router" // 替换为你的实际模块路径
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 创建 SelfObject 实例并初始化 cliConf 字段
	s := &execute.SelfObject{
		CliConf: &execute.ClientConfig{},
	}

	// 创建 SSH 连接
	err := s.CliConf.CreateClient(execute.IP, execute.Port, execute.User, execute.Password)
	if err != nil {
		fmt.Println("SSH连接失败:", err)

		return
	}
	// 捕获系统信号，确保优雅关闭
	cleanupOnInterrupt(s)

	// 设置路由
	r := router.SetupRouter(s)

	// 启动服务器，监听 8080 端口
	fmt.Println("服务器启动，监听端口8080")
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
		s.CleanupRemoteAll()

	}
}

// 捕获中断信号，并在程序退出时清理远程目录
func cleanupOnInterrupt(s *execute.SelfObject) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c // 等待信号
		fmt.Println("收到中断信号，开始清理远程目录...")
		s.CleanupRemoteAll() // 调用远程清理函数
		os.Exit(1)           // 退出程序
	}()
}
