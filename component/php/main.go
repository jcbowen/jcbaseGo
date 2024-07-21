package php

import (
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/command"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"log"
)

type ConfigStruct struct {
	funcFilePath string
}

func New(opt jcbaseGo.Option) *ConfigStruct {
	funcFilePath := "/tmp/php/main.go"
	if opt.RuntimePath != "" {
		funcFilePath = helper.NewFile(&helper.File{Path: opt.RuntimePath}).DirName() + funcFilePath
	} else {
		funcFilePath = helper.NewFile(&helper.File{Path: opt.ConfigFile}).DirName() + funcFilePath
	}

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

	result, err := command.Run("php", args...)
	if err != nil {
		return "", err
	}

	return result, nil
}
