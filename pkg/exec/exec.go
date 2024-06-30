package exec

import (
	"errors"
	"os/exec"
)

type CommandExecutor interface {
	RunWithEnv(command string, args []string, env []string) ([]byte, error)
}

type NativeCommandExecutor struct {
}

func (fs *NativeCommandExecutor) RunWithEnv(command string, args []string, env []string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = env

	return cmd.CombinedOutput()
}

type TestCommandExecutor struct {
	Fail     bool
	Expected []byte
}

func (fs *TestCommandExecutor) RunWithEnv(command string, args []string, env []string) ([]byte, error) {
	if fs.Fail {
		return []byte{}, errors.New("exec error")
	}

	return fs.Expected, nil
}
