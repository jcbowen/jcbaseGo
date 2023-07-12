package helper

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// GetSSHKey 获取 SSH 密钥，如果不存在则生成
func GetSSHKey() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法获取用户主目录：%v", err)
	}
	sshDir := filepath.Join(homeDir, ".ssh")
	pubKeyFile := filepath.Join(sshDir, "id_ed25519.pub")
	if _, err := os.Stat(pubKeyFile); os.IsNotExist(err) {
		// public key file does not exist, generate a new key
		if err := GenerateSSHKey(); err != nil {
			return "", err
		}
	}
	pubKeyBytes, err := ioutil.ReadFile(pubKeyFile)
	if err != nil {
		return "", fmt.Errorf("无法读取公钥文件：%v", err)
	}
	pubKey := string(pubKeyBytes)
	return pubKey, nil
}

// GenerateSSHKey 生成 SSH 密钥
func GenerateSSHKey() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("无法获取用户主目录：%v", err)
	}
	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("无法创建目录：%v", err)
	}
	keyFile := filepath.Join(sshDir, "id_ed25519")
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-C", "jcsite SSH Key", "-f", keyFile, "-N", "", "-q")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("无法生成 SSH 密钥：%v", err)
		}
	}
	return nil
}
