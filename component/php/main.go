package php

import (
	"log"
	"path/filepath"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/command"
	"github.com/jcbowen/jcbaseGo/component/helper"
)

type ConfigStruct struct {
	funcFilePath string
}

func New(opt jcbaseGo.Option) *ConfigStruct {
	baseDir := helper.NewFile(&helper.File{Path: opt.ConfigSource}).DirName()
	if opt.RuntimePath != "" {
		baseDir = helper.NewFile(&helper.File{Path: opt.RuntimePath}).DirName()
	}
	funcFilePath := filepath.Join(baseDir, "tmp", "php", "main.php")

	conf := &ConfigStruct{
		funcFilePath: funcFilePath,
	}

	if !helper.NewFile(&helper.File{Path: conf.funcFilePath}).Exists() {
		err := helper.NewFile(&helper.File{Path: conf.funcFilePath}).CreateFile([]byte(TmpJcbasePHP), true)
		if err != nil {
			log.Panic(err)
		}
	}

	return conf
}

func (c *ConfigStruct) RunFunc(funcName string, args ...string) (string, error) {

	funcName = "--func=" + funcName

	// 往args前面追加"jcbasePHP", funcName
	args = append([]string{c.funcFilePath, funcName}, args...)

	return command.Run("php", args...)
}
