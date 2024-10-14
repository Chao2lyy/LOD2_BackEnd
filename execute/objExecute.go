// objExecute.go
package execute

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// MergeParameters 合并 JSON 参数和默认参数
func MergeParameters(params *RequestParams) map[string]string {
	mergedConfig := make(map[string]string)
	for k, v := range DefaultConfig {
		mergedConfig[k] = v // 复制默认参数
	}

	// 将 JSON 参数中的值覆盖到默认参数中
	if params.DataPath != "" {
		mergedConfig["data_path"] = params.DataPath
	}
	if params.SavePath != "" {
		mergedConfig["save_path"] = params.SavePath
	}
	if params.LogName != "" {
		mergedConfig["log_name"] = params.LogName
	}
	if params.ColorPath != "" {
		mergedConfig["color_path"] = params.ColorPath
	}
	if params.Threads != "" {
		mergedConfig["threads"] = params.Threads
	}
	if params.Scale != "" {
		mergedConfig["scale"] = params.Scale
	}
	if params.LabelNum != "" {
		mergedConfig["labelnum"] = params.LabelNum
	}
	if params.BRenderImage != "" {
		mergedConfig["bRenderImage"] = params.BRenderImage
	}
	if params.BBackProjectImage != "" {
		mergedConfig["bBackProjectImage"] = params.BBackProjectImage
	}

	return mergedConfig
}

// SaveConfigToFile 保存合并后的参数到配置文件中
func SaveConfigToFile(config map[string]string, filePath string) error {
	lines := ""
	for key, value := range config {
		lines += fmt.Sprintf("%s=%s\n", key, value)
	}
	return ioutil.WriteFile(filePath, []byte(lines), 0644)
}

// ObjExecute 执行对象命令，并将结果文件夹传递给前端
func ObjExecute(params *RequestParams, c *gin.Context) {
	// 合并默认参数和前端传递的参数
	mergedConfig := MergeParameters(params)

	// 将前端传入的 Windows 路径转换为 Unix 格式
	dataPathWin := mergedConfig["data_path"]
	savePathWin := mergedConfig["save_path"]

	// 构建远程服务器上的文件夹路径
	remoteDataPath := filepath.Join(LOD2Path, "data_temp") // 临时数据文件夹
	remoteSavePath := filepath.Join(LOD2Path, "save_temp") // 临时保存文件夹
	remoteDataPath = filepath.ToSlash(remoteDataPath)
	remoteSavePath = filepath.ToSlash(remoteSavePath)
	// 创建 SelfObject 实例并初始化 cliConf 字段
	s := &SelfObject{
		cliConf: &ClientConfig{},
	}

	// 创建 SSH 连接
	err := s.cliConf.CreateClient(IP, Port, User, Password)
	if err != nil {
		fmt.Println("SSH连接失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSH连接失败"})
		return
	}

	// 上传本地文件夹到远程服务器
	if err := s.UploadLocalFolderToRemote(dataPathWin, remoteDataPath); err != nil {
		fmt.Println("上传本地文件夹失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传本地文件夹失败"})
		return
	}

	// 更新配置文件中的 data_path 和 save_path 为远程路径
	mergedConfig["data_path"] = remoteDataPath
	mergedConfig["save_path"] = remoteSavePath

	// 构建远程服务器上的配置文件路径（存放在 configSL 目录中）
	remoteTempFilePath := filepath.Join(LOD2Path, "configSL", "objRender_temp.txt")

	// 在远程服务器上创建配置文件
	remoteTempFilePath = filepath.ToSlash(remoteTempFilePath)
	err = s.createRemoteConfigFile(mergedConfig, remoteTempFilePath)
	if err != nil {
		fmt.Println("在远程服务器上创建配置文件失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "在远程服务器上创建配置文件失败"})
		return
	}
	// 在远程服务器上的 save_path 路径中创建 depths、results、rgb 文件夹
	err = s.createRemoteFolders(remoteSavePath)
	if err != nil {
		fmt.Println("创建远程文件夹失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建远程文件夹失败"})
		return
	}

	// 构建执行命令
	command := fmt.Sprintf("bash -c 'export LD_LIBRARY_PATH=%slib && cd %s && ./3DMapSL_obj %s'", LOD2Path, LOD2Path, remoteTempFilePath)

	// 执行命令
	if err := s.RunCommand(command); err != nil {
		fmt.Println("运行命令出错:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "命令执行失败"})
		return
	}

	// 下载远程服务器上的结果文件夹到本地
	localSavePath := savePathWin
	err = s.DownloadRemoteFolderToLocal(remoteSavePath, localSavePath)
	if err != nil {
		fmt.Println("下载远程文件夹失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "下载远程文件夹失败"})
		return
	}

	// 将下载的结果文件夹传递给前端
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "命令执行成功，结果已下载到本地", "local_save_path": localSavePath})
}

// formatWindowsPathToUnix 将 Windows 路径格式转换为 Unix 格式
func formatWindowsPathToUnix(windowsPath string) string {
	return strings.ReplaceAll(windowsPath, "\\", "/")
}

// createRemoteFolders 在远程服务器上创建 depths、results、rgb 文件夹
func (s *SelfObject) createRemoteFolders(remoteSavePath string) error {
	// 构建需要创建的文件夹路径
	remoteDepthsPath := filepath.ToSlash(filepath.Join(remoteSavePath, "depths"))
	remoteResultsPath := filepath.ToSlash(filepath.Join(remoteSavePath, "results"))
	remoteRGBPath := filepath.ToSlash(filepath.Join(remoteSavePath, "rgb"))

	// 打印调试信息
	fmt.Printf("远程文件夹路径：%s, %s, %s\n", remoteDepthsPath, remoteResultsPath, remoteRGBPath)

	// 创建远程目录命令
	createFoldersCommand := fmt.Sprintf("mkdir -p %s %s %s", remoteDepthsPath, remoteResultsPath, remoteRGBPath)

	// 运行命令创建目录
	if err := s.RunCommand(createFoldersCommand); err != nil {
		return fmt.Errorf("创建远程文件夹失败: %v", err)
	}

	fmt.Println("远程文件夹创建成功")
	return nil
}
