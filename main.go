// main.go
package main

import (
	"LOD2_BE/router" // 替换为你的实际模块路径
	"fmt"
)

func main() {
	// 设置路由
	r := router.SetupRouter()

	// 启动服务器，监听 8080 端口
	fmt.Println("服务器启动，监听端口8080")
	r.Run(":8080")
}
