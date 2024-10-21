// fileTransfer.go
package execute

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/sftp"
)

const (
	dataTempPath   = LOD2Path + "data_temp"   // 临时数据路径
	saveTempPath   = LOD2Path + "save_temp"   // 临时保存路径
	inputTempPath  = LOD2Path + "input_temp"  // 临时数据路径
	outputTempPath = LOD2Path + "output_temp" // 临时保存路径
)

// createSFTPClient 创建 SFTP 客户端
func (s *SelfObject) createSFTPClient() (*sftp.Client, error) {
	// 使用现有 SSH 客户端创建 SFTP 客户端
	client, err := sftp.NewClient(s.CliConf.Client)
	if err != nil {
		return nil, fmt.Errorf("无法创建 SFTP 客户端: %v", err)
	}
	return client, nil
}

// UploadLocalFolderToRemote 上传本地文件夹到远程服务器的指定目录
func (s *SelfObject) UploadLocalFolderToRemote(localFolderPath, remoteFolderPath string, part int) error {
	fmt.Println(localFolderPath, remoteFolderPath)
	client, err := s.createSFTPClient()
	if err != nil {
		return fmt.Errorf("无法创建 SFTP 客户端: %v", err)
	}
	defer client.Close()

	// 定义并发限制，控制最多 10 个协程同时上传
	const MaxWorkers = 1000
	fileChan := make(chan string, MaxWorkers) // 文件路径队列
	errChan := make(chan error, 1)            // 错误捕获通道
	var wg sync.WaitGroup
	// 启动 worker pool
	for i := 0; i < MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for localPath := range fileChan {
				var remotePath string
				if part == 1 || part == 3 {
					remotePath = filepath.Join(remoteFolderPath, strings.TrimPrefix(localPath, localFolderPath))
				} else if part == 4 {
					remotePath = filepath.Join(remoteFolderPath, filepath.Base(localPath))
				}
				remotePath = filepath.ToSlash(remotePath) // POSIX 路径格式
				//fmt.Println(remotePath)
				if err := s.uploadFile(client, localPath, remotePath); err != nil {
					select {
					case errChan <- err: // 只传递第一个错误，避免阻塞
					default:
					}
					return
				}
			}
		}()
	}

	// 遍历本地文件夹，将文件路径发送到 channel
	err = filepath.Walk(localFolderPath, func(localPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(localPath, localFolderPath)
		remotePath := filepath.ToSlash(filepath.Join(remoteFolderPath, relPath))
		if info.IsDir() {
			// 创建远程目录
			if err := client.MkdirAll(remotePath); err != nil {
				return fmt.Errorf("无法创建远程目录: %v", err)
			}
		} else {
			fileChan <- localPath // 将文件路径发送到 channel
		}
		return nil
	})
	if err != nil {
		close(fileChan) // 关闭 channel，避免死锁
		return err
	}

	close(fileChan) // 所有文件路径已发送，关闭 channel
	wg.Wait()       // 等待所有 worker 完成
	close(errChan)  // 关闭错误通道

	// 检查是否有错误发生
	if err, ok := <-errChan; ok {
		return err
	}

	fmt.Printf("本地文件夹成功上传到远程: %s -> %s\n", localFolderPath, remoteFolderPath)
	return nil
}

// uploadFile 单个文件上传逻辑
func (s *SelfObject) uploadFile(client *sftp.Client, localPath, remotePath string) error {
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("无法打开本地文件: %v", err)
	}
	defer localFile.Close()

	remoteFile, err := client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("无法创建远程文件: %v", err)
	}
	defer remoteFile.Close()

	if _, err := localFile.Seek(0, 0); err != nil {
		return err
	}
	if _, err := remoteFile.ReadFrom(localFile); err != nil {
		return fmt.Errorf("文件上传失败: %v", err)
	}

	return nil
}

// DownloadRemoteFolderToLocal 下载远程文件夹到本地指定目录
func (s *SelfObject) DownloadRemoteFolderToLocal(remoteFolderPath, localFolderPath string) error {
	client, err := s.createSFTPClient()
	if err != nil {
		return fmt.Errorf("无法创建 SFTP 客户端: %v", err)
	}
	defer client.Close()

	const MaxWorkers = 1000                      // 控制最大并发数
	fileChan := make(chan [2]string, MaxWorkers) // 存储 (remotePath, localPath) 对的 channel
	errorChan := make(chan error, 1)             // 捕获错误的 channel
	var wg sync.WaitGroup

	// 启动 worker pool
	for i := 0; i < MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for paths := range fileChan {
				remotePath, localPath := paths[0], paths[1]
				if err := s.downloadFile(client, remotePath, localPath); err != nil {
					select {
					case errorChan <- err: // 捕获第一个错误
					default:
					}
					return
				}
			}
		}()
	}

	// 遍历远程文件夹，将文件路径发送到 channel
	err = s.walkRemoteFolder(client, remoteFolderPath, localFolderPath, fileChan)
	if err != nil {
		close(fileChan)
		return err
	}

	close(fileChan) // 所有文件路径已发送，关闭 channel
	wg.Wait()       // 等待所有 worker 完成
	close(errorChan)

	// 检查是否有错误发生
	if err, ok := <-errorChan; ok {
		return err
	}

	fmt.Printf("远程文件夹成功下载到本地: %s -> %s\n", remoteFolderPath, localFolderPath)
	return nil
}

// walkRemoteFolder 遍历远程文件夹，将文件路径对发送到 channel
func (s *SelfObject) walkRemoteFolder(client *sftp.Client, remoteFolderPath, localFolderPath string, fileChan chan<- [2]string) error {
	remoteFiles, err := client.ReadDir(remoteFolderPath)
	if err != nil {
		return fmt.Errorf("无法读取远程目录: %v", err)
	}

	for _, file := range remoteFiles {
		remoteFilePath := filepath.ToSlash(filepath.Join(remoteFolderPath, file.Name()))
		localFilePath := filepath.Join(localFolderPath, file.Name())

		if file.IsDir() {
			// 如果是目录，递归调用
			if err := s.walkRemoteFolder(client, remoteFilePath, localFilePath, fileChan); err != nil {
				return err
			}
		} else {
			// 将文件路径对发送到 channel
			fileChan <- [2]string{remoteFilePath, localFilePath}
		}
	}

	return nil
}

// downloadFile 下载单个文件
func (s *SelfObject) downloadFile(client *sftp.Client, remotePath, localPath string) error {
	remoteFile, err := client.Open(remotePath)
	if err != nil {
		return fmt.Errorf("无法打开远程文件: %v", err)
	}
	defer remoteFile.Close()

	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, os.ModePerm); err != nil {
		return fmt.Errorf("无法创建本地目录: %v", err)
	}

	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("无法创建本地文件: %v", err)
	}
	defer localFile.Close()

	if _, err := remoteFile.WriteTo(localFile); err != nil {
		return fmt.Errorf("文件下载失败: %v", err)
	}

	fmt.Printf("成功下载文件: %s -> %s\n", remotePath, localPath)
	return nil
}

// DownloadRemoteFile 下载远程文件到本地
func (s *SelfObject) DownloadRemoteFile(remoteFilePath, localFilePath string) error {
	// 创建 SFTP 客户端
	client, err := s.createSFTPClient()
	if err != nil {
		return fmt.Errorf("无法创建 SFTP 客户端: %v", err)
	}
	defer client.Close()

	// 打开远程文件
	remoteFile, err := client.Open(remoteFilePath)
	if err != nil {
		return fmt.Errorf("无法打开远程文件: %v", err)
	}
	defer remoteFile.Close()

	// 创建本地文件
	localFile, err := os.Create(localFilePath)
	if err != nil {
		return fmt.Errorf("无法创建本地文件: %v", err)
	}
	defer localFile.Close()

	// 将远程文件内容复制到本地文件
	if _, err := remoteFile.WriteTo(localFile); err != nil {
		return fmt.Errorf("文件复制失败: %v", err)
	}

	fmt.Printf("文件成功下载到本地: %s\n", localFilePath)
	return nil
}
func (s *SelfObject) TaskCompleted() {
	fmt.Println("任务成功完成，开始清理...")
	s.CleanupRemote() // 调用清理函数
}

func (s *SelfObject) CleanupRemote() {
	// 删除远程 data_temp 和 save_temp 目录
	if err := s.removeRemoteDir(dataTempPath); err != nil {
		fmt.Printf("清理 data_temp 目录失败: %v\n", err)
	}
	if err := s.removeRemoteDir(saveTempPath); err != nil {
		fmt.Printf("清理 save_temp 目录失败: %v\n", err)
	}
}

// 第四部分清理
func (s *SelfObject) CleanupRemote4() {
	// 删除远程 data_temp 和 save_temp 目录
	if err := s.removeRemoteDir(inputTempPath); err != nil {
		fmt.Printf("清理 input_temp 目录失败: %v\n", err)
	}
	if err := s.removeRemoteDir(outputTempPath); err != nil {
		fmt.Printf("清理 output_temp 目录失败: %v\n", err)
	}
}

// removeDir 删除指定的目录及其内容
func (s *SelfObject) removeRemoteDir(remotePath string) error {
	// 检查 SSH 客户端是否已初始化
	if s.CliConf == nil || s.CliConf.Client == nil {
		return fmt.Errorf("SSH 客户端未初始化，无法删除目录: %s", remotePath)
	}

	// 构建删除目录的命令
	removeCommand := fmt.Sprintf("rm -rf %s", remotePath)

	// 执行远程命令
	if err := s.RunCommand(removeCommand); err != nil {
		return fmt.Errorf("无法删除远程目录 %s: %v", remotePath, err)
	}

	fmt.Printf("成功删除远程目录: %s\n", remotePath)
	return nil
}
