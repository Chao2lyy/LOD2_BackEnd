package execute

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// JsonExecute 处理前端传来的 JSON 配置，上传文件并执行命令
func JsonExecute(c *gin.Context, s *SelfObject) {
	var jsonData map[string]interface{}
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法解析请求的 JSON", "details": err.Error()})
		return
	}

	// 提取 input_dir 和 output_dir，并转换为 Unix 格式
	inputDirWin := jsonData["input_dir"].(string)
	outputDirWin := jsonData["output_dir"].(string)
	inputDirUnix := formatWindowsPathToUnix(inputDirWin)
	outputDirUnix := formatWindowsPathToUnix(outputDirWin)

	// 构建远程路径
	remoteInputDir := filepath.ToSlash(filepath.Join(LOD2Path, "input_temp"))
	remoteOutputDir := filepath.ToSlash(filepath.Join(LOD2Path, "output_temp"))

	// 修改远程的 config_modeling.xml 文件
	remoteConfigFile := filepath.ToSlash(filepath.Join(LOD2Path, "configSL", "config_modeling.xml"))
	if err := modifyRemoteConfigFile(s, jsonData, remoteConfigFile); err != nil {
		handleError(c, "修改远程 config_modeling.xml 失败", err)
		//s.CleanupRemote4()
		return
	}

	// 上传本地 input_dir 和 output_dir 到远程服务器
	if err := s.UploadLocalFolderToRemote(inputDirUnix, remoteInputDir, 4); err != nil {
		handleError(c, "上传本地输入文件夹失败", err)
		s.CleanupRemote4()
		return
	}
	if err := s.UploadLocalFolderToRemote(outputDirUnix, remoteOutputDir, 4); err != nil {
		handleError(c, "上传本地输出文件夹失败", err)
		s.CleanupRemote4()
		return
	}

	// 构建执行命令
	command := fmt.Sprintf(`
    bash -c '
    Xvfb :2 -screen 0 1024x768x24 & 
    export DISPLAY=:2 && 
    export LD_LIBRARY_PATH=%slib && 
    cd %s && 
    ./block_modeling configSL
    '`, LOD2Path, LOD2Path)

	// 执行命令
	if err := s.RunCommand(command); err != nil {
		handleError(c, "命令执行失败", err)
		s.CleanupRemote4()
		return
	}

	// 下载远程 output_dir 到本地
	localOutputDir := outputDirUnix
	if err := s.DownloadRemoteFolderToLocal(remoteOutputDir, localOutputDir); err != nil {
		handleError(c, "下载远程文件夹失败", err)
		s.CleanupRemote4()
		return
	}

	// 返回成功消息和下载路径
	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"message":         "命令执行成功，结果已下载到本地",
		"local_save_path": localOutputDir,
	})
	fmt.Println("任务成功完成，开始清理...")
	s.CleanupRemote4() // 调用清理函数
}

// modifyRemoteConfigFile 修改远程的 config_modeling.xml 文件
func modifyRemoteConfigFile(s *SelfObject, jsonData map[string]interface{}, remoteConfigFile string) error {
	// 将 JSON 数据转换为 XML 结构并写入远程的 config_modeling.xml 文件
	configXML, err := jsonToXML(jsonData)
	if err != nil {
		return fmt.Errorf("无法将 JSON 转换为 XML: %v", err)
	}

	// 上传新的 config_modeling.xml 文件到远程
	if err := s.createRemoteConfigFile4(map[string]string{"config": configXML}, remoteConfigFile); err != nil {
		return fmt.Errorf("无法修改远程配置文件: %v", err)
	}

	return nil
}

func handleError(c *gin.Context, msg string, err error) {
	fmt.Printf("%s: %v\n", msg, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": msg, "details": err.Error()})
}
