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

type Context struct {
	Type string // 类型: hard，default
	Conf jcbaseGo.RepositoryStruct
}

func New(option *Context) *Context {
	if option.Type == "" {
		option.Type = "default"
	}
	return option
}

func (op *Context) Hard() *Context {
	op.Type = "hard"
	return op
}

func (op *Context) Default() *Context {
	op.Type = "default"
	return op
}

func (op *Context) Do(callBack ...any) {
	switch op.Type {
	case "hard":
		hardUpgrade(op.Conf)
	default:
		defaultUpgrade(op.Conf)
	}

	if len(callBack) > 0 {
		// callBack的最后一个参数如果是函数，则执行
		if f, ok := callBack[len(callBack)-1].(func()); ok {
			f()
		}
	}

	os.Exit(0)
}

// 检查git目录是否配置/存在
func checkGitDir(conf jcbaseGo.RepositoryStruct) bool {
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
	exist, _ := helper.DirExists(command.CmdPath+".git/", false, 0)

	if exist {
		fmt.Println("存在.git目录")
	} else {
		fmt.Println("不存在.git目录")
	}

	return exist
}

// hardUpgrade 暴力模式
func hardUpgrade(conf jcbaseGo.RepositoryStruct) {
	if checkGitDir(conf) {
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
		fmt.Printf("\033[31m%s\033[0m\n", "初始化本地仓库失败")
		fmt.Println(result)
		os.Exit(1)
	}
	_, _ = command.Run("touch", "README-TEST.md")
	_, _ = command.Run("git", "add", "README-TEST.md")
	result, _ = command.Run("git", "commit", "-m", "test commit")
	fmt.Printf("提交：\n%s\n", result)
	_, _ = command.Run("git", "remote", "add", conf.RemoteName, conf.RemoteURL)
	result, _ = command.Run("git", "fetch", conf.RemoteName)
	if strings.Contains(result, "fatal:") {
		fmt.Printf("\033[31m%s\033[0m\n", "拉取远程仓库失败")
		fmt.Println(result)
		os.Exit(1)
	}
	fmt.Printf("拉取远程仓库：\n%s\n", result)
	result, _ = command.Run("git", "reset", "--hard", conf.RemoteName+"/"+conf.Branch)
	fmt.Printf("重置到远程仓库最新版本：\n%s\n", result)
	if strings.Contains(result, "fatal:") {
		fmt.Printf("\033[31m%s\033[0m\n", "重置到远程仓库最新版本失败")
		fmt.Println(result)
		os.Exit(1)
	}
	if conf.Branch != "master" {
		// 根据远程分支在本地创建一个同名分支，并切换到该分支
		result, _ = command.Run("git", "checkout", "-t", "remotes/"+conf.RemoteName+"/"+conf.Branch)
		if strings.Contains(result, "fatal:") {
			fmt.Printf("\033[31m%s\033[0m\n", "切换分支失败")
			fmt.Println(result)
			os.Exit(1)
		}
		fmt.Println(result)
		// 删除master分支
		result, _ = command.Run("git", "branch", "-D", "master")
		if strings.Contains(result, "fatal:") {
			fmt.Printf("\033[31m%s\033[0m\n", "删除master分支失败")
			fmt.Println(result)
			os.Exit(1)
		}
	}

	// 输出绿色
	fmt.Printf("\033[32m%s\033[0m\n", "同步成功")
}

// defaultUpgrade 默认模式
func defaultUpgrade(conf jcbaseGo.RepositoryStruct) {
	// git目录不存在就执行hard模式
	if !checkGitDir(conf) {
		fmt.Printf("\033[33m%s\033[0m\n", "不存在.git目录，开始初始化")
		hardUpgrade(conf)
		return
	}

	result, _ := command.Run("git", "fetch", "--all")
	if strings.Contains(result, "fatal:") {
		fmt.Printf("\033[31m%s\033[0m\n", "拉取远程仓库失败")
		fmt.Println(result)
		os.Exit(1)
	}
	fmt.Printf("拉取远程仓库：\n%s\n", result)

	// 获取本地分支
	result, _ = command.Run("git", "branch")
	if strings.Contains(result, "fatal:") {
		fmt.Printf("\033[31m%s\033[0m\n", "获取本地分支失败")
		fmt.Println(result)
		os.Exit(1)
	}
	fmt.Printf("获取本地分支：\n%s\n", result)

	// 根据git branch返回信息，将本地分支存入数组中(过滤掉空元素、*号)
	branchArr := strings.Split(result, "\n")
	branchArr = branchArr[:len(branchArr)-1]
	currentBranch := ""
	for i, v := range branchArr {
		isCurrent := helper.StringStartWith(v, "*")
		branchArr[i] = strings.Replace(v, "*", "", -1)
		branchArr[i] = strings.Replace(branchArr[i], " ", "", -1)
		if isCurrent {
			currentBranch = branchArr[i]
		}
	}
	fmt.Println("当前分支：`", currentBranch, "`")

	// 如果当前分支不是配置的分支，就切换到配置的分支
	if currentBranch != conf.Branch {
		fmt.Println("当前分支不是配置的分支，开始切换到配置的分支")
		if helper.InArray(conf.Branch, branchArr) {
			// 如果远程分支存在本地分支，切换到该分支
			result, _ = command.Run("git", "checkout", conf.Branch)
			if strings.Contains(result, "fatal:") {
				fmt.Printf("\033[31m%s\033[0m\n", "切换本地分支失败")
				fmt.Println(result)
				os.Exit(1)
			}
		} else {
			// 根据远程分支在本地创建一个同名分支，并切换到该分支
			result, _ = command.Run("git", "checkout", "-t", "remotes/"+conf.RemoteName+"/"+conf.Branch)
			if strings.Contains(result, "fatal:") {
				fmt.Printf("\033[31m%s\033[0m\n", "切换远程分支失败")
				fmt.Println(result)
				os.Exit(1)
			}
		}
		fmt.Println(result)
	}

	result, _ = command.Run("git", "reset", "--hard", conf.RemoteName+"/"+conf.Branch)
	fmt.Printf("重置到远程仓库最新版本：\n%s\n", result)
	if strings.Contains(result, "fatal:") {
		fmt.Printf("\033[31m%s\033[0m\n", "重置到远程仓库最新版本失败")
		fmt.Println(result)
		os.Exit(1)
	}
}
