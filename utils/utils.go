package utils

import (
	"errors"
	"os/exec"
	"strings"
)

func ExecCommand(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return errors.New("empty command")
	}
	_, err := exec.Command(parts[0], parts[1:]...).Output()
	return err
}
