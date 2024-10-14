// sshConnect.go
package execute

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/ssh"
)

// SelfObject 定义包含 SSH 配置的结构体
type SelfObject struct {
	CliConf *ClientConfig
}

// ClientConfig 定义包含 SSH 会话的配置
type ClientConfig struct {
	Client  *ssh.Client
	Session *ssh.Session
	Stdin   io.WriteCloser
	Stdout  io.Reader
	Stderr  io.Reader
}

// CreateClient 用于创建 SSH 客户端连接
func (c *ClientConfig) CreateClient(ip string, port int, user, password string) error {
	// 创建 SSH 配置
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password), // 使用密码认证
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查（生产环境不推荐）
	}

	// 连接到 SSH 服务器
	addr := fmt.Sprintf("%s:%d", ip, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}
	c.Client = client

	// 创建 SSH 会话
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	c.Session = session

	// 创建标准输入、输出和错误输出的管道
	c.Stdin, err = session.StdinPipe()
	if err != nil {
		return err
	}

	c.Stdout, err = session.StdoutPipe()
	if err != nil {
		return err
	}

	c.Stderr, err = session.StderrPipe()
	if err != nil {
		return err
	}

	return nil
}

// RunCommand 执行远程命令，并将结果异步实时传回
func (s *SelfObject) RunCommand(command string) error {
	// 创建新的 SSH 会话（Session）
	session, err := s.CliConf.Client.NewSession()
	if err != nil {
		return fmt.Errorf("无法创建新的会话: %v", err)
	}
	defer session.Close() // 确保会话在函数退出时关闭

	// 创建标准输入、输出和错误输出的管道
	// stdin, err := session.StdinPipe()
	// if err != nil {
	// 	return fmt.Errorf("无法创建标准输入管道: %v", err)
	// }
	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("无法创建标准输出管道: %v", err)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("无法创建标准错误输出管道: %v", err)
	}

	// 异步处理标准输出和错误输出
	done := make(chan struct{})

	// 处理标准输出
	go func() {
		defer close(done)
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}
			fmt.Print(string(buf[:n])) // 将命令输出实时打印到本地终端
		}
	}()

	// 处理错误输出
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if err != nil {
				break
			}
			fmt.Print("错误:", string(buf[:n])) // 将错误输出实时打印到本地终端
		}
	}()

	// 启动命令
	if err := session.Start(command); err != nil {
		return fmt.Errorf("启动命令失败: %v", err)
	}

	// 等待命令执行完成
	if err := session.Wait(); err != nil {
		return fmt.Errorf("命令执行失败: %v", err)
	}

	// 等待所有输出处理完成
	<-done

	fmt.Println("命令执行完成")
	return nil
}

// createRemoteConfigFile 在远程服务器上创建配置文件
func (s *SelfObject) createRemoteConfigFile(config map[string]string, remoteFilePath string) error {
	// 将配置内容转换为字符串形式
	fmt.Println("remoteFilePath", remoteFilePath)

	configContent := ""
	for key, value := range config {
		configContent += fmt.Sprintf("%s=%s\n", key, value)
	}

	// 构建远程创建文件的命令，使用 echo 将配置内容写入文件
	createCommand := fmt.Sprintf("echo -e \"%s\" > %s", escapeSpecialChars(configContent), remoteFilePath)

	// 在远程服务器上执行创建文件的命令
	return s.RunCommand(createCommand)
}

// removeRemoteConfigFile 在远程服务器上删除配置文件
func (s *SelfObject) removeRemoteConfigFile(remoteFilePath string) error {
	// 构建删除远程文件的命令
	removeCommand := fmt.Sprintf("rm -f %s", remoteFilePath)

	// 执行删除文件的命令
	return s.RunCommand(removeCommand)
}

// escapeSpecialChars 对字符串中的特殊字符进行转义，以便在 echo 中使用
func escapeSpecialChars(content string) string {
	escaped := content
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"") // 转义双引号
	escaped = strings.ReplaceAll(escaped, "$", "\\$")   // 转义美元符号
	return escaped
}
