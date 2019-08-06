package commands

import (
	"bytes"
	"os/exec"
)

func RunCommand(cmd *exec.Cmd) (string, error) {
	var bOut bytes.Buffer
	var bErr bytes.Buffer

	cmd.Stdout = &bOut
	cmd.Stderr = &bErr

	err := cmd.Run()
	if err == nil {
		return string(bOut.Bytes()), err
	} else {
		return string(bErr.Bytes()), err
	}
}
