package command

import (
	"os/exec"
)

// CmdPath 命令运行路径（绝对/相对路径）
var CmdPath string

// Run 执行cmd命令的封装
func Run(name string, arg ...string) (string, error) {
	if name == "cd" {
		CmdPath = arg[0]
		return "", nil
	}
	cmd := exec.Command(name, arg...)
	if len(CmdPath) > 0 {
		cmd.Dir = CmdPath
	}
	out, err := cmd.CombinedOutput() // 混合输出stdout+stderr
	return string(out), err
}
