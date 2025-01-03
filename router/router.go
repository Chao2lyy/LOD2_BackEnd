// router.go
package router

import (
	"LOD2_BE/execute" // 替换为你的实际模块路径
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter 初始化 Gin 路由，并定义所有的路由规则
func SetupRouter(s *execute.SelfObject) *gin.Engine {
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
		execute.ObjExecute(&params, s, c)
	})

	// 定义 /json-execute 路由，接收更复杂的 JSON 配置结构
	r.POST("/block-execute", func(c *gin.Context) {
		// 调用 JsonExecute 函数来处理 JSON 格式的配置
		execute.JsonExecute(c, s)
	})

	r.POST("/infer-execute", func(c *gin.Context) {
		execute.InferRouteHandler(c, s)
	})

	return r
}
