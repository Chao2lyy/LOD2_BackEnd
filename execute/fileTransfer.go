// fileTransfer.go
package execute

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
)

// createSFTPClient 创建 SFTP 客户端
func (s *SelfObject) createSFTPClient() (*sftp.Client, error) {
	// 使用现有 SSH 客户端创建 SFTP 客户端
	client, err := sftp.NewClient(s.cliConf.Client)
	if err != nil {
		return nil, fmt.Errorf("无法创建 SFTP 客户端: %v", err)
	}
	return client, nil
}

// UploadLocalFolderToRemote 上传本地文件夹到远程服务器的指定目录
func (s *SelfObject) UploadLocalFolderToRemote(localFolderPath, remoteFolderPath string) error {
	// 创建 SFTP 客户端
	client, err := s.createSFTPClient()
	if err != nil {
		return fmt.Errorf("无法创建 SFTP 客户端: %v", err)
	}
	defer client.Close()

	// 遍历本地文件夹，逐个文件上传
	err = filepath.Walk(localFolderPath, func(localPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 构建远程文件路径
		relPath := strings.TrimPrefix(localPath, localFolderPath)
		relPath = filepath.ToSlash(relPath)
		remotePath := filepath.Join(remoteFolderPath, relPath)
		remotePath = filepath.ToSlash(remotePath)
		remoteFolderPath = filepath.ToSlash(remoteFolderPath)

		if info.IsDir() {
			// 创建远程目录
			if err := client.MkdirAll(remotePath); err != nil {
				return fmt.Errorf("无法创建远程目录: %v", err)
			}
		} else {
			// 打开本地文件
			localFile, err := os.Open(localPath)
			if err != nil {
				return fmt.Errorf("无法打开本地文件: %v", err)
			}
			defer localFile.Close()

			// 创建远程文件
			remoteFile, err := client.Create(remotePath)
			if err != nil {
				return fmt.Errorf("无法创建远程文件: %v", err)
			}
			defer remoteFile.Close()

			// 将本地文件内容复制到远程文件中
			if _, err := localFile.Seek(0, 0); err != nil {
				return err
			}
			if _, err := remoteFile.ReadFrom(localFile); err != nil {
				return fmt.Errorf("文件上传失败: %v", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("文件夹上传失败: %v", err)
	}

	fmt.Printf("本地文件夹成功上传到远程: %s -> %s\n", localFolderPath, remoteFolderPath)
	return nil
}

// DownloadRemoteFolderToLocal 下载远程文件夹到本地指定目录
func (s *SelfObject) DownloadRemoteFolderToLocal(remoteFolderPath, localFolderPath string) error {
	// 创建 SFTP 客户端
	client, err := s.createSFTPClient()
	if err != nil {
		return fmt.Errorf("无法创建 SFTP 客户端: %v", err)
	}
	defer client.Close()

	// 检查远程目录是否存在
	remoteFiles, err := client.ReadDir(remoteFolderPath)
	if err != nil {
		return fmt.Errorf("无法读取远程目录: %v", err)
	}

	// 遍历远程目录中的文件，并下载到本地
	for _, file := range remoteFiles {
		remoteFilePath := filepath.ToSlash(filepath.Join(remoteFolderPath, file.Name()))
		localFilePath := filepath.Join(localFolderPath, file.Name())
		if file.IsDir() {
			// 如果是目录，则递归下载
			if err := s.DownloadRemoteFolderToLocal(remoteFilePath, localFilePath); err != nil {
				return err
			}
		} else {
			// 下载文件
			if err := s.DownloadRemoteFile(remoteFilePath, localFilePath); err != nil {
				return fmt.Errorf("文件下载失败: %v", err)
			}
		}
	}

	fmt.Printf("远程文件夹成功下载到本地: %s -> %s\n", remoteFolderPath, localFolderPath)
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
