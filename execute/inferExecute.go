package execute

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func InferRouteHandler(c *gin.Context, s *SelfObject) {
	// 接收前端传递的 JSON 参数
	var jsonData map[string]interface{}
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数解析失败", "details": err.Error()})
		return
	}

	// 获取除 training_folder 外的所有参数
	savePathWin := jsonData["save_path"].(string)
	dataRootWin := jsonData["data_root"].(string)
	dataset := jsonData["dataset"].(string)
	nclasses := int(jsonData["nclasses"].(float64))
	batchSizeVal := int(jsonData["batch_size_val"].(float64))
	patchSize := int(jsonData["patch_size"].(float64))
	gpu := int(jsonData["gpu"].(float64))
	modelName := jsonData["model_name"].(string)
	pretrainedModel := jsonData["pretrained_model"].(string)

	// 转换 Windows 路径为 Unix 格式
	savePathUnix := formatWindowsPathToUnix(savePathWin)
	//dataRootUnix := formatWindowsPathToUnix(dataRootWin)

	// 构建远程服务器的临时文件夹路径
	remoteSavePath := filepath.ToSlash(filepath.Join(LOD2Path, "save2_temp"))
	remoteDataRoot := filepath.ToSlash(filepath.Join(LOD2Path, "data2_temp"))

	// 上传本地 save_path 和 data_root 到远程服务器
	if err := s.UploadLocalFolderToRemote(savePathWin, remoteSavePath, 1); err != nil {
		handleError(c, "上传本地保存路径失败", err)
		s.CleanupRemote2()
		return
	}
	if err := s.UploadLocalFolderToRemote(dataRootWin, remoteDataRoot, 1); err != nil {
		handleError(c, "上传本地数据路径失败", err)
		s.CleanupRemote2()
		return
	}

	// 构建执行命令
	command := fmt.Sprintf(`
		bash -c '
		export LD_LIBRARY_PATH=%slib && cd %s && ./infer_crop_after_2D_sensaturban \
		--training_folder /home/hc/LOD2exe/log_SensatUrban_DeepLabV3Plus_depth_combine_resnet50-512_8_aug/ \
		--save_path %s --data_root %s --dataset %s --nclasses %d --batch_size_val %d --patch_size %d --gpu %d \
		--model_name %s --pretrained_model %s
		'`, LOD2Path, LOD2Path, remoteSavePath, remoteDataRoot, dataset, nclasses, batchSizeVal, patchSize, gpu, modelName, pretrainedModel)

	// 执行命令
	if err := s.RunCommand(command); err != nil {
		handleError(c, "命令执行失败", err)
		s.CleanupRemote2()
		return
	}

	// 下载远程服务器上的 save_path 到本地
	localSavePath := savePathUnix
	if err := s.DownloadRemoteFolderToLocal(remoteSavePath, localSavePath); err != nil {
		handleError(c, "下载远程保存路径失败", err)
		s.CleanupRemote2()
		return
	}

	// 返回成功消息和下载路径
	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"message":         "命令执行成功，结果已下载到本地",
		"local_save_path": localSavePath,
	})

	// 清理远程临时文件夹
	s.CleanupRemote2()
}
