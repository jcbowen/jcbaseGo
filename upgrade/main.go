package upgrade

import (
	"errors"
	"fmt"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/command"
	"github.com/jcbowen/jcbaseGo/helper"
	"os"
	"strings"
)

// Do 执行升级
// 根据配置信息中的仓库配置，执行升级
func Do(conf jcbaseGo.RepositoryStruct, callBack ...any) {
	// 判断是否有配置仓库
	if conf.RemoteURL == "" {
		err := errors.New("repository is empty")
		panic(err)
	}

	command.CmdPath = conf.Dir

	checkExists, err := helper.DirExists(command.CmdPath, true, 0755)
	if err != nil {
		fmt.Printf("checkCmdPath:\033[31m%s\033[0m\n", err)
		os.Exit(1)
	}
	if !checkExists {
		fmt.Printf("checkCmdPath:\033[31m%s\033[0m\n", "命令执行目录不存在，且无法创建")
		os.Exit(1)
	}

	// 检查是否存在.git目录，存在则删除
	if exists, _ := helper.DirExists(command.CmdPath+".git/", false, 0); exists {
		fmt.Printf("\033[33m%s\033[0m\n", "存在.git目录，开始移除")
		err := os.RemoveAll(command.CmdPath + ".git/")
		if err != nil {
			fmt.Printf("rmDir:\033[31m%s\033[0m\n", err)
			os.Exit(1)
		}
		fmt.Printf("\033[32m%s\033[0m\n", "移除成功")
	}
	result, _ := command.Run("git", "init")
	fmt.Printf("初始化本地仓库：\n%s\n", result)
	if strings.Compare(result[:36], "Initialized empty Git repository in") == 0 {
		fmt.Println(result)
		fmt.Printf("\033[31m%s\033[0m\n", "初始化本地仓库失败")
		os.Exit(1)
	}
	_, _ = command.Run("touch", "README-TEST.md")
	_, _ = command.Run("git", "add", "README-TEST.md")
	result, _ = command.Run("git", "commit", "-m", "test commit")
	fmt.Printf("提交：\n%s\n", result)
	_, _ = command.Run("git", "remote", "add", conf.RemoteName, conf.RemoteURL)
	result, _ = command.Run("git", "fetch", conf.RemoteName)
	fmt.Printf("拉取远程仓库：\n%s\n", result)
	if strings.Contains(result, "fatal:") {
		fmt.Printf("\033[31m%s\033[0m\n", "拉取远程仓库失败")
		os.Exit(1)
	}
	result, _ = command.Run("git", "reset", "--hard", conf.RemoteName+"/"+conf.Branch)
	fmt.Printf("重置到远程仓库最新版本：\n%s\n", result)
	if strings.Contains(result, "fatal:") {
		fmt.Printf("\033[31m%s\033[0m\n", "重置到远程仓库最新版本失败")
		os.Exit(1)
	}

	// 输出绿色
	fmt.Printf("\033[32m%s\033[0m\n", "同步成功")

	if len(callBack) > 0 {
		// callBack的最后一个参数如果是函数，则执行
		if f, ok := callBack[len(callBack)-1].(func()); ok {
			f()
		}
	}

	os.Exit(0)
}
