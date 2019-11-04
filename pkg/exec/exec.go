package exec

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

func ExecuteCommand(name string, args ...string) (string, error) {
	fmt.Printf("> Executing command: %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: false,
		Noctty:  false,
	}
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		fmt.Println(buf.String())
		fmt.Printf("> Error: %s\n", err.Error())
		return buf.String(), err
	}
	fmt.Printf("> OK\n")
	return buf.String(), nil
}
