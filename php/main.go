package php

import (
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/command"
	"github.com/jcbowen/jcbaseGo/helper"
)

type ConfigStruct struct {
	funcFilePath string
}

func New(opt jcbaseGo.ConfigOption) *ConfigStruct {
	funcFilePath := "/tmp/php/main.go"
	if opt.RuntimePath != "" {
		funcFilePath = opt.RuntimePath + funcFilePath
	} else {
		funcFilePath = helper.DirName(opt.ConfigFile) + funcFilePath
	}

	conf := &ConfigStruct{
		funcFilePath: funcFilePath,
	}

	if !helper.FileExists(conf.funcFilePath) {
		err := helper.CreateFile(conf.funcFilePath, []byte(TmpJcbasePHP), 0755, true)
		if err != nil {
			panic(err)
		}
	}

	return conf
}

func (c *ConfigStruct) RunFunc(funcName string, args ...string) (string, error) {

	funcName = "--func=" + funcName

	// 往args前面追加"jcbasePHP", funcName
	args = append([]string{c.funcFilePath, funcName}, args...)

	result, err := command.Run("php", args...)
	if err != nil {
		return "", err
	}

	return result, nil
}
