package php

import (
	"github.com/jcbowen/jcbaseGo/command"
	"github.com/jcbowen/jcbaseGo/helper"
)

var funcFilePath = "./data/tmp/php/jcbasePHP"

func init() {
	if !helper.FileExists(funcFilePath) {
		err := helper.CreateFile(funcFilePath, []byte(TmpJcbasePHP), 0755, true)
		if err != nil {
			panic(err)
		}
	}
}

func RunFunc(funcName string, args ...string) (string, error) {

	funcName = "--func=" + funcName

	// 往args前面追加"jcbasePHP", funcName
	args = append([]string{funcFilePath, funcName}, args...)

	result, err := command.Run("php", args...)
	if err != nil {
		return "", err
	}

	return result, nil
}
