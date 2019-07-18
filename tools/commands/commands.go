package commands

import (
	"bytes"
	"os/exec"
)

func RunCommand(cmd *exec.Cmd) (string, error) {
	var b bytes.Buffer
	cmd.Stdout = &b
	err := cmd.Run()

	if err == nil {
		return string(b.Bytes()), err
	}
	return "", err
}
