// router.go
package router

import (
	"LOD2_BE/execute" // 替换为你的实际模块路径
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter 初始化 Gin 路由，并定义所有的路由规则
func SetupRouter() *gin.Engine {
	// 创建 Gin 路由实例
	r := gin.Default()

	// 定义 /execute 路由，接收前端传递的 JSON 参数
	r.POST("/execute", func(c *gin.Context) {
		// 解析前端传递的 JSON 参数
		var params execute.RequestParams
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数解析失败", "details": err.Error()})
			return
		}

		// 执行对象命令，并将结果文件传递给前端
		execute.ObjExecute(&params, c)
	})

	return r
}
